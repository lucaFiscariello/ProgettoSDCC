package handler

import (
	"context"
	"encoding/json"
	Log "fiscariello/luca/node/Logger"
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
				queueMessage = append(queueMessage, messageReceved)
				Log.Println("Il leader riceve richiesta autorizzazione da: " + messageReceved.Id_node + ". Richiesta: " + fmt.Sprint(messageReceved))
				Log.Println("Attualmente la coda delle richieste contiene: " + fmt.Sprint(queueMessage))

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

		freeCriticSection = true

		if messageReceved.TypeMessage == AutorizationRelease || freeCriticSection {

			if !freeCriticSection {
				Log.Println("E' stato rilasciato l'accesso alla sezione critica da parte di: " + messageReceved.Id_node)
				Log.Println("Attualmente la coda delle richieste contiene: " + fmt.Sprint(queueMessage))
			}

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

	Log.Println("Il leader autorizza il nodo " + lastIdSend + " ad accedere alla sezione critica")
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

		if messageReceved.TypeMessage == AutorizationACK && messageReceved.Id_message == lastNumberMessage && messageReceved.Id_node == lastIdSend {
			Log.Println("Il leader riceve l'ack dal nodo: " + lastIdSend)
			return true
		}

	}

	Log.Println("Il leader NON riceve l'ack dal nodo: " + lastIdSend)
	return false
}
