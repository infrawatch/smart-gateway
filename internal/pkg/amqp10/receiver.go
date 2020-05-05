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
	"os/signal"
	"reflect"
	"strings"
	"sync"

	"github.com/infrawatch/smart-gateway/internal/pkg/cacheutil"
	"github.com/infrawatch/smart-gateway/internal/pkg/saconfig"
	"github.com/prometheus/client_golang/prometheus"
	"qpid.apache.org/amqp"
	"qpid.apache.org/electron"
)

var debugr = func(format string, data ...interface{}) {} // Default no debugging output

//AMQPServer msgcount -1 is infinite
type AMQPServer struct {
	urlStr          string
	debug           bool
	msgcount        int
	notifier        chan string
	status          chan int
	done            chan bool
	connection      electron.Connection
	method          func(s *AMQPServer) (electron.Receiver, error)
	prefetch        int
	amqpHandler     *AMQPHandler
	uniqueName      string
	reconnect       bool
	collectinterval float64
}

//AMQPServerItem hold information about data source which is AMQPServer listening to.
type AMQPServerItem struct {
	Server     *AMQPServer
	DataSource saconfig.DataSource
}

//AMQPHandler ...
type AMQPHandler struct {
	totalCount              int
	totalProcessed          int
	totalReconnectCount     int
	totalCountDesc          *prometheus.Desc
	totalProcessedDesc      *prometheus.Desc
	totalReconnectCountDesc *prometheus.Desc
}

