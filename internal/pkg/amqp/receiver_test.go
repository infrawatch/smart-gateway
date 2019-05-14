package amqp10

import (
	"encoding/json"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TstMessage struct {
	message string `json:"message"`
}

func TestPut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	url := "amqp://127.0.0.1:5672/multicast"
	outgoing := TstMessage{"communication test"}
	result := make(chan TstMessage)
	go func() {
		var incomming TstMessage
		done := make(chan bool)
		amqpServer := NewAMQPServer(url, true, 1, 0, nil, done, false, "metrics-test")
		data := <-amqpServer.notifier
		err := json.Unmarshal([]byte(data), &incomming)
		log.Print("received testing message")
		if err != nil {
			t.Errorf("Failed to parse testing message from QDR: %s", err)
		}
		result <- incomming
	}()

	log.Print("sending testing message through to qdr")
	sender := NewAMQPSender(url, true)
	body, err := json.Marshal(outgoing)
	if err != nil {
		t.Errorf("Failed to send testing message to QDR: %s", err)
	}
	sender.Send(string(body))
	sender.Close()

	incomming := <-result
	assert.Equal(t, outgoing.message, incomming.message)
}
