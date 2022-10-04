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
	Url         string
	ID_NODE     string
	TOKEN_TOPIC string
	TOKEN       string
}

type HandlerTK interface {
	SendToken()
	ListenReceveToken()
}

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

	for i := 1; ; i++ {
		numId := (i + id) % numberNode
		idToSend := idTemplatenew + fmt.Sprint(numId)

		if is_active[idToSend] {
			configWrite := kafka.WriterConfig{
				Brokers: []string{h.Url},
				Topic:   idToSend}

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

func (h TokenHandler) ListenReceveToken(channel chan string) {
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
			channel <- YES
		}
	}

}
