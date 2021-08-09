package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	log "github.com/sirupsen/logrus"
	"github.com/umi0410/streamingChat/adapter"
	"github.com/umi0410/streamingChat/client"
	"github.com/umi0410/streamingChat/streamingChat"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var (
	mode                  string
	scanner               = bufio.NewScanner(os.Stdin)
	redisAddr             = flag.String("redisAddr", "localhost:6379", "접속할 Redis의 주소와 포트")
	serverAddr            = flag.String("serverAddr", "localhost:50051", "접속할 gRPC 서버 주소와 포트")
	username              = flag.String("username", "", "클라이언트로 이용 시 채팅방에 접속할 username")
	randomUsername        = flag.Bool("randomUsername", false, "username을 랜덤으로 부여받을 것인지")
	sendRandomChatMessage = flag.Bool("sendRandomChatMessage", false, "개발 과정을 위해 dummy 채팅 메시지를 자동으로 전송하는 클라이언트를 실행할 것인지")
)

func init() {
	flag.Parse()
	mode = flag.Arg(0)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableColors:    false,
		DisableTimestamp: true,
	})
	rand.Seed(time.Now().Unix())
}

func main() {
	ctx := WithGracefullyShutDownContext()

	if mode == "server" {
		messageAdapter := adapter.NewRedisMessageAdapter(ctx, *redisAddr, "chatroom")
		log.Error(streamingChat.NewChatServer(*serverAddr, messageAdapter).Start(ctx))
	} else if mode == "client" {
		if *randomUsername {
			*username += getRandomUsername()
		}
		if *username == "" {
			fmt.Print("Please input your username: ")
			scanner.Scan()
			*username += scanner.Text()
		}

		c := client.NewChatClient(*serverAddr, *sendRandomChatMessage, *username+getRandomAvatar())
		log.Error(c.Start(ctx))
	} else if mode == "fakeClient" {
		for _, fakeName := range []string{"Dummy", "Mike", "Coke", "Pizza", "Pasta"} {
			c := client.NewChatClient(*serverAddr, *sendRandomChatMessage, fakeName+getRandomAvatar())
			log.Error(c.Start(ctx))
		}
		c := client.NewChatClient(*serverAddr, *sendRandomChatMessage, "Chocolate")
		log.Error(c.Start(ctx))
	} else {
		log.Error(flag.Args(), mode)
		panic("Invalid mode, given " + mode)
	}
}

func WithGracefullyShutDownContext() context.Context {
	sigInt := make(chan os.Signal, 10)
	signal.Notify(sigInt, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sigInt
		log.Warning("인터럽트 발생! context를 Cancel합니다.")
		cancel()
	}()
	return ctx
}

func getRandomAvatar() string {
	randomAvatars := []string{
		"🚀", "🐵", "🦍", "🐶", "🐺", "🐱", "🦁", "🐅", "🐷", "🐑", "🍎", "🍐", "🍑", "🍅", "🥝", "🥦", "🏖️", "🌋", "🏛️", "🏞️", "🏦", "🏭", "📊", "🎨", "⏩", "🤖", "🌻", "🌼", "🌲", "💮", "🌊", "🎄", "🔥",
	}
	return randomAvatars[rand.Intn(len(randomAvatars))]
}

func getRandomUsername() string {
	return gofakeit.LastName() + strconv.Itoa(rand.Intn(100))
}
