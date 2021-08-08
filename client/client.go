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

// Start 는 ChatClient 가 수행해야할 모든 동작들을 시작시킵니다.
// 각 동작들은 Context가 Cancel되어 Done 상태가 되면 그것을 감지하고
// 종료됩니다.
func (c *ChatClient) Start(ctx context.Context) error{
    c.Ctx = ctx
    if err := c.Connect(); err != nil {
        return err
    }
    if err := c.Login(); err != nil {
        return err
    }
    fmt.Printf("🚀🚀🚀 LOGIN SUCCESSED. Welcome %s !!!", c.Username)
    receiveDone := c.Receive()
    sendDone := c.Send()
    <- receiveDone
    <- sendDone

    return nil
}

// Connect 는 Dialing을 비롯한 초기 연결을 담당합니다.
// 서버와 연결되면 채팅을 주고 받을 수 있는 스트림을 생성하고 저장합니다.
// 연결에 실패할 경우 특정 시간만큼 대기 후 반복해서 재시도합니다.
// Context가 Cancel되어 Done 상태인 경우 종료합니다.
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

// Login 은 생성된 Stream으로 첫 번째 요청을 보냅니다.
// 첫 번째 요청은 채팅 메시지 전송이 아닌 로그인이어야합니다.
// 따라서 *ChatClient.Send 보다 먼저 호출되어야합니다.
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


// Receive 는 Concurrently하게 ChatStream으로부터 메시지를 전달받습니다.
// Context 가 Cancel되어 Done 상태인 경우 종료합니다.
func (c *ChatClient) Receive() <-chan struct{}{
    done := make(chan struct{})
    go func() {
        for {
            select {
            case <-c.Ctx.Done():
                c.logger.Debug("Context에 의해..? Receive 종료")
                done <- struct{}{}
                return
            default:
                newMessage, err := c.streamClient.Recv()
                if err != nil{
                    if c.Ctx.Err() == context.Canceled{
                        // pass. err is context canceled error
                    } else{
                        c.logger.Error(fmt.Errorf("메시지를 전달받지 못했습니다: %w",  err))
                        time.Sleep(time.Second)
                    }
                } else{
                    fmt.Printf("|%-8s| %s\n", newMessage.GetMessage().Author, newMessage.GetMessage().Content)
                }

            }
        }
    }()

    return done
}

// Send 는 Concurrently하게 ChatStream에게 메시지를 전송합니다.
// Context 가 Cancel되어 Done 상태인 경우 종료합니다.
func (c *ChatClient) Send() <-chan struct{}{
    done := make(chan struct{})
    go func() {
        for {
            select{
            case <-c.Ctx.Done():
                c.logger.Debug("Context에 의해..? Send 종료")
                done <- struct{}{}
                return
            case input := <- c.Scan():
                err := c.streamClient.Send(&pb.ChatStream{
                    Event: &pb.ChatStream_Message_{
                        Message: &pb.ChatStream_Message{Author: c.Username, Content: input},
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

// Scan 하는 동안 Block 되지 않고 Context.Done()과 함께 select 문에 놓일 수 있게 함.
func (c *ChatClient) Scan() <-chan string{
    input := make(chan string)
    go func (){
        c.scanner.Scan()
        input <- c.scanner.Text()
    }()

    return input
}
