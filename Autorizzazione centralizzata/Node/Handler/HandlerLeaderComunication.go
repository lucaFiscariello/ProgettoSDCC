package handler

import (
	"context"
	"encoding/json"
	"time"

	Log "fiscariello/luca/node/Logger"

	"github.com/segmentio/kafka-go"
)

type LeaderComunicationHandler struct {
	Url       string
	ID_LEADER string
	ID_NODE   string
}

var lastID = 0
var reader *kafka.Reader = nil

func (lh *LeaderComunicationHandler) CanExecute() bool {

	lh.singletonReader()

	config := kafka.WriterConfig{
		Brokers: []string{lh.Url},
		Topic:   lh.ID_LEADER}

	writer := kafka.NewWriter(config)

	lastID = lastID + 1

	message := Message{TypeMessage: RequestAutorization, Id_node: lh.ID_NODE, Id_message: lastID}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

	Log.Println("Il node corrente richiede di entrare in sezione critica.")

	contextTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		var messageReceved Message

		messageKafka, err := reader.ReadMessage(contextTimeout)
		err = json.Unmarshal(messageKafka.Value, &messageReceved)

		if err != nil {
			Log.Println("Il leader non è contattabile.")
			return false
		}

		if messageReceved.TypeMessage == AutorizationOK && messageReceved.Id_message == lastID {
			lh.SendAck(writer)

			Log.Println("Il nodo corrente ottiene autorizzazione dal leader.")
			return true
		}

	}
}

func (lh *LeaderComunicationHandler) SendAck(writer *kafka.Writer) {
	message := Message{TypeMessage: AutorizationACK, Id_node: lh.ID_NODE, Id_message: lastID}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

	Log.Println("Il nodo corrente invia un ack al leader.")

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

	Log.Println("Il nodo corrente rilascia l'autorizzazione")

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
