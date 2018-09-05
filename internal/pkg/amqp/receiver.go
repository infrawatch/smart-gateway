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
	"math"
	"net"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
	done        chan bool
	connections chan electron.Connection
	method      func(s *AMQPServer) (electron.Receiver, error)
	prefetch    int
	amqpHandler *AMQPHandler
	uniqueName  string
}

//AMQPHandler ...
type AMQPHandler struct {
	totalCount         int64
	totalProcessed     int64
	totalCountDesc     *prometheus.Desc
	totalProcessedDesc *prometheus.Desc
}

//MockAmqpServer  Create Mock AMQP server
func MockAmqpServer(notifier chan string) *AMQPServer {
	server := &AMQPServer{
		notifier: notifier,
	}
	return server
}

//NewAMQPServer   ...
func NewAMQPServer(urlStr string, debug bool, msgcount int, prefetch int, amqpHanlder *AMQPHandler, done chan bool, isTest bool, uniqueName string) *AMQPServer {
	if len(urlStr) == 0 {
		log.Println("No URL provided")
		//usage()
		os.Exit(1)
	}
	var server *AMQPServer

	if isTest == true {
		server = &AMQPServer{
			urlStr:      "127.0.0.1:5672",
			debug:       debug,
			notifier:    make(chan string, 100),
			status:      make(chan int),
			msgcount:    msgcount,
			connections: make(chan electron.Connection, 1),
			method:      (*AMQPServer).serverTest,
			done:        done,
			prefetch:    prefetch,
			amqpHandler: amqpHanlder,
			uniqueName:  uniqueName,
		}
	} else {
		server = &AMQPServer{
			urlStr:      urlStr,
			debug:       debug,
			notifier:    make(chan string),
			status:      make(chan int),
			msgcount:    msgcount,
			connections: make(chan electron.Connection, 1),
			method:      (*AMQPServer).connect,
			done:        done,
			prefetch:    prefetch,
			amqpHandler: amqpHanlder,
			uniqueName:  uniqueName,
		}
	}

	if debug {
		debugr = func(format string, data ...interface{}) {
			log.Printf(format, data...)
		}
	}
	// Spawn off the server's main loop immediately
	// not exported
	go server.start(isTest)

	return server
}

//GetHandler  ...
func (s *AMQPServer) GetHandler() *AMQPHandler {
	return s.amqpHandler
}

//NewAMQPHandler  ...
func NewAMQPHandler(source string) *AMQPHandler {
	plabels := prometheus.Labels{}
	plabels["source"] = source
	return &AMQPHandler{
		totalCount:     0,
		totalProcessed: 0,
		totalCountDesc: prometheus.NewDesc("sa_collectd_total_amqp_message_recv_count",
			"Total count of amqp message received.",
			nil, plabels,
		),
		totalProcessedDesc: prometheus.NewDesc("sa_collectd_total_amqp_processed_message_count",
			"Total count of amqp message processed.",
			nil, plabels,
		),
	}
}

//IncTotalMsgRcv ...
func (a *AMQPHandler) IncTotalMsgRcv() {
	a.totalCount++
}

//IncTotalMsgProcessed ...
func (a *AMQPHandler) IncTotalMsgProcessed() {
	a.totalProcessed++
}

//GetTotalMsgRcv ...
func (a *AMQPHandler) GetTotalMsgRcv() int64 {
	return a.totalCount
}

//GetTotalMsgProcessed ...
func (a *AMQPHandler) GetTotalMsgProcessed() int64 {
	return a.totalProcessed
}

//Describe ...
func (a *AMQPHandler) Describe(ch chan<- *prometheus.Desc) {

	ch <- a.totalCountDesc
	ch <- a.totalProcessedDesc
}

//Collect implements prometheus.Collector.
func (a *AMQPHandler) Collect(ch chan<- prometheus.Metric) {

	ch <- prometheus.MustNewConstMetric(a.totalCountDesc, prometheus.CounterValue, float64(a.totalCount))
	ch <- prometheus.MustNewConstMetric(a.totalProcessedDesc, prometheus.CounterValue, float64(a.totalProcessed))

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
		log.Printf("Closing...%s\n", c)
		debugr("Debug:close %s", c)
		c.Close(nil)
	default:
		return
	}
}

