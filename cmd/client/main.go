package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ibice/go-rsh"
)

var (
	port           = flag.Uint("p", 22222, "server port")
	addr           = flag.String("a", "127.0.0.1", "server address")
	remoteExitCode = flag.Bool("e", false, "use exit code of remote process")

	command string
	args    []string
)

func parseArgs() {
	flag.Parse()

	if port == nil || *port == 0 {
		log.Fatal("-p is required")
	}

	if *port > 65535 {
		log.Fatal("Invalid port: ")
	}

	if addr == nil || *addr == "" {
		log.Fatal("-a is required")
	}

	// Parse remote command arguments
	var argsAfterDash []string

	for i, arg := range os.Args {
		if arg == "--" {
			argsAfterDash = os.Args[i+1:]
			break
		}
	}

	if len(argsAfterDash) > 0 {
		command = argsAfterDash[0]
		if len(argsAfterDash) > 1 {
			args = argsAfterDash[1:]
		}
	}
}

func main() {
	parseArgs()

	client := rsh.NewClientInsecure(fmt.Sprintf("%s:%d", *addr, *port))

	opts := &rsh.ExecOptions{
		Command: command,
		Args:    args,
	}

	exitCode, err := client.Exec(opts)

	if err != nil {
		log.Fatalf("Exec: %v", err)
	}

	if exitCode != nil && *remoteExitCode {
		os.Exit(*exitCode)
	}
}
