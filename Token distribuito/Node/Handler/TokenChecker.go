package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type TokenChecker struct {
	Url       string
	ID_NODE   string
	WAIT_TIME string
}

func (tc TokenChecker) Start() {

	var messageReceved Message

	var timeout time.Time = time.Now()
	var TokenGenerated = false
	var concessToken = true

	configReader := kafka.ReaderConfig{
		Brokers:  []string{tc.Url},
		Topic:    tc.ID_NODE,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configReader)

	for {
		messageKafka, err := reader.ReadMessage(context.Background())
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		if messageReceved.TypeMessage == TokenRequest {
			configWrite := kafka.WriterConfig{
				Brokers: []string{tc.Url},
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

			message := Message{TypeMessage: TokenResponse, ConcessToken: concessToken, Id_message: messageReceved.Id_message}
			messageByte, _ := json.Marshal(message)
			err = writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
			checkErr(err)
		} else if messageReceved.TypeMessage == TokenCheck {
			timeout = time.Now()
			fmt.Println("check ricevuto da: " + messageReceved.Id_node)
		}

	}

}
