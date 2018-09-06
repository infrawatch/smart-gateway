package amqp10

import (
	"testing"
)

func TestPut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	var url = "amqp://10.19.110.5:5672/collectd/telemetry"
	var amqpServer *AMQPServer

	done := make(chan bool)

	amqpServer = NewAMQPServer(url, true, 10, 0, nil, done, false, "metrics-test")

	for i := 0; i < 10; i++ {
		data := <-amqpServer.notifier
		t.Logf("%s\n", data)
	}

}
