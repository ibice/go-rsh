package rsh

import (
	"fmt"
	"net"

	"github.com/ibice/go-rsh/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	address string
	shell   string
}

func NewServer(address string, shell string) *Server {
	return &Server{
		address: address,
		shell:   shell,
	}
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("listen: %v", err)
	}

	g := grpc.NewServer()

	pb.RegisterRemoteShellServer(g, newRSHServer(s.shell))

	reflection.Register(g)

	return g.Serve(l)
}
