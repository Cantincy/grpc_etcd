package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"pro01/grpc_etcd/proto/service"
	"time"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	cli := service.NewHelloServiceClient(conn)
	req := &service.HelloRequest{
		Hello: &service.Hello{
			HelloID:  1,
			Msg:      "Hello from Client",
			SendTime: time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	resp, err := cli.SayHello(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", resp)
}
