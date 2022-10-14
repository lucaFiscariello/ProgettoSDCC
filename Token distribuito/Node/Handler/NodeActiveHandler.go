/*************************************************************************************
*	Questo handler ha il compito di memorizzare quali sono i nodi della rete attualmente
* 	attivi. Queste informazioni sono memorizzate in una mappa.
*	Questo handler verrà sfruttato prevalentemente dall'handler dell'heartbeat
* 	il quale dopo aver completato un ciclo di heart-beat aggiornerà i dati nella mappa.
**************************************************************************************/

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

/*
 * Questa funzione attiva un listner che si mette in ascolto di messaggi di presentazione di altri nodi.
 * Infatti ogni qual volta un nuovo nodo accede alla rete, pubblica un nuovo messaggio su un topic comune specificando il
 * proprio id.
 */
func (nh NodeActiveHandler) StartListnerNewNode() {
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

/*
 * Questa funzione restituisce una copia della mappa dei nodi attivi
 */
func (nh NodeActiveHandler) GetNode() map[string]bool {
	var copyAllNode map[string]bool = make(map[string]bool)
	for key, value := range nh.ALL_NODE {
		copyAllNode[key] = value
	}
	return copyAllNode
}

/*
 * Questa funzione restituisce una copia della mappa dei nodi inizializzati tutti come non attivi
 */
func (nh NodeActiveHandler) GetAllNode() map[string]bool {
	allNodeNoTActive := make(map[string]bool)
	for key := range nh.ALL_NODE {
		allNodeNoTActive[key] = false
	}
	return allNodeNoTActive
}

func (nh NodeActiveHandler) GetNumberNode() int {
	return len(nh.ALL_NODE)
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
