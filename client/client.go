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

type Client struct{

}
func RunClient(username string){
    scanner := bufio.NewScanner(os.Stdin)
    for {
        logger := log.WithField("Username", username)
        logger.Info("Dial to 0.0.0.0:50051")
        conn, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
        if err != nil {
            logger.Error(err)
        }
        client := pb.NewChatClient(conn)
        chatStream, err := client.Stream(context.Background())
        if err != nil {
            logger.Error("서버에 연결하지 못했습니다. ", err)
            time.Sleep(time.Second)
            continue
        }

        err = chatStream.Send(&pb.ChatStream{
            Event: &pb.ChatStream_Login_{
                Login: &pb.ChatStream_Login{
                    Username: username,
                },
            },
        })
        if err != nil {
            logger.Error(err)
        } else{
            ctx, cancel := context.WithCancel(context.Background())
            go func(ctx context.Context) {
                for {
                    select{
                    case <-ctx.Done():
                        return
                    default:
                        scanner.Scan()
                        err := chatStream.Send(&pb.ChatStream{
                            Event: &pb.ChatStream_Message_{
                                Message: &pb.ChatStream_Message{Author: username, Content: scanner.Text()},
                            },
                        })
                        if err != nil {
                            logger.Error(err)
                        }
                    }
                }
            }(ctx)

            for{
                newMessage, err := chatStream.Recv()
                if err != nil{
                    logger.Error(err)
                    break
                }
                fmt.Printf("%10s: %s\n", newMessage.GetMessage().Author, newMessage.GetMessage().Content)
            }
            cancel()
        }
    }
}