//NewAMQPServer   ...
func NewAMQPServer(urlStr string, debug bool, msgcount int, prefetch int, amqpHanlder *AMQPHandler, uniqueName string) *AMQPServer {
	if len(urlStr) == 0 {
		log.Println("No URL provided")
		//usage()
		os.Exit(1)
	}
	server := &AMQPServer{
		urlStr:          urlStr,
		debug:           debug,
		notifier:        make(chan string),
		status:          make(chan int),
		done:            make(chan bool),
		msgcount:        msgcount,
		method:          (*AMQPServer).connect,
		prefetch:        prefetch,
		amqpHandler:     amqpHanlder,
		uniqueName:      uniqueName,
		reconnect:       false,
		collectinterval: 30,
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

//GetHandler  ...
func (s *AMQPServer) GetHandler() *AMQPHandler {
	return s.amqpHandler
}

//NewAMQPHandler  ...
func NewAMQPHandler(source string) *AMQPHandler {
	plabels := prometheus.Labels{}
	plabels["source"] = source
	return &AMQPHandler{
		totalCount:          0,
		totalProcessed:      0,
		totalReconnectCount: 0,
		totalCountDesc: prometheus.NewDesc("collectd_total_amqp_message_recv_count",
			"Total count of amqp message received.",
			nil, plabels,
		),
		totalProcessedDesc: prometheus.NewDesc("collectd_total_amqp_processed_message_count",
			"Total count of amqp message processed.",
			nil, plabels,
		),
		totalReconnectCountDesc: prometheus.NewDesc("collectd_total_amqp_reconnect_count",
			"Total count of amqp reconnection .",
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

//IncTotalReconnectCount ...
func (a *AMQPHandler) IncTotalReconnectCount() {
	a.totalReconnectCount++
}

//GetTotalMsgRcv ...
func (a *AMQPHandler) GetTotalMsgRcv() int {
	return a.totalCount
}

//GetTotalMsgProcessed ...
func (a *AMQPHandler) GetTotalMsgProcessed() int {
	return a.totalProcessed
}

//GetTotalReconnectCount ...
func (a *AMQPHandler) GetTotalReconnectCount() int {
	return a.totalReconnectCount
}

//Describe ...
func (a *AMQPHandler) Describe(ch chan<- *prometheus.Desc) {
	ch <- a.totalCountDesc
	ch <- a.totalProcessedDesc
	ch <- a.totalReconnectCountDesc
}

//Collect implements prometheus.Collector.
func (a *AMQPHandler) Collect(ch chan<- prometheus.Metric) {
	ch <- prometheus.MustNewConstMetric(a.totalCountDesc, prometheus.CounterValue, float64(a.totalCount))
	ch <- prometheus.MustNewConstMetric(a.totalProcessedDesc, prometheus.CounterValue, float64(a.totalProcessed))
	ch <- prometheus.MustNewConstMetric(a.totalReconnectCountDesc, prometheus.CounterValue, float64(a.totalReconnectCount))
}

//GetNotifier  Get notifier
func (s *AMQPServer) GetNotifier() chan string {
	return s.notifier
}

//GetStatus  Get Status
func (s *AMQPServer) GetStatus() chan int {
	return s.status
}

//GetDoneChan ...
func (s *AMQPServer) GetDoneChan() chan bool {
	return s.done
}

//Close connections it is exported so users can force close
func (s *AMQPServer) Close() {
	s.connection.Close(nil)
	debugr("Debug: close receiver connection %s", s.connection)
}

//UpdateMinCollectInterval ...
func (s *AMQPServer) UpdateMinCollectInterval(interval float64) {
	if interval < s.collectinterval {
		s.collectinterval = interval
	}
}

//start  starts amqp server
func (s *AMQPServer) start() {
	msgBuffCount := 10
	if s.msgcount > 0 {
		msgBuffCount = s.msgcount
	}
	messages := make(chan amqp.Message, msgBuffCount) // Channel for messages from goroutines to main()
	connectionStatus := make(chan int)
	done := make(chan bool)

	defer close(done)
	defer close(messages)
	defer close(connectionStatus)

	go func() {
		r, err := s.method(s)
		if err != nil {
			log.Fatalf("Could not connect to Qpid-dispatch router. is it running? : %v", err)
		}
		connectionStatus <- 1
		untilCount := s.msgcount
	theloop:
		for {
			if rm, err := r.Receive(); err == nil {
				rm.Accept()
				debugr("Message ACKed: %v", rm.Message)
				messages <- rm.Message
			} else if err == electron.Closed {
				log.Printf("Channel closed...\n")
				return
			} else {
				log.Fatalf("Received error %v: %v", s.urlStr, err)
			}
			if untilCount > 0 {
				untilCount--
			}
			if untilCount == 0 {
				break theloop
			}
		}
		done <- true
		s.done <- true
		s.Close()
		log.Println("Closed AMQP...letting loop know")
	}()

msgloop:
	for {
		select {
		case <-done:
			debugr("Done received...\n")
			break msgloop
		case m := <-messages:
			debugr("Message received... %v\n", m.Body())
			handler := s.GetHandler()
			if handler != nil {
				handler.IncTotalMsgRcv()
			}
			switch msg := m.Body().(type) {
			case amqp.Binary:
				s.notifier <- msg.String()
			case string:
				s.notifier <- msg
			default:
				// do nothing and report
				log.Printf("Invalid type of AMQP message received: %t", msg)
			}
		case status := <-connectionStatus:
			debugr("Status received...\n")
			s.status <- status
		}
	}
}

//connect Connect to an AMQP server returning an electron receiver
func (s *AMQPServer) connect() (electron.Receiver, error) {
	// Wait for one goroutine per URL
	// Make name unique-ish
	container := electron.NewContainer(fmt.Sprintf("rcv[%v]", s.uniqueName))
	url, err := amqp.ParseURL(s.urlStr)
	fatalIf(err)
	c, err := container.Dial("tcp", url.Host) // NOTE: Dial takes just the Host part of the URL
	if err != nil {
		log.Printf("AMQP Dial tcp %v\n", err)
		return nil, err
	}

	s.connection = c // Save connection so we can Close() when start() ends

	addr := strings.TrimPrefix(url.Path, "/")
	opts := []electron.LinkOption{electron.Source(addr)}
	if s.prefetch > 0 {
		debugr("Amqp Prefetch set to %d\n", s.prefetch)
		opts = append(opts, electron.Capacity(s.prefetch), electron.Prefetch(true))
	}

	r, err := c.Receiver(opts...)
	return r, err
}

func fatalIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//SpawnSignalHandler spawns goroutine which will wait for interruption signal(s)
// and end smart gateway in case any of the signal is received
func SpawnSignalHandler(finish chan bool, watchedSignals ...os.Signal) {
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, watchedSignals...)
	go func() {
	signalLoop:
		for sig := range interruptChannel {
			log.Printf("Stopping execution on caught signal: %+v\n", sig)
			close(finish)
			break signalLoop
		}
	}()
}

//SpawnQpidStatusReporter builds dynamic select for reporting status of AMQP connections
func SpawnQpidStatusReporter(wg *sync.WaitGroup, applicationHealth *cacheutil.ApplicationHealthCache, qpidStatusCases []reflect.SelectCase) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		finishCase := len(qpidStatusCases) - 1
	statusLoop:
		for {
			switch index, status, _ := reflect.Select(qpidStatusCases); index {
			case finishCase:
				break statusLoop
			default:
				// Note: status here is always very low integer, so we don't need to be afraid of int64>int conversion
				applicationHealth.QpidRouterState = int(status.Int())
			}
		}
		log.Println("Closing QPID status reporter")
	}()
}

//CreateMessageLoopComponents creates signal select cases for configured AMQP1.0 connections and connects to all of thos
func CreateMessageLoopComponents(config interface{}, finish chan bool, amqpHandler *AMQPHandler, uniqueName string) ([]reflect.SelectCase, []reflect.SelectCase, []AMQPServerItem) {
	var (
		debug       bool
		prefetch    int
		connections []saconfig.AMQPConnection
	)
	switch config.(type) {
	case *saconfig.EventConfiguration:
		conf := config.(*saconfig.EventConfiguration)
		debug = conf.Debug
		prefetch = conf.Prefetch
		connections = conf.AMQP1Connections
	case *saconfig.MetricConfiguration:
		conf := config.(*saconfig.MetricConfiguration)
		debug = conf.Debug
		prefetch = conf.Prefetch
		connections = conf.AMQP1Connections
	default:
		panic("Invalid type of configuration file struct.")
	}

	processingCases := make([]reflect.SelectCase, 0, len(connections))
	qpidStatusCases := make([]reflect.SelectCase, 0, len(connections))
	amqpServers := make([]AMQPServerItem, 0, len(connections))
	for _, conn := range connections {
		amqpServer := NewAMQPServer(conn.URL, debug, -1, prefetch, amqpHandler, uniqueName)
		//create select case for this listener
		processingCases = append(processingCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(amqpServer.GetNotifier()),
		})
		qpidStatusCases = append(qpidStatusCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(amqpServer.GetStatus()),
		})
		amqpServers = append(amqpServers, AMQPServerItem{amqpServer, conn.DataSourceID})
	}
	log.Println("Listening for AMQP1.0 messages")
	// include also case for finishing the loops
	processingCases = append(processingCases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(finish),
	})
	qpidStatusCases = append(qpidStatusCases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(finish),
	})
	return processingCases, qpidStatusCases, amqpServers
}
