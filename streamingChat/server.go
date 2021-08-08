package streamingChat

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/umi0410/streamingChat/pb"
	"google.golang.org/grpc"
	"io"
	"net"
	"sync"
)

type ChatServer  struct{
	pb.UnimplementedChatServer
	connectionLock *sync.Mutex
	Ctx context.Context
    Connections map[string]*connection
    NewMessages chan *pb.ChatStream
}

// connection 은 Chat 서비스의 StreamServer와 유사하게 동작합니다.
// Embedding을 통해 편리하게 사용할 수 있습니다.
type connection struct{
	pb.Chat_StreamServer
	username string
	logger log.FieldLogger
	gracefulTTLLeft int
}

func NewChatServer() *ChatServer{
    return &ChatServer{
    	Connections: make(map[string]*connection),
    	NewMessages: make(chan *pb.ChatStream),
    	connectionLock: new(sync.Mutex),
	}
}

// Start 는 전체 ChatServer의 작업을 시작합니다.
func (s *ChatServer) Start(ctx context.Context) error {
	s.Ctx = ctx
	srv := grpc.NewServer()
	pb.RegisterChatServer(srv, s)
	log.Info("Listening 0.0.0.0:50051")
	listener, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
	    return fmt.Errorf("failed to listen the port: %w", err)
	}

	broadcastDone := s.broadcast()

	if err := srv.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	srv.GracefulStop()
	<- broadcastDone

	return nil
}

func (s *ChatServer) broadcast() <-chan struct{}{
	done := make(chan struct{})
	go func (){
		for {
			select {
			case <-s.Ctx.Done():
				done <- struct{}{}
				return
			case message := <-s.NewMessages:
				for username, conn := range s.Connections {
					// author가 아닌 경우에만 메시지 전송
					if conn.username != message.GetMessage().Author {
						if err := conn.Send(message); err != nil {
							if conn.Context().Err() != context.Canceled {
								log.WithField("Username", username).Error("failed to send message: ", err)
							}
							s.connectionLock.Lock()
							// 여기서 삭제하면 for 문은 어떻게 되는거지
							// 참고: https://stackoverflow.com/questions/23229975/is-it-safe-to-remove-selected-keys-from-map-within-a-range-loop
							delete(s.Connections, username)
							s.connectionLock.Unlock()
						}
					}

				}
			}
		}
	}()

	return done
}

func (s *ChatServer) Stream(srv pb.Chat_StreamServer) error {
	stream, err := srv.Recv()
	if err != nil {
		log.Error(err)
		return err
	}

	login := stream.GetLogin()
	if login == nil {
		log.Error("로그인 해주세요.")
		return nil
	}
	log.Info("새 유저 연결: ", login.Username)
	conn := &connection{
		Chat_StreamServer: srv,
		username: login.Username,
		logger: log.WithField("Username", login.Username),
		gracefulTTLLeft: 5,
	}
	s.connectionLock.Lock()
	s.Connections[login.Username] = conn
	s.connectionLock.Unlock()

	for {
		select{
		case <- s.Ctx.Done():
			if 0 < conn.gracefulTTLLeft {
				conn.logger.Warn("Context는 종료되었지만 gracefulTTL이 남아있어 잠시 기다립니다.")
			} else {
				conn.logger.Warn("gracefulTTL이 0보다 작아져 요청을 끊습니다.")
				s.connectionLock.Lock()
				delete(s.Connections, conn.username)
				s.connectionLock.Unlock()

				return nil
			}
			conn.gracefulTTLLeft -= 1

		case stream := <- s.Recv(conn):
			if message := stream.GetMessage(); message != nil {
				conn.logger.Infof("Server) Client sent a message. %s", message)
				s.NewMessages <- stream
			} else if logout := stream.GetLogout(); logout != nil {
				conn.logger.Info("Logout")
			} else{
				conn.logger.Error("요청 끊어진 듯")
				return nil
			}
			if s.Ctx.Err() == context.Canceled {
				conn.logger.Info("안전하게 종료합니다.")
				return nil
			}
		}
	}
}

func (s *ChatServer) Recv(conn *connection) <-chan *pb.ChatStream{
	done := make(chan *pb.ChatStream)
	go func (){
		stream, err := conn.Recv()
		if err != nil{
			if err == io.EOF {
				conn.logger.Warning("EOF 에러 발생. 연결 끊어진 듯: ", err)
			} else if conn.Context().Err() == context.Canceled{
				conn.logger.Warning("종료된 Context입니다: ", err)
			} else{
				conn.logger.Error("알 수 없는 에러네요: ", err)
			}
		}
		done <- stream
	}()

	return done
}