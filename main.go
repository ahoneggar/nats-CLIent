package main

import (
	"flag"
	"fmt"
	"github.com/nats-io/go-nats"
)

type options struct {
	test bool
	host string
	tls bool
	cert string
	key string
	ca string
}

func main() {
	opts := parseOpts()
	if opts.test {
		testConn(opts)
		return
	}

	fullClient(opts)
}

func parseOpts() *options {
	ret := &options{}

	ret.tls = *flag.Bool("tls", false, "Use TLS")
	ret.test = *flag.Bool("test", false, "Just test connection, prints pass or fail then returns")
	ret.host = *flag.String("host", "localhost:4222", "address of server to connect to")
	ret.cert = *flag.String("cert", "cert.pem", "Path to the client certificate to use for TLS connection")
	ret.key = *flag.String("key", "key.pem", "Path to the client key to use for TLS connection")
	ret.ca = *flag.String("ca", "ca.pem", "Path to the Certificate Authority to use for TLS connection")

	return ret
}

func testConn(opts *options) {
	if !opts.tls {
		nc, err := nats.Connect(opts.host)
		if err != nil {
			fmt.Errorf("Connection Failed: %+v", err)
			return
		}
		defer nc.Close()
		if nc.IsConnected() {
			fmt.Println("Connection Successful")
			return
		}
		fmt.Errorf("Connection Failed: %+v", err)
		return
	}

	// TODO: TLS connection test
}

func fullClient(opts *options) {
	// TODO: CLI client similar to telnet that parses sub/pub commands
}
