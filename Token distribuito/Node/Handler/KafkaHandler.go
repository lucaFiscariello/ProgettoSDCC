/*************************************************************************************
*	Questo handler ha il compito di gestire alcune interazioni tra il nodo corrente
*	e il broker del cluster kafka. In particolare vengono modellate due funzioni che permettono:
*		- la creazione di un nuovo topic
*		- l'invio di un messaggio di presentazione su un topic comune a tutti i nodi
**************************************************************************************/

package handler

import (
	"context"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type KafkatHandler struct {
	Url     string
	ID_NODE string
}

type HandlerKF interface {
	CreateNewTopicKafka()
	CreatePresentation()
}

func (kh KafkatHandler) CreateNewTopicKafka() {

	conn, err := kafka.Dial("tcp", kh.Url)
	checkErr(err)
	defer conn.Close()

	controller, err := conn.Controller()
	checkErr(err)

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	checkErr(err)
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{{Topic: kh.ID_NODE, NumPartitions: 1, ReplicationFactor: 1}}

	err = controllerConn.CreateTopics(topicConfigs...)
	checkErr(err)

}

func (kh KafkatHandler) CreatePresentation(PRESENTATION_TOPIC string) {

	config := kafka.WriterConfig{
		Brokers: []string{kh.Url},
		Topic:   PRESENTATION_TOPIC}

	writer := kafka.NewWriter(config)
	message := kh.ID_NODE
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: []byte(message)})
	checkErr(err)
}
