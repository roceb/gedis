package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/roceb/gedis/client"
	"github.com/roceb/gedis/server"
)

var (
	cli *client.Client
	srv *server.Server
)

func main() {
	// creates chan to reviece stop signal
	stop := make(chan os.Signal, 1)
	// registers the given channel to receive notifications of the specified signals.
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	// adds -cli flag and sets it to false
	args := flag.Bool("cli", false, "To activate client")
	// parse for -cli flag
	flag.Parse()

	// checks if cli is true to start cli, if false starts server
	switch *args {
	case true:
		cli = client.NewClient()
	default:
		srv = server.NewServer()
	}
	// send stop signal through chan
	// using a select here because I would like to add more chans in future
	select {
	case <-stop:
		cli.Stop()
		srv.Stop()
	}
}
