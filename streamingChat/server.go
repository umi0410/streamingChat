package streamingChat

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/umi0410/streamingChat/pb"
	"google.golang.org/grpc"
	"io"
	"net"
)

type Server  struct{
	pb.UnimplementedChatServer
    Connections []*connection
    NewMessages chan *pb.ChatStream
}

type connection struct{
	pb.Chat_StreamServer
	username string
}

func NewServer() *Server{
    return &Server{
    	NewMessages: make(chan *pb.ChatStream),
	}
}

func (s *Server) Run(ctx context.Context) error {
    ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	srv := grpc.NewServer()
	pb.RegisterChatServer(srv, s)
	log.Info("Listening 0.0.0.0:50051")
	l, err := net.Listen("tcp", "0.0.0.0:50051")

	if err != nil {
	    log.Error(err)
		return err
	}

	go s.broadcast()

	go func() {
		_ = srv.Serve(l)
		cancel()
	}()

	<-ctx.Done()

	//s.Broadcast <- &chat.StreamResponse{
	//	Timestamp: ptypes.TimestampNow(),
	//	Event: &chat.StreamResponse_ServerShutdown{
	//		ServerShutdown: &chat.StreamResponse_Shutdown{}}}
	//
	//close(s.Broadcast)

	srv.GracefulStop()

	return nil
}

func (s *Server) broadcast() {
	for message := range s.NewMessages{
		//log.Info(message)
		for i, conn := range s.Connections {
			// author가 아닌 경우에만 메시지 전송
			if conn.username != message.GetMessage().Author {
				if err := conn.Send(message); err != nil {
					log.Error(err, i)
					//s.Connections = append(s.Connections[:i], s.Connections[i+1:]...)
				}
			}

		}
	}
}

func (s *Server) Stream(srv pb.Chat_StreamServer) error {
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
	s.Connections = append(s.Connections, &connection{Chat_StreamServer: srv, username: login.Username})

	logger := log.WithField("username", login.Username)
	for {
		req, err := srv.Recv()
		if errors.Is(err, io.EOF) {
			logger.Error("EOF 에러 발생. 연결 끊어진 듯")
		}
		if message := req.GetMessage(); message != nil {
			logger.Infof("Server) Client sent a message. %s", message)
			s.NewMessages <- req
		} else if logout := req.GetLogout(); logout != nil {
			logger.Info("Logout")
			return nil
		} else{
			logger.Error("읽을 바디가 없네")
			return nil
		}
	}

	return nil
}