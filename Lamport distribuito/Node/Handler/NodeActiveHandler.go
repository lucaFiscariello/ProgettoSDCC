package handler

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type NodeActiveHandler struct {
	Url                string
	ID_NODE            string
	PRESENTATION_TOPIC string
	ALL_NODE           map[string]bool
}

func (nh NodeActiveHandler) ListenNewNode() {
	config := kafka.ReaderConfig{
		Brokers:  []string{nh.Url},
		Topic:    nh.PRESENTATION_TOPIC,
		MaxBytes: 10e6}
	reader := kafka.NewReader(config)

	for {
		message, err := reader.ReadMessage(context.Background())
		checkErr(err)

		id_node := string(message.Value[:])

		if !nh.contains(id_node) {
			nh.ALL_NODE[id_node] = false
		}

	}
}

func (nh NodeActiveHandler) GetNode() map[string]bool {
	var copyAllNode map[string]bool = make(map[string]bool)
	for key, value := range nh.ALL_NODE {
		copyAllNode[key] = value
	}
	return copyAllNode
}

func (nh NodeActiveHandler) GetNumberNode() int {
	return len(nh.ALL_NODE)
}

func (nh NodeActiveHandler) GetNumberActiveNode() int {

	nodeActive := 0

	for _, isActive := range nh.ALL_NODE {
		if isActive {
			nodeActive++
		}
	}
	return nodeActive
}

func (nh NodeActiveHandler) GetAllNode() map[string]bool {
	allNodeNoTActive := make(map[string]bool)
	for key := range nh.ALL_NODE {
		allNodeNoTActive[key] = false
	}
	return allNodeNoTActive
}

func (nh NodeActiveHandler) SetNode(nodes map[string]bool) {

	for node, isActive := range nodes {
		nh.ALL_NODE[node] = isActive
	}

}

func (nh NodeActiveHandler) contains(id_search string) bool {
	if id_search == nh.ID_NODE {
		return true
	}

	for id := range nh.ALL_NODE {

		if id == id_search {
			return true
		}
	}
	return false
}
