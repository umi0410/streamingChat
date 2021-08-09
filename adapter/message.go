package adapter

import (
    "context"
    "encoding/json"
    "github.com/go-redis/redis/v8"
    log "github.com/sirupsen/logrus"
    "github.com/umi0410/streamingChat/data"
    "github.com/umi0410/streamingChat/pb"
)

type MessageAdapter interface {
    PublishMessage(ctx context.Context, message *pb.ChatStream_Message)
    GetNewMessage(ctx context.Context) <-chan *ReceiveResult
}

type RedisMessageAdapter struct{
    client *redis.Client
    pubsub *redis.PubSub
    channel string
}

type ReceiveResult struct{
    MessageDTO *data.MessageDTO
    Error error
}

func NewRedisMessageAdapter(ctx context.Context, redisAddr string, channel string) *RedisMessageAdapter {
    rdb := redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    pubsub := rdb.Subscribe(ctx, channel)

    return &RedisMessageAdapter{
        client: rdb,
        pubsub: pubsub,
        channel: channel,
    }
}

func (ma *RedisMessageAdapter) PublishMessage(ctx context.Context, message *pb.ChatStream_Message) {
    intCmd := ma.client.Publish(ctx, ma.channel, MessageWrapper(*message))
    log.Info("Redis에 메시지를 Publish했습니다. ", intCmd)
}

func (ma *RedisMessageAdapter) GetNewMessage(ctx context.Context) <-chan *ReceiveResult{
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

type MessageWrapper pb.ChatStream_Message

func (m MessageWrapper) MarshalBinary() ([]byte, error) {
    return json.Marshal(m)
}



