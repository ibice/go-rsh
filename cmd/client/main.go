package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ibice/go-rsh"
)

var (
	port = flag.Uint("p", 22222, "server port")
	addr = flag.String("a", "127.0.0.1", "server address")
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
}

func main() {
	parseArgs()

	client := rsh.NewClientInsecure(fmt.Sprintf("%s:%d", *addr, *port))

	if err := client.Exec(); err != nil {
		log.Fatalf("Exec: %v", err)
	}
}
