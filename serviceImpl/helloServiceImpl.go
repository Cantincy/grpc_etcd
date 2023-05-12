package serviceImpl

import (
	"context"
	"log"
	"pro01/grpc_etcd/proto/service"
	"time"
)

type HelloService struct {
}

var _ service.HelloServiceServer = (*HelloService)(nil)

func (h *HelloService) SayHello(ctx context.Context, req *service.HelloRequest) (*service.HelloResponse, error) {
	log.Printf("[Server]:接收到消息...消息ID为%d,消息类型为%s,发送时间为%s", req.Hello.HelloID, req.Hello.Msg, req.Hello.SendTime)
	return &service.HelloResponse{
		Resp: &service.Hello{
			HelloID:  req.Hello.HelloID + 1,
			Msg:      "Hello from Server",
			SendTime: time.Now().Format("2006-01-02 15:04:05"),
		},
		Code: 0,
	}, nil
}
