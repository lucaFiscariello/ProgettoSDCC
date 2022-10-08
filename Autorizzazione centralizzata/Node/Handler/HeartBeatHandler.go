package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type HeartBeatHandler struct {
	ID_NODE string
	Url     string
}

type Message struct {
	TypeMessage string `json:"TypeMessage"`
	Id_node     string `json:"Id_node"`
	Id_message  int    `json:"Id_message"`
}

const (
	messageType         string = "messageType"
	HeartBeat                  = "HeartBeat"
	Token                      = "Token"
	TokenRequest               = "TokenRequest"
	TokenCheck                 = "TokenCheck"
	RequestAutorization        = "RequestAutorization"
	AutorizationOK             = "AutorizationOK"
	AutorizationRelease        = "AutorizationRelease"
	AutorizationACK            = "AutorizationACK"
)

type HandlerHB interface {
	SendBeat(reader *kafka.Reader)
	SendHeart(writer *kafka.Writer)
	Wait(channel chan string, ALL_NODE map[string]bool)
}

var last_id_message int = 0

func (h HeartBeatHandler) SendBeat(reader *kafka.Reader) {

	for {
		var message Message

		messageKafka, err := reader.ReadMessage(context.Background())
		checkErr(err)

		err = json.Unmarshal(messageKafka.Value, &message)
		idNodeToRespons := message.Id_node
		idMessageToRespons := message.Id_message

		if idNodeToRespons != h.ID_NODE {

			configWrite := kafka.WriterConfig{
				Brokers: []string{h.Url},
				Topic:   idNodeToRespons}

			writer := kafka.NewWriter(configWrite)

			message := Message{TypeMessage: HeartBeat, Id_node: h.ID_NODE, Id_message: idMessageToRespons}
			messageByte, _ := json.Marshal(message)
			err = writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
		}
	}
}

func (h HeartBeatHandler) SendHeart(writer *kafka.Writer) {
	last_id_message = last_id_message + 1
	messageHeart := Message{TypeMessage: HeartBeat, Id_node: h.ID_NODE, Id_message: last_id_message}
	messageByte, err := json.Marshal(messageHeart)
	err = writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)
}

func (h HeartBeatHandler) Wait(reader *kafka.Reader, handlerNode *NodeActiveHandler) {

	var messageReceved Message
	var idNodebeat string

	totalResponse := 0
	totalNode := handlerNode.GetNumberNode()
	nodeActive := handlerNode.GetAllNode()

	timeout := time.Now()
	for totalResponse < totalNode && time.Since(timeout) < 5*time.Second {

		contextTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		messageKafka, _ := reader.ReadMessage(contextTimeout)
		json.Unmarshal(messageKafka.Value, &messageReceved)

		if messageReceved.TypeMessage == HeartBeat && messageReceved.Id_message == last_id_message {
			idNodebeat = string(messageReceved.Id_node)
			nodeActive[idNodebeat] = true
			totalResponse++

		}

	}

	handlerNode.SetNode(nodeActive)

}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}

}

func contains(allNode map[string]bool, id_search string) bool {

	for id := range allNode {

		if id == id_search {
			return true
		}
	}
	return false
}
