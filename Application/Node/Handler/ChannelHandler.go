package handler

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
)

type ChanneltHandler struct {
	Url             string
	ID_NODE         string
	HEARTBEAT_TOPIC string
}

type HandlerCH interface {
	SendBeat()
}

func (ch ChanneltHandler) SendBeat() {

	configReadNode := kafka.ReaderConfig{
		Brokers:  []string{ch.Url},
		Topic:    ch.HEARTBEAT_TOPIC,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configReadNode)

	for {
		message, err := reader.ReadMessage(context.Background())
		idNodeToRespons := string(message.Value[:])

		if idNodeToRespons != ch.ID_NODE {

			configWrite := kafka.WriterConfig{
				Brokers: []string{ch.Url},
				Topic:   idNodeToRespons}

			writer := kafka.NewWriter(configWrite)

			message := Message{TypeMessage: HeartBeat, Id_node: ch.ID_NODE}
			messageByte, _ := json.Marshal(message)
			err = writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
			checkErr(err)
		}
	}

}
