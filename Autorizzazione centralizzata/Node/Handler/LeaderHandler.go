/****************************************************************************************
* Questo handler verrà utilizzato dal nodo della rete eletto leader che dovrà gestire	*
* l'accesso alla sezione critica tramite autorizzazioni. L'handler implementa due		*
* funzionalità per:																		*
* 	- mettersi in ascolti di messaggi di richiesta di autorizzazione degli altri nodi	*
* 	- mettersi in ascolti di messaggi di release										*
* Il leader prima di rilasciare l'autorizzazione si assicura che il nodo richiedente	*
* sia ancora attivo attendendo un suo ack. Allo scadere di un timeout se il nodo che	*
* aveva l'accesso alla sezione critica non rilascia l'accesso gli viene sottratto		*
* in quanto considerato non attivo.														*
*****************************************************************************************/
package handler

import (
	"context"
	"encoding/json"
	Log "fiscariello/luca/node/Logger"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type LeaderHandler struct {
	Url       string
	ID_LEADER string
}

var queueMessage []Message
var freeCriticSection bool

/*
 * Questa funzione gestisce iterativamente le richieste di accesso degli altri nodi. Ad ogni ciclo il leader:
 * 	- seleziona una richiesta dalla coda
 * 	- invia autorizzazione al nodo richiedente
 *	- attende suo ack
 *	- attende rilascio
 */
func (lh LeaderHandler) StartLeaderHandler() {

	go lh.startListnerAutorization()

	for {

		for len(queueMessage) > 0 {
			oldMessage := queueMessage[0]
			queueMessage = queueMessage[1:]

			lastIdSend := oldMessage.Id_node
			lastNumberMessage := oldMessage.Id_message

			lh.sendAutorization(lastIdSend, lastNumberMessage)
			ack := lh.waitAck(lastIdSend, lastNumberMessage)

			if ack {
				lh.waitRelease(lastIdSend, lastNumberMessage)
			}

		}

	}

}

/*
 * Questa funzione avvia un listner che si mette in ascolto di richieste di autorizzazioni
 * che vengono memorizzate in una coda.
 */
func (lh LeaderHandler) startListnerAutorization() {
	configRead := kafka.ReaderConfig{
		Brokers:  []string{lh.Url},
		Topic:    lh.ID_LEADER,
		MaxBytes: 10e6}

	readerLeader := kafka.NewReader(configRead)

	for {
		var messageReceved Message

		messageKafka, err := readerLeader.ReadMessage(context.Background())
		err = json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		if messageReceved.TypeMessage == RequestAutorization {
			queueMessage = append(queueMessage, messageReceved)
			Log.Println("Il leader riceve richiesta autorizzazione da: " + messageReceved.Id_node + ". Richiesta: " + fmt.Sprint(messageReceved))
			Log.Println("Attualmente la coda delle richieste contiene: " + fmt.Sprint(queueMessage))

		}

	}

}

/*
 * Con questa funzione il leader concede l'autorizzazione a un nodo ad entrare nella sezione critica
 */
func (lh LeaderHandler) sendAutorization(lastIdSend string, lastNumberMessage int) {

	configWrite := kafka.WriterConfig{
		Brokers: []string{lh.Url},
		Topic:   lastIdSend}

	writer := kafka.NewWriter(configWrite)

	message := Message{TypeMessage: AutorizationOK, Id_node: lh.ID_LEADER, Id_message: lastNumberMessage}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	Log.Println("Il leader autorizza il nodo " + lastIdSend + " ad accedere alla sezione critica")
	checkErr(err)
}

/*
 * Con questa funzione il leader attende l'ack dal nodo a cui è stato appena
 * concesso di entrare in sezione critica. Il messaggio deve arrivare prima dello scadere
 * di un timeoout.
 */
func (lh LeaderHandler) waitAck(lastIdSend string, lastNumberMessage int) bool {
	configRead := kafka.ReaderConfig{
		Brokers:  []string{lh.Url},
		Topic:    lh.ID_LEADER,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configRead)

	var messageReceved Message

	timeout := time.Now()
	for time.Since(timeout) < 3*time.Second {

		contextTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		messageKafka, _ := reader.ReadMessage(contextTimeout)
		json.Unmarshal(messageKafka.Value, &messageReceved)

		if messageReceved.TypeMessage == AutorizationACK && messageReceved.Id_message == lastNumberMessage && messageReceved.Id_node == lastIdSend {
			Log.Println("Il leader riceve l'ack dal nodo: " + lastIdSend)
			return true
		}

	}

	Log.Println("Il leader NON riceve l'ack dal nodo: " + lastIdSend)
	return false
}

/*
 * Con questa funzione il leader attende il rilascio dal nodo a cui è
 * concesso di entrare in sezione critica. Il messaggio deve arrivare prima dello scadere
 * di un timeoout.
 */
func (lh LeaderHandler) waitRelease(lastIdSend string, lastNumberMessage int) {
	configRead := kafka.ReaderConfig{
		Brokers:  []string{lh.Url},
		Topic:    lh.ID_LEADER,
		MaxBytes: 10e6}
	readerLeader := kafka.NewReader(configRead)

	var messageReceved Message

	timeout := time.Now()
	for time.Since(timeout) < 3*time.Second {

		contextTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		messageKafka, _ := readerLeader.ReadMessage(contextTimeout)
		json.Unmarshal(messageKafka.Value, &messageReceved)

		if messageReceved.TypeMessage == AutorizationRelease && messageReceved.Id_message == lastNumberMessage && messageReceved.Id_node == lastIdSend {
			Log.Println("Il leader riceve release dal nodo: " + lastIdSend)
			return
		}

	}

	Log.Println("Il leader NON riceve release dal nodo: " + lastIdSend)

}
