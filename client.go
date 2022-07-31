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

// Client is the remote shell client.
type Client struct {
	server   string
	creds    credentials.TransportCredentials
	ttyState *term.State
}

// NewClientInsecure creates an insecure client.
func NewClientInsecure(server string) *Client {
	return &Client{
		server: server,
		creds:  insecure.NewCredentials(),
	}
}

// ExecOptions are the options for Exec.
type ExecOptions struct {
	Command string
	Args    []string
}

// Exec executes a command in the server.
func (c *Client) Exec(opts *ExecOptions) (*int, error) {
	return c.ExecContext(context.Background(), opts)
}

// ExecContext is like Exec, but with context.
func (c *Client) ExecContext(ctx context.Context, opts *ExecOptions) (*int, error) {

	conn, err := grpc.Dial(c.server, grpc.WithTransportCredentials(c.creds))
	if err != nil {
		return nil, fmt.Errorf("dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewRemoteShellClient(conn)

	stream, err := client.Session(ctx)
	if err != nil {
		return nil, fmt.Errorf("start session: %v", err)
	}

	if opts == nil {
		opts = &ExecOptions{}
	}

	log.Println("DEBUG", "ExecOpts", opts)

	err = stream.Send(&pb.Input{
		Start:   true,
		Command: opts.Command,
		Args:    opts.Args,
	})
	if err != nil {
		return nil, fmt.Errorf("send cmd: %v", err)
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

	go c.readTTY(stream.Context(), inc)

	defer c.restoreTTY()

	go c.writeStream(stream, inc, sigc)

	sigc <- syscall.SIGWINCH

	return c.readStream(stream)
}

func (c *Client) readTTY(ctx context.Context, inc chan<- rune) {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	c.ttyState, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			r, err := tty.ReadRune()
			if err != nil {
				fmt.Println("Error reading from terminal:", err)
			}
			inc <- r
		}
	}()

	<-ctx.Done()
	log.Println("DEBUG", "Exiting readTTY")
	close(inc)
	return
}

func (c *Client) restoreTTY() {
	if c.ttyState == nil {
		return
	}

	err := term.Restore(int(os.Stdin.Fd()), c.ttyState)
	if err != nil {
		log.Println("Error restoring old terminal state:", err)
	}

	log.Println("DEBUG", "Restored old terminal state,", c.ttyState)
}

func (c *Client) readStream(stream pb.RemoteShell_SessionClient) (*int, error) {
	for {
		select {
		case <-stream.Context().Done():
			log.Println("Client stream context done")
			return nil, nil

		default:
			out, err := stream.Recv()
			if err == io.EOF {
				log.Print("Server returned EOF")
				return nil, nil
			}

			if err != nil {
				return nil, err
			}

			if out.ExitCode != 0 || len(out.Bytes) == 0 {
				var exitCode int = int(out.ExitCode)
				return &exitCode, nil
			}

			os.Stdout.Write(out.Bytes)
		}
	}
}

func (c *Client) writeStream(stream pb.RemoteShell_SessionClient, inc <-chan rune, sigc <-chan os.Signal) {

	for {
		select {
		case <-stream.Context().Done():
			return

		case r := <-inc:
			stream.Send(&pb.Input{Bytes: []byte{byte(r)}})

		case sig := <-sigc:
			if sig == nil {
				continue
			}

			s, ok := sig.(syscall.Signal)
			if !ok {
				log.Println("Error forwarding signal: os.Signal is not syscall.Signal, signal:", sig.String())
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
		}
	}
}
