package server

import (
	"fmt"
	"log"
	"net"
)

// writes back to Client, needs connection and string
func writer(conn net.Conn, s string) {
	_, err := fmt.Fprintf(conn, "%s\n>> ", s)
	if err != nil {
		log.Fatal(err)
	}
}
