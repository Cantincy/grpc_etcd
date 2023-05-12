package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"pro01/grpc_etcd/proto/service"
	"pro01/grpc_etcd/serviceImpl"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:9090")
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	service.RegisterHelloServiceServer(server, &serviceImpl.HelloService{})
	err = server.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
