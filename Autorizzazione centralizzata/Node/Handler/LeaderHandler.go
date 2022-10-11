package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type LeaderHandler struct {
	Url       string
	ID_LEADER string
}

var queueMessage []Message
var freeCriticSection bool

func (lh LeaderHandler) StartLeaderHandler() {

	freeCriticSection = true
	go lh.startListnerRelease()
	lh.startListnerAutorization()

}

func (lh LeaderHandler) startListnerAutorization() {
	configRead := kafka.ReaderConfig{
		Brokers:  []string{lh.Url},
		Topic:    lh.ID_LEADER,
		MaxBytes: 10e6}

	readerLeader := kafka.NewReader(configRead)

	for {

		for {
			var messageReceved Message

			messageKafka, err := readerLeader.ReadMessage(context.Background())
			err = json.Unmarshal(messageKafka.Value, &messageReceved)
			checkErr(err)

			if messageReceved.TypeMessage == RequestAutorization {
				fmt.Println("leader riceve autorizzazione: " + messageReceved.Id_node)
				queueMessage = append(queueMessage, messageReceved)
			}

		}
	}
}

func (lh LeaderHandler) startListnerRelease() {
	configRead := kafka.ReaderConfig{
		Brokers:  []string{lh.Url},
		Topic:    lh.ID_LEADER,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configRead)

	var messageReceved Message

	for {

		if messageReceved.TypeMessage == AutorizationRelease || freeCriticSection {

			for freeCriticSection && len(queueMessage) > 0 {
				oldMessage := queueMessage[0]
				queueMessage = queueMessage[1:]

				lastIdSend := oldMessage.Id_node
				lastNumberMessage := oldMessage.Id_message

				lh.sendAutorization(lastIdSend, lastNumberMessage)
				ack := lh.waitAck(lastIdSend, lastNumberMessage)
				freeCriticSection = !ack
			}
		}

		messageKafka, err := reader.ReadMessage(context.Background())
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

	}
}

func (lh LeaderHandler) sendAutorization(lastIdSend string, lastNumberMessage int) {

	configWrite := kafka.WriterConfig{
		Brokers: []string{lh.Url},
		Topic:   lastIdSend}

	writer := kafka.NewWriter(configWrite)

	message := Message{TypeMessage: AutorizationOK, Id_node: lh.ID_LEADER, Id_message: lastNumberMessage}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	fmt.Println("leader autorizza: " + lastIdSend)

	checkErr(err)
}

func (lh LeaderHandler) waitAck(lastIdSend string, lastNumberMessage int) bool {
	configRead := kafka.ReaderConfig{
		Brokers:  []string{lh.Url},
		Topic:    lh.ID_LEADER,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configRead)

	var messageReceved Message

	timeout := time.Now()
	for time.Since(timeout) < 3*time.Second {

		contextTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		messageKafka, _ := reader.ReadMessage(contextTimeout)
		json.Unmarshal(messageKafka.Value, &messageReceved)

		if messageReceved.TypeMessage == AutorizationACK && messageReceved.Id_message == lastNumberMessage && messageReceved.Id_message == lastID {
			return true
		}

	}

	return false
}
