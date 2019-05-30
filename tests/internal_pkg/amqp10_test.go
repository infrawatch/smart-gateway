package tests

import (
	"fmt"
	"testing"

	"github.com/redhat-service-assurance/smart-gateway/internal/pkg/amqp10"
	"github.com/stretchr/testify/assert"
)

const QDR_URL = "amqp://127.0.0.1:5672/collectd/telemetry"
const QDR_MSG = "{\"message\": \"smart gateway test\"}"

func TestSendAndReceiveMessage(t *testing.T) {
	sender := amqp10.NewAMQPSender(QDR_URL, true)
	receiver := amqp10.NewAMQPServer(QDR_URL, true, 1, 0, nil, "metrics-test")
	ackChan := sender.GetAckChannel()
	t.Run("Test receive", func(t *testing.T) {
		t.Parallel()
		data := <-receiver.GetNotifier()
		assert.Equal(t, QDR_MSG, data)
		fmt.Printf("Finished send")
	})
	t.Run("Test send and ACK", func(t *testing.T) {
		t.Parallel()
		sender.Send(QDR_MSG)
		// otherwise receiver blocks
		assert.Equal(t, 1, <-receiver.GetStatus())
		assert.Equal(t, true, <-receiver.GetDoneChan())
		outcome := <-ackChan
		assert.Equal(t, "smart-gateway-ack", outcome.Value.(string))
	})
}
