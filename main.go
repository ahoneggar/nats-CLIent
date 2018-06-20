package main

import (
	"flag"
	"fmt"
	"github.com/nats-io/go-nats"
	"strings"
	"time"
	"bufio"
	"os"
)

type options struct {
	test bool
	host string
	tls bool
	cert string
	key string
	ca string
}

const (
	welcome = `
#####################################################################
#                      WELCOME TO NATS-CLIent!                      #
#-------------------------------------------------------------------#
#    PUBLISH   - PUB <subject> <message>                            #
#    SUBSCRIBE - SUB <subject>                                      #
#    REQUEST   - REQ <subject> <message>                            #
#    HELP      - H | HELP (prints this message again)               #
#####################################################################

`
	subUsage = "Sub usage: SUB <subject>\n"
	pubUsage = "Pub usage: PUB <subject> <message>\n"
	reqUsage = "Req usage: REQ <subject> <message>\n"
)

var nc *nats.Conn

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
	ret.host = "nats://" + *flag.String("host", "localhost:4222", "address of server to connect to")
	ret.cert = *flag.String("cert", "cert.pem", "Path to the client certificate to use for TLS connection")
	ret.key = *flag.String("key", "key.pem", "Path to the client key to use for TLS connection")
	ret.ca = *flag.String("ca", "ca.pem", "Path to the Certificate Authority to use for TLS connection")

	return ret
}

func testConn(opts *options) {
	if opts.tls {
		//testConnWithTLS(opts)
	}
	var err error
	nc, err = nats.Connect(opts.host)
	if err != nil {
		fmt.Printf("Connection Failed: %+v", err)
		return
	}
	defer nc.Close()

	if nc.IsConnected() {
		fmt.Println("Connection Successful")
		return
	}
	fmt.Printf("Connection Failed: %+v", err)
	return

	// TODO: TLS connection test
}

func fullClient(opts *options) {
	// TODO: CLI client similar to telnet that parses sub/pub commands
	if opts.tls {
		//fullClientWithTLS(opts)
	}

	var err error

	fmt.Printf("Connecting to %s\n", opts.host)
	nc, err = nats.Connect(opts.host)
	if err != nil {
		fmt.Printf("Connection Failed: %+v", err)
		return
	}
	defer nc.Close()

	fmt.Print(welcome)
	for true {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		line = strings.Trim(line, " \n")
		input := strings.Split(line, " ")
		switch strings.ToUpper(input[0]) {
		case "P", "PUB", "PUBLISH":
			publish(input)
		case "S", "SUB", "SUBSCRIBE":
			subscribe(input)
		case "R", "REQ", "REQUEST":
			request(input)
		case "H", "HELP":
			fmt.Print(welcome)
		default:
			fmt.Println("Unrecognized command")
		}
	}
}

func publish(input []string) {
	if len(input) < 3 {
		fmt.Print(pubUsage)
		return
	}
	err := nc.Publish(input[1], []byte(strings.Join(input[2:], " ")))
	if err != nil {
		fmt.Printf("Error Publishing: %+v\n", err)
	}
}

func subscribe(input []string) {
	if len(input) < 2 {
		fmt.Print(subUsage)
		return
	}
	_, err := nc.Subscribe(input[1], handleIncomingMessage)
	if err != nil {
		fmt.Printf("Error Subscribing: %+v\n", err)
		return
	}
	fmt.Printf("+OK\n")
}

func request(input []string) {
	if len(input) < 3 {
		fmt.Print(reqUsage)
		return
	}
	msg, err := nc.Request(input[1], []byte(strings.Join(input[2:], " ")), 10*time.Millisecond)
	if err != nil {
		fmt.Printf("Error Requesting: %+v\n", err)
		return
	}
	fmt.Printf("Response: %s", string(msg.Data))
}

func handleIncomingMessage(m *nats.Msg) {
	fmt.Printf("\n+MSG %s: %s\n>", m.Subject, string(m.Data))
}
