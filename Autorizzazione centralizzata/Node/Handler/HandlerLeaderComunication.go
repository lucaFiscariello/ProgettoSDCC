package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type LeaderComunicationHandler struct {
	Url          string
	ID_LEADER    string
	ID_NODE      string
	ReaderLeader *kafka.Reader
}

var lastID = 0
var reader *kafka.Reader = nil

func (lh *LeaderComunicationHandler) CanExecute() bool {

	config := kafka.WriterConfig{
		Brokers: []string{lh.Url},
		Topic:   lh.ID_LEADER}

	writer := kafka.NewWriter(config)

	lastID = lastID + 1

	message := Message{TypeMessage: RequestAutorization, Id_node: lh.ID_NODE, Id_message: lastID}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

	fmt.Println("Richiesta autorizzazione inviata")

	contextTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		var messageReceved Message

		messageKafka, err := lh.ReaderLeader.ReadMessage(contextTimeout)
		err = json.Unmarshal(messageKafka.Value, &messageReceved)

		if err != nil {
			return false
		}

		if messageReceved.TypeMessage == AutorizationOK && messageReceved.Id_message == lastID {
			lh.SendAck(writer)
			fmt.Println("Autorizzazione ottenuta")
			return true
		}

	}
}

func (lh *LeaderComunicationHandler) SendAck(writer *kafka.Writer) {
	message := Message{TypeMessage: AutorizationACK, Id_node: lh.ID_NODE, Id_message: lastID}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)
}

func (lh *LeaderComunicationHandler) Release() {

	config := kafka.WriterConfig{
		Brokers: []string{lh.Url},
		Topic:   lh.ID_LEADER}

	writer := kafka.NewWriter(config)

	lastID = lastID + 1

	message := Message{TypeMessage: AutorizationRelease, Id_node: lh.ID_NODE, Id_message: lastID}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

	fmt.Println("Autorizzazione rilasciata")

}

func (lh *LeaderComunicationHandler) singletonReader() {

	if reader == nil {
		configRead := kafka.ReaderConfig{
			Brokers:  []string{lh.Url},
			Topic:    lh.ID_NODE,
			MaxBytes: 10e6}

		reader = kafka.NewReader(configRead)
	}

}
