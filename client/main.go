package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"os/signal"
	"pro01/grpc_etcd/config"
	"pro01/grpc_etcd/etcd"
	"pro01/grpc_etcd/proto/service"
	"syscall"
	"time"
)

func main() {
	c := make(chan os.Signal, 1)
	go func() {
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	}()

	etcdDiscovery, err := etcd.NewServiceDiscovery([]string{config.EtcdServerAddr})
	if err != nil {
		log.Fatal(err)
	}
	etcdDiscovery.WatchService("/hello")

	defer etcdDiscovery.Close()

	for {
		time.Sleep(time.Second * 2)
		serviceAddr, err := etcdDiscovery.GetService(config.ServiceName)

		conn, err := grpc.Dial(serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}

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
		conn.Close()
	}
}
