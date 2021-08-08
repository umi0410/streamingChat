package client

import (
    "bufio"
    "context"
    "fmt"
    log "github.com/sirupsen/logrus"
    "github.com/umi0410/streamingChat/pb"
    "google.golang.org/grpc"
    "os"
    "time"
)

type ChatClient struct{
    streamClient pb.Chat_StreamClient
    chatClient pb.ChatClient
    Username string
    Ctx context.Context
    scanner *bufio.Scanner
    logger log.FieldLogger
}

func NewChatClient(username string) *ChatClient{
    scanner := bufio.NewScanner(os.Stdin)
    return &ChatClient{
        scanner: scanner,
        Username: username,
        logger: log.WithField("Username", username),
    }
}

func (c *ChatClient) Start(ctx context.Context) error{
    c.Ctx = ctx
    if err := c.Connect(); err != nil {
        return err
    }

    if err := c.Login(); err != nil {
        return err
    }

    receiveDone := c.Receive()
    sendDone := c.Send()
    <- receiveDone
    <- sendDone

    return nil
}

func (c *ChatClient) Connect() error {
    for i := 0;;{
        select {
        case <-c.Ctx.Done():
            return fmt.Errorf("서버에 연결하지 못했습니다")

        default:
            c.logger.Info("Dial to 0.0.0.0:50051")
            conn, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
            if err != nil {
                c.logger.Error(fmt.Errorf("%d 번째 시도: 서버에 연결이 실패했습니다: %w", i, err))
            } else {
                c.logger.Info("서버에 연결되었습니다.")
                c.chatClient = pb.NewChatClient(conn)
                streamClient, err := c.chatClient.Stream(c.Ctx)
                if err != nil {
                    c.logger.Error(fmt.Errorf("%d 번째 시도: 채팅 스트림 생성에 실패했습니다: %w", i, err))
                } else{
                    c.logger.Info("채팅을 주고 받을 스트림을 생성했습니다.")
                    c.streamClient = streamClient
                    return nil
                }
            }
        }
        time.Sleep(time.Second)
    }
}

func (c *ChatClient) Login() error{
    err := c.streamClient.Send(&pb.ChatStream{
        Event: &pb.ChatStream_Login_{
            Login: &pb.ChatStream_Login{
                Username: c.Username,
            },
        },
    })
    if err != nil {
        return fmt.Errorf("failed to request for login: %w", err)
    }

    if err != nil {
        return fmt.Errorf("failed to login: %w", err)
    }

    return nil
}

func (c *ChatClient) Receive() <-chan struct{}{
    done := make(chan struct{})
    go func() {
        defer c.logger.Debug("Context에 의해..? Receive 종료")
        for {
            select {
            case <-c.Ctx.Done():
                done <- struct{}{}
                return
            default:
                for{
                    newMessage, err := c.streamClient.Recv()
                    if err != nil{
                        c.logger.Error(err)
                    } else{
                        fmt.Printf("🧐 %8s: %s\n", newMessage.GetMessage().Author, newMessage.GetMessage().Content)
                    }
                }
            }
        }
    }()

    return done
}

func (c *ChatClient) Send() <-chan struct{}{
    done := make(chan struct{})
    go func() {
        defer c.logger.Debug("Context에 의해..? Send 종료")
        for {
            select{
            case <-c.Ctx.Done():
                done <- struct{}{}
                return
            default:
                c.scanner.Scan()
                err := c.streamClient.Send(&pb.ChatStream{
                    Event: &pb.ChatStream_Message_{
                        Message: &pb.ChatStream_Message{Author: c.Username, Content: c.scanner.Text()},
                    },
                })
                if err != nil {
                    c.logger.Error(fmt.Errorf("메시지 전송을 실패했습니다: %w", err))
                }
            }
        }
    }()

    return done
}
