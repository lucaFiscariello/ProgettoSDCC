/************************************************************************************
* Questo handler ha il compito di gestire le richieste di accesso alla sezione		*
* critica in accordo con l'algoritmo di lamport distribuito. In particolare viene	*
* implementato il clock logico e vengono previste funzionalità per:					*
*	- Invio richiesta accesso alla sezione critica;									*
*	- Invio ack in seguito alla ricezione di una richiesta;							*
*	- Attesa di tutti gli ack degli altri nodi;										*
*	- Rilascio. Invocata quando il nodo intende uscire dalla sezione critica.		*
*************************************************************************************/

package handler

import (
	"context"
	"encoding/json"
	Log "fiscariello/luca/node/Logger"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type RequestHandler struct {
	Url           string
	Topic_Request string
	ID_NODE       string
	HandlerNodes  *NodeActiveHandler
}

type MessageLamport struct {
	Type         string `json:"Type"`
	LogicalClock int    `json:"LogicalClock"`
	NodeID       string `json:"NodeID"`
}

var logicalClock = 0
var requestCriticSection []MessageLamport
var reader *kafka.Reader = nil

/*
 * Funzione usata per inviare una richiesta di accesso alla sezione critica. La richiesta è inviata su un canale di comunicazione
 * condiviso a tutti i nodi.
 */
func (rh *RequestHandler) SendRequest() {

	config := kafka.WriterConfig{
		Brokers: []string{rh.Url},
		Topic:   rh.Topic_Request}

	writer := kafka.NewWriter(config)

	logicalClock = logicalClock + 1
	request := MessageLamport{Type: RequestLamport, NodeID: rh.ID_NODE, LogicalClock: logicalClock}
	messageByte, _ := json.Marshal(request)
	writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	requestCriticSection = append(requestCriticSection, request)

	Log.Println("Il nodo corrente richiede di accedere alla sezione critica.")

}

/*
 * Funzione usata per inviare messaggi di ack dopo aver ricevuto una richiesta di accesso alla sezione critica.
 * Il messaggio di ack verrà inviato esclusivamente sul canale privato del nodo richiedente.
 */
func (rh *RequestHandler) SendAck(nodeID string) {

	logicalClock = logicalClock + 1

	config := kafka.WriterConfig{
		Brokers: []string{rh.Url},
		Topic:   nodeID}

	writer := kafka.NewWriter(config)

	request := MessageLamport{Type: ACKLamport, NodeID: rh.ID_NODE, LogicalClock: logicalClock}
	messageByte, _ := json.Marshal(request)
	writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

}

/*
 * Funzione usata per attendere tutti i messaggi di ack proveniente dagli altri nodi dopo aver inviato una richiesta di accesso.
 * Il metodo non implementa meccanismi particolari , se un nodo della rete non risponde il programma va in deadlock.
 */
func (rh *RequestHandler) WaitAllAck() {

	rh.singletonReader()

	for i := 0; i < rh.HandlerNodes.GetNumberActiveNode(); {
		var messageReceved MessageLamport

		messageKafka, err := reader.ReadMessage(context.Background())
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		if messageReceved.Type == ACKLamport {
			if messageReceved.LogicalClock > logicalClock {
				logicalClock = messageReceved.LogicalClock
			}

			i++
		}

	}

	Log.Println("Il nodo corrente ha ricevuto l'ack da tutti i nodi.")

}

/*
 * Funzione usata per avviare il listner che si mette in ascolto di nuove richieste provenienti da altri nodi
 * ma ascolta anche messaggi di release della sezione critica.
 * All'arrivo di una richiesta è inviato un messaggio di ack e la richiesta è inserita in una coda.
 * All'arrivo di un messaggio di rilascio la richiesta corrispondete è eliminata dalla coda.
 */
func (rh *RequestHandler) StartListnerRequest() {
	var messageReceved MessageLamport

	configRead := kafka.ReaderConfig{
		Brokers:  []string{rh.Url},
		Topic:    rh.Topic_Request,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configRead)

	// leggo messaggi in arrivo sul canale privato del nodo corrente.
	for {
		messageKafka, _ := reader.ReadMessage(context.Background())
		json.Unmarshal(messageKafka.Value, &messageReceved)

		//Potrebbero arrivare messaggi di richiesta di accesso alla sezione critica
		if messageReceved.Type == RequestLamport && messageReceved.NodeID != rh.ID_NODE {
			if messageReceved.LogicalClock > logicalClock {
				logicalClock = messageReceved.LogicalClock
			}

			requestCriticSection = append(requestCriticSection, messageReceved)
			rh.SendAck(messageReceved.NodeID)

			Log.Println("Il nodo corrente riceve una nuova richiesta: " + fmt.Sprint(messageReceved) + ". Attuamente le richieste in corso  sono: " + fmt.Sprint(requestCriticSection) + ". Ack inviato")
		}

		//Potrebbero arrivare messaggi di rilascio della sezione critica
		if messageReceved.Type == ReleaseLamport {

			positionRequest := -1
			found := false

			for i := range requestCriticSection {
				if requestCriticSection[i].NodeID == messageReceved.NodeID {
					positionRequest = i
					found = true
				}
			}

			//Elimino la richiesta di accesso appena soddisfatta dalla coda
			if found {
				requestCriticSection = append(requestCriticSection[:positionRequest], requestCriticSection[positionRequest+1:]...)
				Log.Println("Accesso alla sezione critica rilasciato dal nodo: " + messageReceved.NodeID + ". Altre rihieste ancora attive: " + fmt.Sprint(requestCriticSection))
			}
		}
	}

}

/*
 * Funzione usata per avvisare gli altri nodi della rete che il nodo corrente ha intenzione di uscire dalla sezione critica.
 */
func (rh *RequestHandler) Release() {

	logicalClock = logicalClock + 1

	config := kafka.WriterConfig{
		Brokers: []string{rh.Url},
		Topic:   rh.Topic_Request}

	writer := kafka.NewWriter(config)

	message := MessageLamport{Type: ReleaseLamport, NodeID: rh.ID_NODE, LogicalClock: logicalClock}
	messageByte, _ := json.Marshal(message)
	writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	positionRequest := -1
	found := false

	for i := range requestCriticSection {
		if requestCriticSection[i].NodeID == rh.ID_NODE {
			positionRequest = i
			found = true
		}
	}

	fmt.Println(requestCriticSection)

	if found {
		requestCriticSection = append(requestCriticSection[:positionRequest], requestCriticSection[positionRequest+1:]...)
		Log.Println("Il nodo corrente rilascia l'accesso alla sezione critica.")
	}

}

/*
 * Funzione usata per verificare se il nodo corrente rispetta le condizioni per poter accedere alla sezione critica.
 */
func (rh *RequestHandler) CanExecute() bool {

	for {
		minClock := requestCriticSection[0].LogicalClock
		minIDnode := requestCriticSection[0].NodeID

		for i := range requestCriticSection {

			if requestCriticSection[i].LogicalClock < minClock {

				//Per accedere alla sezione critica il clock logico della richiesta del nodo corrente deve essere il minimo.
				minClock = requestCriticSection[i].LogicalClock
				minIDnode = requestCriticSection[i].NodeID

			} else if requestCriticSection[i].LogicalClock == minClock && minIDnode > requestCriticSection[i].NodeID {

				//Se ci sono più nodi con lo stesso clock logico ha la precedenza quello con id minore
				minClock = requestCriticSection[i].LogicalClock
				minIDnode = requestCriticSection[i].NodeID

			}
		}

		if minIDnode == rh.ID_NODE {
			return true
		}

		time.Sleep(1 * time.Second)

	}

}

func (rh *RequestHandler) singletonReader() {

	if reader == nil {
		configRead := kafka.ReaderConfig{
			Brokers:  []string{rh.Url},
			Topic:    rh.ID_NODE,
			MaxBytes: 10e6}

		reader = kafka.NewReader(configRead)
	}

}
