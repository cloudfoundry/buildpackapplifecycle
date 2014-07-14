package main

import (
	"flag"
	"net"
	"os"
	"time"
    "fmt"
)

var network = flag.String(
	"network",
	"tcp",
	"network type to dial with (e.g. unix, tcp)",
)

var addr = flag.String(
	"addr",
	":8080",
	"address to dial",
)

var timeout = flag.Duration(
	"timeout",
	1*time.Second,
	"dial timeout",
)

func main() {
	flag.Parse()

	conn, err := net.DialTimeout(*network, *addr, *timeout)
	if err != nil {
		os.Stderr.Write([]byte(fmt.Sprintf("healthcheck failed: %s\n", err.Error())))
		os.Exit(1)
	}

	conn.Close()

	os.Stdout.Write([]byte("healtcheck passed"))
	os.Exit(0)
}
