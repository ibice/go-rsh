package main

import pb "github.com/carpioldc/go-rsh-grpc"

type streamWriter struct {
	stream pb.RemoteShell_SessionServer
}

func (s streamWriter) Write(p []byte) (int, error) {
	n := len(p)
	if n > 0 {
		s.stream.Send(&pb.Output{Bytes: p})
	}
	return n, nil
}
