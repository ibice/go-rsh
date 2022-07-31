// Package server will execute all commands issued by connected clients.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/ibice/go-rsh"
)

var (
	port  = flag.Uint("port", 22222, "listen port")
	addr  = flag.String("address", "", "listen address")
	shell = flag.String("shell", os.Getenv("SHELL"), "shell to use")
)

func main() {
	address := *addr + fmt.Sprintf(":%d", *port)
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterRemoteShellServer(s, newServer())
	reflection.Register(s)

	log.Println("Serving at", address)
	panic(s.Serve(l))
}

func init() {
	flag.Parse()
}