//start  starts amqp server
func (s *AMQPServer) start(isTest bool) {
	messages := make(chan amqp.Message, 10) // Channel for messages from goroutines to main()
	connectionStatus := make(chan int)
	done := make(chan bool)

	defer close(messages)
	defer close(connectionStatus)
	//var wait sync.WaitGroup // Used by main() to wait for all goroutines to end.
	//wait.Add(1)
	go func() {
		/*if *prefetch > 0 {
		  opts = append(opts, electron.Capacity(*prefetch), electron.Prefetch(true))
		}*/
		r, err := s.method(s)
		if err != nil {
			log.Fatalf("Could not connect to Qpid-dispatch router. is it running? : %v", err)
		}
		//s.status <- 1
		connectionStatus <- 1
		// Loop receiving messages and sending them to the main() goroutine
		if isTest == false && s.msgcount == -1 {
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
			untilCount := s.msgcount
			if s.msgcount == -1 {
				untilCount = math.MaxInt32
			}
			for i := 0; i < untilCount; i++ {
				if rm, err := r.Receive(); err == nil {
					rm.Accept()
					messages <- rm.Message
				} else if err == electron.Closed {
					log.Printf("Channel closed...\n")
					return
				} else {
					log.Fatalf("receive error %v: %v, %d", s.urlStr, err, i)
				}
			}
			s.Close()
			log.Println("Closed AMQP...letting loop know")
			done <- true
			s.done <- true
		}
	}()
	//outside go routin reciveve and process
	//var firstmsg=0
msgloop:
	for {
		select {
		case m := <-messages:
			s.GetHandler().IncTotalMsgRcv()
			debugr("Debug: Getting message from AMQP%v\n", m.Body())
			amqpBinary := m.Body().(amqp.Binary)
			debugr("Debug: Sending message to Notifier channel")
			s.notifier <- amqpBinary.String()
			continue //priority to process exiting messages
		case status := <-connectionStatus:
			s.status <- status
		case <-done:
			log.Println("Done Received...")
			break msgloop

			//		default: //default makes this non-blocking
		}
	}

	//wait.Wait() // Wait for all goroutines to finish.
}

func (s *AMQPServer) connect() (electron.Receiver, error) {
	// Wait for one goroutine per URL
	// Make name unique-ish
	container := electron.NewContainer(fmt.Sprintf("rcv[%v]", s.uniqueName))
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
	if s.prefetch > 0 {
		debugr("Amqp Prefetch set to %d\n", s.prefetch)
		opts = append(opts, electron.Capacity(s.prefetch), electron.Prefetch(true))
	}

	r, err := c.Receiver(opts...)
	return r, err
}

//
// serverTest -- Creates a local listening endpoint
//    that allows a single test client to attach and send
//    amqp / collectd test data
//
func (s *AMQPServer) serverTest() (electron.Receiver, error) {
	l, err := net.Listen("tcp", s.urlStr) // tcp4 so example will work on ipv6-disabled platforms
	if err != nil {
		log.Fatal(err)
	}

	cont := electron.NewContainer(fmt.Sprintf("receive[%v]", os.Getpid()))

	log.Printf("Listening for connection at...%s\n", s.urlStr)

	c, err := cont.Accept(l)
	if err != nil {
		log.Fatal(err)
	}

	l.Close()

	// Process incoming endpoints till we get a Receiver link
	var r electron.Receiver
	for r == nil {
		in := <-c.Incoming()
		switch in := in.(type) {
		case *electron.IncomingSession, *electron.IncomingConnection:
			in.Accept() // Accept the incoming connection and session for the receiver
			log.Printf("Accepted incomming session...%v\n", in)

		case *electron.IncomingReceiver:
			if s.prefetch > 0 {
				in.SetCapacity(s.prefetch)
				in.SetPrefetch(true) // Automatic flow control for a buffer of 10 messages.
			}
			r = in.Accept().(electron.Receiver)
			log.Printf("Accepted incomming receiever...%v", r)
		case nil:
			return nil, err // Connection is closed
		default:
			in.Reject(amqp.Errorf("example-server", "unexpected endpoint %v", in))
		}
	}
	go func() { // Reject any further incoming endpoints
		for in := range c.Incoming() {
			in.Reject(amqp.Errorf("example-server", "unexpected endpoint %v", in))
		}
	}()

	s.connections <- c // Save connection so we can Close() when start() ends

	log.Println("Start processing messages...")

	return r, err
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
