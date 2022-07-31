package rsh

import (
	"context"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ibice/go-rsh/pb"
)

type Client struct {
	server string
	creds  credentials.TransportCredentials
}

func NewClientInsecure(server string) *Client {
	return &Client{
		server: server,
		creds:  insecure.NewCredentials(),
	}
}

func (c *Client) Exec() error {
	return c.ExecContext(context.Background())
}

func (c *Client) ExecContext(ctx context.Context) error {

	conn, err := grpc.Dial(c.server, grpc.WithTransportCredentials(c.creds))
	if err != nil {
		return fmt.Errorf("dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRemoteShellClient(conn)

	stream, err := client.Session(ctx)
	if err != nil {
		return fmt.Errorf("start session: %v", err)
	}

	var (
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

	go c.readTTY(ctx, inc)

	go c.readStream(ctx, stream)

	sigc <- syscall.SIGWINCH

	c.writeStream(ctx, inc, sigc, stream)

	return nil
}

func (c *Client) readTTY(ctx context.Context, inc chan<- rune) {
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
			inc <- r
		}
	}()

	for {
		<-ctx.Done()
		close(inc)
		return
	}
}

func (c *Client) readStream(ctx context.Context, stream pb.RemoteShell_SessionClient) {
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

func (c *Client) writeStream(ctx context.Context, inc <-chan rune, sigc <-chan os.Signal,
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
