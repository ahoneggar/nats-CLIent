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

// Stores program options
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
#    QUIT      - Q | QUIT                                           #
#####################################################################

`
	subUsage = "Sub usage: SUB <subject>\n"
	pubUsage = "Pub usage: PUB <subject> <message>\n"
	reqUsage = "Req usage: REQ <subject> <message>\n"
)

var nc *nats.Conn

func main() {
	opts := parseOpts()

	if opts == nil {
		return
	}

	if opts.test {
		testConn(opts)
		return
	}

	fullClient(opts)
}

// Get command-line args
func parseOpts() *options {
	// Set flags
	tls := flag.Bool("tls", false, "Use TLS")
	test := flag.Bool("test", false, "Just test connection, prints pass or fail then returns")
	hostPtr := flag.String("host", "localhost:4222", "address of server to connect to")
	cert := flag.String("cert", "", "Path to the client certificate to use for TLS connection")
	key := flag.String("key", "", "Path to the client key to use for TLS connection")
	ca := flag.String("ca", "", "Path to the Certificate Authority to use for TLS connection")
	help := flag.Bool("help", false, "Print flag usage")
	h := flag.Bool("h", false, "Same as -help")

	flag.Parse()

	// If help, print help and exit
	if *help || *h {
		flag.Usage()
		return nil
	}

	// If TLS, use tls:// instead of nats://
	host := ""
	if *tls {
		host = "tls://" + *hostPtr
	} else {
		host = "nats://" + *hostPtr
	}

	return &options{tls: *tls, test: *test, host: host, cert: *cert, key: *key, ca: *ca}
}

// Just test connection, not provide client
func testConn(opts *options) {
	var err error
	if opts.tls {
		err = connectTLS(opts)
	} else {
		nc, err = nats.Connect(opts.host)
	}
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
}

// Run the actual CLI client
func fullClient(opts *options) {
	var err error

	fmt.Printf("Connecting to %s\n", opts.host)
	if opts.tls {
		err = connectTLS(opts)
	} else {
		nc, err = nats.Connect(opts.host)
	}
	if err != nil {
		fmt.Printf("Connection Failed: %+v", err)
		return
	}
	defer nc.Close()

	fmt.Print(welcome)

	// Loop the CLI
	for true {
		fmt.Print("> ")
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		line = strings.Trim(line, " \n")
		input := strings.Split(line, " ")

		// Parse input
		switch strings.ToUpper(input[0]) {
		case "P", "PUB", "PUBLISH":
			publish(input)
		case "S", "SUB", "SUBSCRIBE":
			subscribe(input)
		case "R", "REQ", "REQUEST":
			request(input)
		case "H", "HELP":
			fmt.Print(welcome)
		case "Q", "QUIT":
			break
		default:
			fmt.Println("Unrecognized command")
		}
	}
}

// Make necessary TLS connection options then connect
func connectTLS(opts *options) error {
	var err error
	srvOpts := make([]nats.Option, 0)
	if opts.ca != "" {
		srvOpts = append(srvOpts, nats.RootCAs(opts.ca))
	}
	if opts.cert != "" {
		srvOpts = append(srvOpts, nats.ClientCert(opts.cert, opts.key))
	}

	nats.Connect(opts.host, srvOpts...)
	if err != nil {
		return err
	}
	return nil
}

// Publish input[2:] to input[1]
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

// Subscribe to input[1]
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

// Request input[2:] from input[1]
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

// Print incoming messages
func handleIncomingMessage(m *nats.Msg) {
	fmt.Printf("\n+MSG %s: %s\n>", m.Subject, string(m.Data))
}
