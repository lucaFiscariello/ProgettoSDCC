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

type HandlerHB interface {
	SendBeat(reader *kafka.Reader)
	SendHeart(writer *kafka.Writer)
	Wait(reader *kafka.Reader, handlerNode *NodeActiveHandler)
	StartHeartBeatHandler(HEARTBEAT_TOPIC string, handlerNode *NodeActiveHandler)
}

var last_id_message int = 0

func (h HeartBeatHandler) SendBeat(reader *kafka.Reader) {

	var message Message

	for {

		//Leggo messaggi di heart dal topic comune
		messageKafka, err := reader.ReadMessage(context.Background())
		json.Unmarshal(messageKafka.Value, &message)
		checkErr(err)

		idNodeToRespons := message.Id_node
		idMessageToRespons := message.Id_message

		if idNodeToRespons != h.ID_NODE {

			configWrite := kafka.WriterConfig{
				Brokers: []string{h.Url},
				Topic:   idNodeToRespons}

			writer := kafka.NewWriter(configWrite)

			//Rispondo al nodo che ha inviato un messaggio di Heart con un Beat
			message := Message{TypeMessage: HeartBeat, Id_node: h.ID_NODE, Id_message: idMessageToRespons}
			messageByte, _ := json.Marshal(message)
			err = writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
		}
	}
}

func (h HeartBeatHandler) SendHeart(writer *kafka.Writer) {
	last_id_message = last_id_message + 1

	//invio messaggio di heart a tutti i nodi tramite topic comune
	messageHeart := Message{TypeMessage: HeartBeat, Id_node: h.ID_NODE, Id_message: last_id_message}
	messageByte, err := json.Marshal(messageHeart)
	err = writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)
}

func (h HeartBeatHandler) Wait(reader *kafka.Reader, handlerNode *NodeActiveHandler) {

	var messageReceved Message

	totalResponse := 0
	totalNode := handlerNode.GetNumberNode()
	nodeActive := handlerNode.GetAllNode()

	//rimango in attesa delle risposte di Beat fino allo scadere di un timeout
	timeout := time.Now()
	for totalResponse < totalNode && time.Since(timeout) < 5*time.Second {

		contextTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//leggo Beat
		messageKafka, _ := reader.ReadMessage(contextTimeout)
		json.Unmarshal(messageKafka.Value, &messageReceved)

		if messageReceved.TypeMessage == HeartBeat && messageReceved.Id_message == last_id_message {
			nodeActive[messageReceved.Id_node] = true
			totalResponse++

		}

	}

	handlerNode.SetNode(nodeActive)

}

func (handler *HeartBeatHandler) StartHeartBeatHandler(HEARTBEAT_TOPIC string, handlerNode *NodeActiveHandler) {
	configReadNodeBeat := kafka.ReaderConfig{
		Brokers:  []string{handler.Url},
		Topic:    handler.ID_NODE,
		MaxBytes: 10e6}

	configReadNodeHeart := kafka.ReaderConfig{
		Brokers:  []string{handler.Url},
		Topic:    HEARTBEAT_TOPIC,
		MaxBytes: 10e6}

	configWriteHeart := kafka.WriterConfig{
		Brokers: []string{handler.Url},
		Topic:   HEARTBEAT_TOPIC}

	readerBeat := kafka.NewReader(configReadNodeBeat)   //Reader che legge messaggi di beat sul canale "privato" del nodo
	readerHeart := kafka.NewReader(configReadNodeHeart) // Reader che legge messagi di heart sul canale condiviso tra tutti i nodi
	writerHeart := kafka.NewWriter(configWriteHeart)    // Writer che pubblica messaggi di heart sul canale condiviso

	//go routine che si mette in ascolto dei messaggi di heart e risponde con beat
	go handler.SendBeat(readerHeart)

	for {

		//Periodicamente invio messaggi di heart e attendo che arrivino tutti i messaggi di beat.
		handler.SendHeart(writerHeart)
		handler.Wait(readerBeat, handlerNode)

		time.Sleep(5 * time.Second)

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
