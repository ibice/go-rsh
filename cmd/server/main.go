package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ibice/go-rsh"
)

var (
	port  = flag.Uint("port", 22222, "listen port")
	addr  = flag.String("addr", "127.0.0.1", "listen address")
	shell = flag.String("shell", os.Getenv("SHELL"), "default shell to use")

	lastResortShell = "/bin/sh"
)

func parseArgs() {
	flag.Parse()

	if port == nil || *port == 0 {
		log.Fatal("-port is required")
	}

	if *port > 65535 {
		log.Fatal("Invalid port: ")
	}

	if addr == nil || *addr == "" {
		log.Fatal("-addr is required")
	}

	if shell == nil || *shell == "" {
		shell = &lastResortShell
	}
}

func main() {
	parseArgs()

	server := rsh.NewServer(fmt.Sprintf("%s:%d", *addr, *port), *shell)

	log.Printf("Serving at %s:%d", *addr, *port)

	if err := server.Serve(); err != nil {
		log.Fatalf("Serve: %v", err)
	}
}
