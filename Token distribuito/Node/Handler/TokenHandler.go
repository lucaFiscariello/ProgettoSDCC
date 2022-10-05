package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"github.com/segmentio/kafka-go"
)

const (
	HaveToken string = "HaveToken"
	YES              = "YES"
	NO               = "NO"
)

type TokenHandler struct {
	Url                  string
	ID_NODE              string
	TOKEN_REQUEST_TOPIC  string
	TOKEN_CHECK_TOPIC    string
	TOKEN                bool
	ReaderTokenGenerator *kafka.Reader
	WriterTokenGenerator *kafka.Writer
}

type HandlerTK interface {
	SendToken()
	ListenReceveToken()
}

var Id_message = 0

func (h TokenHandler) SendToken(is_active map[string]bool) {

	idTemplate := h.ID_NODE
	idNumber := h.ID_NODE
	regexprTemplate := regexp.MustCompile(`[^a-zA-Z ]+`)
	idTemplatenew := regexprTemplate.ReplaceAllString(idTemplate, "")

	regexprNumber := regexp.MustCompile("[^0-9]+")
	idNumbernew := regexprNumber.ReplaceAllString(idNumber, "")

	id, err := strconv.Atoi(idNumbernew)
	checkErr(err)

	numberNode := len(is_active) + 1

	for i := 1; i < 2*numberNode; i++ {
		numId := (i + id) % numberNode
		idToSend := idTemplatenew + fmt.Sprint(numId)

		if is_active[idToSend] {
			configWrite := kafka.WriterConfig{
				Brokers: []string{h.Url},
				Topic:   idToSend}

			fmt.Println(is_active)
			fmt.Println("send token to: " + idToSend)
			writer := kafka.NewWriter(configWrite)

			message := Message{TypeMessage: Token, Id_node: h.ID_NODE}
			messageByte, _ := json.Marshal(message)
			err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
			checkErr(err)
			return
		}
	}
}

func (h TokenHandler) ListenReceveToken(channel chan bool) {
	configRead := kafka.ReaderConfig{
		Brokers:  []string{h.Url},
		Topic:    h.ID_NODE,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configRead)

	for {
		var messageReceved Message

		messageKafka, err := reader.ReadMessage(context.Background())
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		if messageReceved.TypeMessage == Token {
			channel <- true
		}
	}

}

func (h TokenHandler) RequestToken() bool {

	if h.TOKEN {
		return true
	}

	var messageReceved Message
	Id_message = Id_message + 1

	message := Message{TypeMessage: TokenRequest, Id_node: h.ID_NODE, Id_message: Id_message}
	messageByte, _ := json.Marshal(message)
	err := h.WriterTokenGenerator.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

	for {
		messageKafka, err := h.ReaderTokenGenerator.ReadMessage(context.Background())
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		if messageReceved.TypeMessage == TokenRequest && messageReceved.Id_message == Id_message {
			h.TOKEN = messageReceved.ConcessToken
			return messageReceved.ConcessToken
		}
	}

}

func (h TokenHandler) SendHackToken() {

	message := Message{TypeMessage: TokenCheck, Id_node: h.ID_NODE}
	messageByte, _ := json.Marshal(message)
	err := h.WriterTokenGenerator.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

}
