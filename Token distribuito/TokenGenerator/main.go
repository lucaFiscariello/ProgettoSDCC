package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

var TIMEOUT = os.Getenv("TIMEOUT")
var ID_KAFKA = os.Getenv("ID_KAFKA")
var TOKENREQUEST_TOPIC = os.Getenv("TOKENREQUEST_TOPIC")

type Message struct {
	TypeMessage  string `json:"TypeMessage"`
	Id_node      string `json:"Id_node"`
	Id_message   int    `json:"Id_message"`
	ConcessToken bool   `json:"ConcessToken"`
}

const (
	messageType  string = "messageType"
	HeartBeat           = "HeartBeat"
	Token               = "Token"
	TokenRequest        = "TokenRequest"
	TokenCheck          = "TokenCheck"
)

func main() {

	waitStartNode()

	var messageReceved Message

	var timeout time.Time = time.Now()
	var TokenGenerated = false
	var concessToken = true

	configReader := kafka.ReaderConfig{
		Brokers:  []string{ID_KAFKA},
		Topic:    TOKENREQUEST_TOPIC,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configReader)

	for {
		messageKafka, err := reader.ReadMessage(context.Background())
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		if messageReceved.TypeMessage == TokenRequest {
			configWrite := kafka.WriterConfig{
				Brokers: []string{ID_KAFKA},
				Topic:   messageReceved.Id_node}

			writer := kafka.NewWriter(configWrite)

			if !TokenGenerated {
				TokenGenerated = true
				timeout = time.Now()
			} else if time.Since(timeout) > 15*time.Second {
				concessToken = true
			} else {
				concessToken = false
			}

			fmt.Println("Richiesta token da " + messageReceved.Id_node + ". Token concesso: " + strconv.FormatBool(concessToken))

			message := Message{TypeMessage: TokenRequest, ConcessToken: concessToken, Id_message: messageReceved.Id_message}
			messageByte, _ := json.Marshal(message)
			err = writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
			checkErr(err)
		} else {
			timeout = time.Now()
		}

	}

}

func waitStartNode() {
	timeout, err := strconv.Atoi(TIMEOUT)
	checkErr(err)
	time.Sleep(time.Duration(timeout) * time.Second)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
