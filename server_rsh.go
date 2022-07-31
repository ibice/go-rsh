package rsh

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"syscall"

	"github.com/creack/pty"

	"github.com/ibice/go-rsh/pb"
)

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

type rshServer struct {
	pb.UnimplementedRemoteShellServer
	shell string
}

func newRSHServer(shell string) *rshServer {
	return &rshServer{shell: shell}
}

func (s *rshServer) Session(stream pb.RemoteShell_SessionServer) error {
	log.Println("Opening remote shell session")

	var (
		ctx, cancel = context.WithCancel(stream.Context())
		cmd         = exec.CommandContext(ctx, s.shell)
		inc         = make(chan *pb.Input)
	)
	defer cancel()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return fmt.Errorf("starting pty: %v", err)
	}
	defer ptmx.Close()

	go s.readStream(ctx, cancel, stream, inc)

	go func() {
		defer cancel()
		io.Copy(streamWriter{stream}, ptmx)
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("Closing remote shell session")
			return nil

		case in := <-inc:
			if in.Signal != 0 {
				s := syscall.Signal(in.Signal)
				switch s {
				case syscall.SIGWINCH:
					sizeParts := strings.Split(string(in.Bytes), " ")
					size := &pty.Winsize{
						Cols: parseUint16(sizeParts[0]),
						Rows: parseUint16(sizeParts[1]),
						X:    parseUint16(sizeParts[2]),
						Y:    parseUint16(sizeParts[3]),
					}
					log.Println("Setting window size to", size)
					if err := pty.Setsize(ptmx, size); err != nil {
						log.Println("Error setting window size:", err)
					}
				default:
					cmd.Process.Signal(s)
				}
			} else {
				_, err := ptmx.Write(in.Bytes)
				if err != nil {
					log.Printf("Error writing PTY: %v", err)
					cancel()
				}
			}
		}
	}
}

func (s *rshServer) readStream(ctx context.Context, cancel context.CancelFunc, stream pb.RemoteShell_SessionServer, c chan<- *pb.Input) {
	for {
		select {
		case <-ctx.Done():
			return

		case <-stream.Context().Done():
			return

		default:
			in, err := stream.Recv()
			if err == io.EOF {
				log.Println("Exiting ptmx write routine")
				cancel()
				return
			}
			if err != nil {
				log.Println("Error reading from stream:", err)
				cancel()
				return
			}
			c <- in
		}
	}
}
