package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
	"github.com/umi0410/streamingChat/data"
	"github.com/umi0410/streamingChat/pb"
	"strings"
	"time"
)

type MessageAdapter interface {
	PublishMessage(ctx context.Context, message *pb.ChatStream_Message)
	GetNewMessage(ctx context.Context) <-chan *ReceiveResult
	GetWordFrequency(ctx context.Context) []*data.HotWordDTO
	AddNewWordEvent(ctx context.Context, message, username string, ttl time.Duration)
}

type RedisMessageAdapter struct {
	client  *redis.Client
	pubsub  *redis.PubSub
	channel string
}

type ReceiveResult struct {
	MessageDTO *data.MessageDTO
	Error      error
}

func NewRedisMessageAdapter(ctx context.Context, redisAddr string, channel string) *RedisMessageAdapter {
	log.Info("Redis connecting to ", redisAddr)
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pubsub := rdb.Subscribe(ctx, channel)

	return &RedisMessageAdapter{
		client:  rdb,
		pubsub:  pubsub,
		channel: channel,
	}
}

func (ma *RedisMessageAdapter) PublishMessage(ctx context.Context, message *pb.ChatStream_Message) {
	intCmd := ma.client.Publish(ctx, ma.channel, MessageWrapper(*message))
	log.Info("Redis에 메시지를 Publish했습니다. ", intCmd)
}

func (ma *RedisMessageAdapter) GetNewMessage(ctx context.Context) <-chan *ReceiveResult {
	result := make(chan *ReceiveResult)
	go func() {
		msg, err := ma.pubsub.ReceiveMessage(ctx)
		log.Info("Redis에서 메시지를 받았습니다.")
		if err != nil {
			result <- &ReceiveResult{Error: err}
			return
		}
		messageDTO := new(data.MessageDTO)

		if err := json.Unmarshal([]byte(msg.Payload), messageDTO); err != nil {
			result <- &ReceiveResult{Error: err}
			return
		}

		result <- &ReceiveResult{MessageDTO: messageDTO}
		return
	}()

	return result
}

func (a *RedisMessageAdapter) GetWordFrequency(ctx context.Context) []*data.HotWordDTO {
	scanCmd := a.client.Scan(ctx, 0, "hotWord@*", 30000)
	keys, nextCursor, err := scanCmd.Result()
	if err != nil {
		log.Error(err)
	}
	if nextCursor != 0 {
		log.Warning("Scan시에 다음 커서가 존재합니다. count의 크기를 늘리는 방안을 고려해보세요.")
	}
	ranking := make([]*data.HotWordDTO, 0)
	wordSet := make(map[string]struct{})
	for _, key := range keys {
		splitted := strings.Split(key, "@")
		if len(splitted) != 3 {
			log.Warning("조회한 키 중 hotWord@{{word}}@{{username}} 의 형태가 아닌 키가 존재합니다: ", key)
		} else {
			if _, isExists := wordSet[splitted[1]]; !isExists {
				scanCmd := a.client.Scan(ctx, 0, fmt.Sprintf("hotWord@%s@*", splitted[1]), 30000)
				keys, nextCursor, err := scanCmd.Result()
				if err != nil {
					log.Error(err)
				}
				if nextCursor != 0 {
					log.Warning("Scan시에 다음 커서가 존재합니다. count의 크기를 늘리는 방안을 고려해보세요.")
				}
				ranking = append(ranking, &data.HotWordDTO{Word: splitted[1], Frequency: len(keys)})
				wordSet[splitted[1]] = struct{}{}
				//log.Debugf("단어 %s의 빈도: %d", splitted[1], ranking[len(ranking)-1].Frequency)
			}
		}
	}

	return ranking
}

func (a *RedisMessageAdapter) AddNewWordEvent(ctx context.Context, message, username string, ttl time.Duration) {
	words := strings.Split(message, " ")
	for _, word := range words {
		go func() {
			statusCmd := a.client.Set(ctx, fmt.Sprintf("hotWord@%s@%s", word, username), 0, ttl)
			if statusCmd.Err() != nil {
				log.Error(statusCmd.Err())
			}
		}()
	}
}

type MessageWrapper pb.ChatStream_Message

func (m MessageWrapper) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}
