package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type Server struct {
	listener         net.Listener     // active server connected
	db               gedis            // db stored
	totalConnections map[int]net.Conn // keeps track of all active clients
	connTimeout      time.Duration    // connection timeout
	quit             chan struct{}    // send quit command
	killed           chan struct{}    // send kill command to clients
}

// returns New server
func NewServer() *Server {
	// listening to port 8080
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("unable to connect to listener")
	}
	server := &Server{
		listener:         l,
		quit:             make(chan struct{}),
		db:               newDB(),
		totalConnections: map[int]net.Conn{},
	}
	// start server
	go server.serve()
	return server
}

// Serve function for the server
func (s *Server) serve() {
	// keep track of clients
	var clientid int
	for {
		// if quit command send, disconnect clients
		select {
		case <-s.quit:
			err := s.listener.Close()
			if err != nil {
				fmt.Println("Unable to quit listener")
			}
			if len(s.totalConnections) > 0 {
				fmt.Printf("Closing connection to ")
				<-time.After(s.connTimeout)
				s.closeConnections()
			}
			close(s.killed)
			return
		// create a new connection with new clients
		default:
			listener := s.listener.(*net.TCPListener)
			err := listener.SetDeadline(time.Now().Add(2 * time.Second))
			if err != nil {
				fmt.Println("Unable to set deadline for listener")
			}
			conn, err := listener.Accept()
			opErr, ok := err.(*net.OpError)
			// if no errors, connect
			if ok && opErr.Timeout() {
				continue
			}
			if err != nil {
				fmt.Println("Unable to accept connection. ", err.Error())
			}
			// tell client they have connected
			writer(conn, "Tech Assessment DB")
			// set client conn to client id in server
			s.totalConnections[clientid] = conn
			// goroutine anonymous function that gives conn to handler function

			go func(id int) {
				fmt.Println("client ", id, "joined")
				s.handler(conn)
				delete(s.totalConnections, id)
				fmt.Println("client with id", id, "left")
			}(clientid)
			// increase client id to add be ready for another connection
			clientid++
		}

	}
}

// handles commands sent by client
func (s *Server) handler(conn net.Conn) {
	// enqueue is used to track to see if we are in a transaction
	enqueue := false
	// commitID keeps track of multiple nested transaction
	var commitID = -1
	// sc = data sent from client
	sc := bufio.NewScanner(conn)
	for sc.Scan() {
		input := strings.TrimSpace(sc.Text())
		data := strings.Split(input, " ")
		// cmd = first word from client
		cmd := strings.ToUpper(data[0])
		switch {
		case (cmd == "SET" && len(data) == 3) && !enqueue:
			s.db.set(data[1], data[2])
			writer(conn, "")
		case (cmd == "GET" && len(data) == 2):
			v, ok := s.db.get(data[1])
			if !ok {
				writer(conn, "NULL")
			} else {
				writer(conn, v)
			}
		case (cmd == "DELETE" && len(data) == 2) && !enqueue:
			s.db.delete(data[1])
			writer(conn, "")
		case (cmd == "COUNT" && len(data) == 2):
			v := s.db.count(data[1])
			writer(conn, fmt.Sprintf("%d", v))
		case (cmd == "END" && len(data) == 1) && !enqueue:
			if err := conn.Close(); err != nil {
				fmt.Println("could not end", err.Error())
			}
		case (cmd == "BEGIN" && len(data) == 1):
			// sets up transaction
			enqueue = true
			commitID++
			writer(conn, s.db.begin())
		case (cmd == "ROLLBACK" && len(data) == 1):
			enqueue = false
			trxId := len(s.db.trans) - 1
			req := s.db.rollback(trxId)
			// if there are still transaction in queue, keep enqueue active
			if len(s.db.trans) > 0 {
				enqueue = true
			}
			writer(conn, req)
		case (cmd == "COMMIT" && len(data) == 1):
			// msgQue are commands that need to be sent to client after transaction
			msgQue := []string{}
			// used to track which transaction is being processed
			trxID := 0
			for len(s.db.trans) > 0 {
				trx := s.db.trans[trxID]
				delete(s.db.trans, trxID)
				for len(trx) > 0 {
					// takes the first queued trx
					trxData := trx[0]
					// removes trx from queue
					trx = trx[1:]
					unparsedData := strings.Split(trxData, " ")
					// commit command
					cCmd := strings.ToUpper(unparsedData[0])
					// commit data
					cData := unparsedData[1:]
					switch {
					case cCmd == "SET" && len(cData) == 2:
						s.db.set(cData[0], cData[1])
					case cCmd == "GET" && len(cData) == 1:
						cGet, ok := s.db.get(cData[0])
						if !ok {
							msgQue = append(msgQue, "NULL")
						}
						msgQue = append(msgQue, cGet)
						// writer(conn, cGet)
					case cCmd == "DELETE" && len(cData) == 1:
						s.db.delete(cData[0])
					case cCmd == "COUNT" && len(cData) == 1:
						v := s.db.count(cData[0])
						msgQue = append(msgQue, fmt.Sprint(v))
					default:
						writer(conn, "Unknown command")
					}
				}
				trxID++

			}
			enqueue = false

			writer(conn, strings.Join(msgQue, ", "))
		case (cmd != "BEGIN" && enqueue):
			// this will only run if currently in a trx and receive the begin command
			s.db.trans[commitID] = append(s.db.trans[commitID], input)
			writer(conn, "Added")
		default:
			writer(conn, "Unknown command")
		}
	}
}

// closes connections
func (s *Server) closeConnections() {
	fmt.Println("killing connect clients")
	for clientid, conn := range s.totalConnections {
		err := conn.Close()
		if err != nil {
			fmt.Printf("Error killing: %v \n", clientid)
		}
	}
}

// stops server
func (s *Server) Stop() {
	// would like to add a save command here
	fmt.Println("Gedis is shutting down")
	close(s.quit)
	<-s.killed
	fmt.Println("Gedis successfully killed")
}
