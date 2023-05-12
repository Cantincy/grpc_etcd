package main

import (
	"google.golang.org/grpc"
	"log"
	"net"
	"pro01/grpc_etcd/config"
	"pro01/grpc_etcd/etcd"
	"pro01/grpc_etcd/proto/service"
	"pro01/grpc_etcd/serviceImpl"
)

func main() {

	etcdRegister, err := etcd.NewEtcdRegister(config.EtcdServerAddr)
	if err != nil {
		log.Fatal(err)
	}
	err = etcdRegister.RegisterService(config.ServiceName, config.ServiceAddr, 30)
	if err != nil {
		log.Fatal(err)
	}
	defer etcdRegister.Close()

	listener, err := net.Listen("tcp", config.ServiceAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	server := grpc.NewServer()
	service.RegisterHelloServiceServer(server, &serviceImpl.HelloService{})

	err = server.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}
