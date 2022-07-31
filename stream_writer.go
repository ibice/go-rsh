package rsh

import "github.com/ibice/go-rsh/pb"

type streamWriter struct {
	stream pb.RemoteShell_SessionServer
}

// Write implements the io.Writer interface
func (s streamWriter) Write(p []byte) (int, error) {
	n := len(p)
	if n > 0 {
		s.stream.Send(&pb.Output{Bytes: p})
	}
	return n, nil
}
