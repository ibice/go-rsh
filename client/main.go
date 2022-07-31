// Package client which will connect to a server and run a Go command.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/mattn/go-tty"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/ibice/go-rsh"
)

var (
	port = flag.Uint("port", 22222, "listen port")
	addr = flag.String("address", "", "listen address")
)

func address() string {
	return *addr + fmt.Sprintf(":%d", *port)
}

func readTTY(ctx context.Context, c chan<- rune) {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	go func() {
		for {
			r, err := tty.ReadRune()
			if err != nil {
				fmt.Println("Error reading from terminal:", err)
			}
			c <- r
		}
	}()

	for {
		<-ctx.Done()
		close(c)
		return
	}
}

func readStream(ctx context.Context, stream pb.RemoteShell_SessionClient) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			out, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				log.Println(err)
			}
			os.Stdout.Write(out.Bytes)
		}
	}
}

func writeStream(ctx context.Context, inc <-chan rune, sigc <-chan os.Signal,
	stream pb.RemoteShell_SessionClient) {

	for {
		select {
		case r := <-inc:
			stream.Send(&pb.Input{Bytes: []byte{byte(r)}})

		case sig := <-sigc:
			s, ok := sig.(syscall.Signal)
			if !ok {
				log.Println("Error forwarding signal: are you in windows?")
				break
			}

			switch s {
			case syscall.SIGWINCH:
				size, err := pty.GetsizeFull(os.Stdin)
				if err != nil {
					log.Printf("Error getting terminal size: %v", err)
					break
				}
				stream.Send(&pb.Input{Signal: int32(s), Bytes: []byte(fmt.Sprintf(
					"%d %d %d %d",
					size.Cols,
					size.Rows,
					size.X,
					size.Y,
				))})

			default:
				stream.Send(&pb.Input{Signal: int32(s)})
			}

		case <-ctx.Done():
			stream.CloseSend()
			return
		}
	}
}

func main() {
	flag.Parse()

	conn, err := grpc.Dial(address(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}

	defer conn.Close()

	c := pb.NewRemoteShellClient(conn)

	stream, err := c.Session(context.Background())
	if err != nil {
		panic(err)
	}

	var (
		ctx  = stream.Context()
		inc  = make(chan rune, 1024)
		sigc = make(chan os.Signal, 1)
	)

	signal.Notify(sigc,
		syscall.SIGWINCH,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go readTTY(ctx, inc)

	go readStream(ctx, stream)

	sigc <- syscall.SIGWINCH

	writeStream(ctx, inc, sigc, stream)
}
