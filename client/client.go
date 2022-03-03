package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Client struct {
	conn net.Conn      //active tcp Connection
	addr string        // used to keep track of address
	quit chan struct{} // channel used to send quit command
}

// returns a new Client. Could be changed to use a different address
func NewClient() *Client {
	// this is not needed since we will always have the correct address
	tcpAddr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}
	// Creates a connection with server
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal("Unable to connect")
	}

	client := &Client{
		addr: "localhost:8080",
		quit: make(chan struct{}),
		conn: conn,
	}
	// connects client
	go client.serve()

	return client
}

// connects a client
func (c *Client) serve() {
	// creates a loop that reads and writes to STDIN
	for {
		// Output is text sent from server
		output, _ := bufio.NewReader(c.conn).ReadString('\n')
		fmt.Print("" + output)
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(">> ")
		// Input is from user
		input, _ := reader.ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		fmt.Fprintf(c.conn, fmt.Sprintf("%v \n", input))
		// Checks to see if the first word sent is END
		if strings.Split((strings.TrimSpace(string(input))), " ")[0] == "END" {
			fmt.Println("TCP client exiting...")
			// Closes client
			c.Stop()
			return
		}
	}

}

// Closes client
func (c *Client) Stop() {
	fmt.Println("Gedis CLI is shutting down")
	c.conn.Close()
	close(c.quit)
	os.Exit(0)
}
