package handler

import (
	"context"
	"encoding/json"
	Log "fiscariello/luca/node/Logger"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type RequestHandler struct {
	Url           string
	Topic_Request string
	ID_NODE       string
	HandlerNodes  *NodeActiveHandler
}

type MessageLamport struct {
	Type         string `json:"Type"`
	LogicalClock int    `json:"LogicalClock"`
	NodeID       string `json:"NodeID"`
}

var logicalClock = 0
var requestCriticSection []MessageLamport
var reader *kafka.Reader = nil

func (rh *RequestHandler) SendRequest() {

	config := kafka.WriterConfig{
		Brokers: []string{rh.Url},
		Topic:   rh.Topic_Request}

	writer := kafka.NewWriter(config)

	logicalClock = logicalClock + 1
	request := MessageLamport{Type: RequestLamport, NodeID: rh.ID_NODE, LogicalClock: logicalClock}
	messageByte, _ := json.Marshal(request)
	writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	requestCriticSection = append(requestCriticSection, request)

	Log.Println("Il nodo corrente richiede di accedere alla sezione critica.")

}

func (rh *RequestHandler) WaitAllAck() {

	rh.singletonReader()

	for i := 0; i < rh.HandlerNodes.GetNumberActiveNode(); {
		var messageReceved MessageLamport

		messageKafka, err := reader.ReadMessage(context.Background())
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		if messageReceved.Type == ACKLamport {
			if messageReceved.LogicalClock > logicalClock {
				logicalClock = messageReceved.LogicalClock
			}

			i++
		}

	}

	Log.Println("Il nodo corrente ha ricevuto l'ack da tutti i nodi.")

}

func (rh *RequestHandler) StartListnerRequest() {
	var messageReceved MessageLamport

	configRead := kafka.ReaderConfig{
		Brokers:  []string{rh.Url},
		Topic:    rh.Topic_Request,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configRead)

	for {
		messageKafka, _ := reader.ReadMessage(context.Background())
		json.Unmarshal(messageKafka.Value, &messageReceved)

		if messageReceved.Type == RequestLamport && messageReceved.NodeID != rh.ID_NODE {
			if messageReceved.LogicalClock > logicalClock {
				logicalClock = messageReceved.LogicalClock
			}

			requestCriticSection = append(requestCriticSection, messageReceved)
			rh.SendAck(messageReceved.NodeID)

			Log.Println("Il nodo corrente riceve una nuova richiesta: " + fmt.Sprint(messageReceved) + ". Attuamente le richieste in corso  sono: " + fmt.Sprint(requestCriticSection) + ". Ack inviato")
		}

		if messageReceved.Type == ReleaseLamport {

			positionRequest := -1
			found := false

			for i := range requestCriticSection {
				if requestCriticSection[i].NodeID == messageReceved.NodeID {
					positionRequest = i
					found = true
				}
			}

			if found {
				requestCriticSection = append(requestCriticSection[:positionRequest], requestCriticSection[positionRequest+1:]...)
				Log.Println("Accesso alla sezione critica rilasciato dal nodo: " + messageReceved.NodeID + ". Altre rihieste ancora attive: " + fmt.Sprint(requestCriticSection))
			}
		}
	}

}

func (rh *RequestHandler) SendAck(nodeID string) {

	logicalClock = logicalClock + 1

	config := kafka.WriterConfig{
		Brokers: []string{rh.Url},
		Topic:   nodeID}

	writer := kafka.NewWriter(config)

	request := MessageLamport{Type: ACKLamport, NodeID: rh.ID_NODE, LogicalClock: logicalClock}
	messageByte, _ := json.Marshal(request)
	writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

}

func (rh *RequestHandler) Release() {

	logicalClock = logicalClock + 1

	config := kafka.WriterConfig{
		Brokers: []string{rh.Url},
		Topic:   rh.Topic_Request}

	writer := kafka.NewWriter(config)

	message := MessageLamport{Type: ReleaseLamport, NodeID: rh.ID_NODE, LogicalClock: logicalClock}
	messageByte, _ := json.Marshal(message)
	writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	positionRequest := -1
	found := false

	for i := range requestCriticSection {
		if requestCriticSection[i].NodeID == rh.ID_NODE {
			positionRequest = i
			found = true
		}
	}

	fmt.Println(requestCriticSection)

	if found {
		requestCriticSection = append(requestCriticSection[:positionRequest], requestCriticSection[positionRequest+1:]...)
		Log.Println("Il nodo corrente rilascia l'accesso alla sezione critica.")
	}

}

func (rh *RequestHandler) CanExecute() bool {

	for {
		minClock := requestCriticSection[0].LogicalClock
		minIDnode := requestCriticSection[0].NodeID

		for i := range requestCriticSection {
			if requestCriticSection[i].LogicalClock < minClock {

				minClock = requestCriticSection[i].LogicalClock
				minIDnode = requestCriticSection[i].NodeID

			} else if requestCriticSection[i].LogicalClock == minClock && minIDnode > requestCriticSection[i].NodeID {

				minClock = requestCriticSection[i].LogicalClock
				minIDnode = requestCriticSection[i].NodeID

			}
		}

		if minIDnode == rh.ID_NODE {
			return true
		}

		time.Sleep(1 * time.Second)

	}

}

func (rh *RequestHandler) singletonReader() {

	if reader == nil {
		configRead := kafka.ReaderConfig{
			Brokers:  []string{rh.Url},
			Topic:    rh.ID_NODE,
			MaxBytes: 10e6}

		reader = kafka.NewReader(configRead)
	}

}
