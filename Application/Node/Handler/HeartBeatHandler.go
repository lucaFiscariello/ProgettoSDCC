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
}

type Message struct {
	TypeMessage string `json:"TypeMessage"`
	Id_node     string `json:"Id_node"`
}

const (
	messageType string = "messageType"
	HeartBeat          = "HeartBeat"
	Token              = "Token"
)

type HandlerHB interface {
	FetchResponseBeat(reader *kafka.Reader, channel chan string)
	SendHeart(writer *kafka.Writer)
	Wait(channel chan string, ALL_NODE map[string]bool)
}

func (h HeartBeatHandler) SendHeart(writer *kafka.Writer) {
	messageHeart := h.ID_NODE
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: []byte(messageHeart)})
	checkErr(err)
}

func (h HeartBeatHandler) Wait(reader *kafka.Reader, ALL_NODE map[string]bool) {

	var messageReceved Message
	var idNodebeat string

	totalResponse := 0
	totalNode := 0

	for id := range ALL_NODE {
		ALL_NODE[id] = false
		totalNode++
	}

	timeout := time.Now()
	for totalResponse < totalNode && time.Since(timeout) < 5*time.Second {

		contextTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		messageKafka, err := reader.ReadMessage(contextTimeout)
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		idNodebeat = string(messageReceved.Id_node)

		if contains(ALL_NODE, idNodebeat) {
			totalResponse++
			ALL_NODE[idNodebeat] = true
		}

	}

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
