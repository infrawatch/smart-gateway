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

	"qpid.apache.org/amqp"
	"qpid.apache.org/electron"
)

var debugsf = func(format string, data ...interface{}) {} // Default no debugging output

//AMQPSender msgcount -1 is infinite
type AMQPSender struct {
	urlStr      string
	debug       bool
	connections chan electron.Connection
}

//MockAMQPSender  Create Mock AMQP server
func MockAMQPSender(notifier chan string) *AMQPServer {
	server := &AMQPServer{}
	return server
}

//NewAMQPSender   ...
func NewAMQPSender(urlStr string, debug bool) *AMQPSender {
	if len(urlStr) == 0 {
		log.Println("No URL provided")
		//usage()
		os.Exit(1)
	}
	server := &AMQPSender{
		urlStr:      urlStr,
		debug:       debug,
		connections: make(chan electron.Connection, 1),
	}
	// Spawn off the server's main loop immediately
	// not exported
	if debug {
		debugsf = func(format string, data ...interface{}) { log.Printf(format, data...) }
	}

	return server
}

//Close connections it is exported so users can force close
func (as *AMQPSender) Close() {
	c := <-as.connections
	debugsf("Debug:close %s", c)
	c.Close(nil)
}

//Send  starts amqp server
func (as *AMQPSender) Send(jsonmsg string) {
	debugsf("Debug: AMQP send is invoked")
	//sentChan := make(chan electron.Outcome) // Channel to receive acknowledgements.
	go func(body string) {
		//defer wait.Done()
		// Wait for one goroutine per URL
		container := electron.NewContainer(fmt.Sprintf("send[%v]", os.Getpid()))
		//connections := make(chan electron.Connection, 1) // Connections to close on exit
		//log.Printf("PArsing URL %s\n", as.urlStr)
		url, err := amqp.ParseURL(as.urlStr)
		fatalsIf(err)
		c, err := container.Dial("tcp", url.Host) // NOTE: Dial takes just the Host part of the URL
		fatalsIf(err)
		as.connections <- c // Save connection so we can Close() when start() ends
		addr := strings.TrimPrefix(url.Path, "/")
		s, err := c.Sender(electron.Target(addr))
		fatalsIf(err)
		m := amqp.NewMessage()
		//body := fmt.Sprintf("%v%v", addr, jsonmsg)
		m.Marshal(body)
		debugsf("Debug:Sending alerts on a bus URL %s\n", body)
		// Note: can block if there is no space to buffer the message.
		s.SendForget(m)
		as.Close()
		//s.SendAsync(m, sentChan, body) // Outcome will be sent to sentChan
	}(jsonmsg)
	//outside go routin reciveve and processurlStr
	//var firstmsg=0
	/*for {
		out := <-sentChan // Outcome of async sends.
		if out.Error != nil {
			log.Fatalf("acknowledgement[%v] %v error: %v", jsonmsg, out.Value, out.Error)
		} else if out.Status != electron.Accepted {
			log.Fatalf("acknowledgement[%v] unexpected status: %v", jsonmsg, out.Status)
		} else {
			debugsf("acknowledgement[%v]  %v (%v)\n", jsonmsg, out.Value, out.Status)
		}
	}*/
	debugsf("Debug:Closing connections")
	//wait.Wait()

	//wait.Wait() // Wait for all goroutines to finish.

}

func fatalsIf(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
