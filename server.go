package rsh

import (
	"fmt"
	"net"

	"github.com/ibice/go-rsh/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server is the remote shell server.
type Server struct {
	address string
	shell   string
}

// NewServer creates a new remote shell server.
func NewServer(address string, shell string) *Server {
	return &Server{
		address: address,
		shell:   shell,
	}
}

// Serve starts the server.
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
