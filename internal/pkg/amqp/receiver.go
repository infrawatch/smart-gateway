/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package amqp10

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"qpid.apache.org/amqp"
	"qpid.apache.org/electron"
)

// Usage and command-line flags
/*func usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s url [url ...]
Receive messages from all URLs concurrently and print them.
URLs are of the form "amqp://<host>:<port>/<amqp-address>"
`, os.Args[0])
	flag.PrintDefaults()
}*/

var debugr = func(format string, data ...interface{}) {} // Default no debugging output

//AMQPServer msgcount -1 is infinite
type AMQPServer struct {
	urlStr      string
	debug       bool
	msgcount    int
	notifier    chan string
	status      chan int
	connections chan electron.Connection
}

//MockAmqpServer  Create Mock AMQP server
func MockAmqpServer(notifier chan string) *AMQPServer {
	server := &AMQPServer{
		notifier: notifier,
	}
	return server
}

//NewAMQPServer   ...
func NewAMQPServer(urlStr string, debug bool, msgcount int) *AMQPServer {
	if len(urlStr) == 0 {
		log.Println("No URL provided")
		//usage()
		os.Exit(1)
	}
	server := &AMQPServer{
		urlStr:      urlStr,
		debug:       debug,
		notifier:    make(chan string),
		status:      make(chan int),
		msgcount:    msgcount,
		connections: make(chan electron.Connection, 1),
	}
	if debug {
		debugr = func(format string, data ...interface{}) {
			log.Printf(format, data...)
		}
	}
	// Spawn off the server's main loop immediately
	// not exported
	go server.start()
	return server
}

//GetNotifier  Get notifier
func (s *AMQPServer) GetNotifier() chan string {
	return s.notifier
}

//GetStatus  Get Status
func (s *AMQPServer) GetStatus() chan int {
	return s.status
}

//Close connections it is exported so users can force close
func (s *AMQPServer) Close() {
	select {
	case c := <-s.connections:
		log.Printf("%s\n", c)
		debugr("Debug:close %s", c)
		c.Close(nil)
	default:
		return
	}
}

//start  starts amqp server
func (s *AMQPServer) start() {
	messages := make(chan amqp.Message) // Channel for messages from goroutines to main()
	connectionStatus := make(chan int)
	defer close(messages)
	defer close(connectionStatus)
	//var wait sync.WaitGroup // Used by main() to wait for all goroutines to end.
	//wait.Add(1)
	go func() {
		/*if *prefetch > 0 {
		  opts = append(opts, electron.Capacity(*prefetch), electron.Prefetch(true))
		}*/
		r, err := s.connect()
		if err != nil {
			log.Fatalf("Could not connect to Qpid-dispatch router. is it running? : %v", err)
		}
		s.status <- 1
		// Loop receiving messages and sending them to the main() goroutine
		if s.msgcount == -1 {
			for {
				if rm, err := r.Receive(); err == nil {
					rm.Accept()
					debugr("AMQP Receiving messages.")
					messages <- rm.Message
				} else if err == electron.Closed {
					log.Fatalf("Connection Closed %v: %v", s.urlStr, err)
					connectionStatus <- 0
					return
				} else {
					log.Printf("AMQP Listener error, will try to reconnect %v: %v\n", s.urlStr, err)
					debugr("AMQP Receiver trying to connect")
					connectionStatus <- 0

				CONNECTIONLOOP:
					for {
						s.Close()
						log.Println("Reconnect attempt in 2 secs")
						time.Sleep(2 * time.Second)
						r, err = s.connect()
						debugr("%v", err)
						if err == nil {
							log.Println("Connection with QDR established.")
							connectionStatus <- 1
							break CONNECTIONLOOP
						}
					}

				}
			}
		} else {
			for i := 0; i < s.msgcount; i++ {
				if rm, err := r.Receive(); err == nil {
					rm.Accept()
					messages <- rm.Message
				} else if err == electron.Closed {
					return
				} else {
					log.Fatalf("receive error %v: %v", s.urlStr, err)
				}
			}
			s.Close()
		}
	}()
	//outside go routin reciveve and process
	//var firstmsg=0
	for {
		select {
		case m := <-messages:
			debugr("Debug: Getting message from AMQP%v\n", m.Body())
			amqpBinary := m.Body().(amqp.Binary)
			debugr("Debug: Sending message to Notifier channel")
			s.notifier <- amqpBinary.String()
			continue //priority to process exiting messages
		case status := <-connectionStatus:
			s.status <- status
		default: //default makes this non-blocking
		}
	}
	//wait.Wait() // Wait for all goroutines to finish.
}

func (s *AMQPServer) connect() (electron.Receiver, error) {
	// Wait for one goroutine per URL
	container := electron.NewContainer(fmt.Sprintf("receive[%v]", os.Getpid()))
	//connections := make(chan electron.Connection, 1) // Connections to close on exit
	url, err := amqp.ParseURL(s.urlStr)
	debugr("Parsing %s\n", s.urlStr)
	fatalIf(err)
	c, err := container.Dial("tcp", url.Host) // NOTE: Dial takes just the Host part of the URL
	if err != nil {
		log.Printf("AMQP Dial tcp %v\n", err)
		return nil, err
	}

	s.connections <- c // Save connection so we can Close() when start() ends
	addr := strings.TrimPrefix(url.Path, "/")
	opts := []electron.LinkOption{electron.Source(addr)}
	/*if *prefetch > 0 {
		opts = append(opts, electron.Capacity(*prefetch), electron.Prefetch(true))
	}*/
	r, err := c.Receiver(opts...)
	return r, err
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
