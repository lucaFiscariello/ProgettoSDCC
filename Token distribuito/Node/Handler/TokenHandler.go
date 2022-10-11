package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	HaveToken string = "HaveToken"
	YES              = "YES"
	NO               = "NO"
)

type TokenHandler struct {
	Url     string
	ID_NODE string
	TOKEN   bool
}

type HandlerTK interface {
	SendToken()
	ListenReceveToken()
}

var Id_message = 0
var reader *kafka.Reader = nil

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
			writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
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

		messageKafka, _ := reader.ReadMessage(context.Background())
		json.Unmarshal(messageKafka.Value, &messageReceved)

		if messageReceved.TypeMessage == Token {
			channel <- true
		}
	}

}

func (h TokenHandler) RequestToken(leaderID string) bool {

	configWrite := kafka.WriterConfig{
		Brokers: []string{h.Url},
		Topic:   leaderID}

	writerLeader := kafka.NewWriter(configWrite)

	if h.TOKEN {
		return true
	}

	var messageReceved Message
	Id_message = Id_message + 1

	fmt.Println("invio richiesta token se leader crush")
	message := Message{TypeMessage: TokenRequest, Id_node: h.ID_NODE, Id_message: Id_message}
	messageByte, _ := json.Marshal(message)
	writerLeader.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	h.singletonReader()
	for {
		contextTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		messageKafka, err := reader.ReadMessage(contextTimeout)
		err = json.Unmarshal(messageKafka.Value, &messageReceved)

		defer cancel()

		if err != nil {
			fmt.Println("attesa risposta leader timeout")
			return false
		}

		if messageReceved.TypeMessage == TokenResponse && messageReceved.Id_message == Id_message {
			fmt.Println("leader risponde richiesta token: " + strconv.FormatBool(messageReceved.ConcessToken))
			h.TOKEN = messageReceved.ConcessToken
			return messageReceved.ConcessToken
		}
	}

}

func (h TokenHandler) SendHackToken(leaderID string) {

	configWrite := kafka.WriterConfig{
		Brokers: []string{h.Url},
		Topic:   leaderID}

	writerLeader := kafka.NewWriter(configWrite)

	message := Message{TypeMessage: TokenCheck, Id_node: h.ID_NODE}
	messageByte, _ := json.Marshal(message)
	writerLeader.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	fmt.Println("send ack to: " + leaderID)

}

func (h TokenHandler) singletonReader() {

	if reader == nil {
		configRead := kafka.ReaderConfig{
			Brokers:  []string{h.Url},
			Topic:    h.ID_NODE,
			MaxBytes: 10e6}

		reader = kafka.NewReader(configRead)
	}

}
