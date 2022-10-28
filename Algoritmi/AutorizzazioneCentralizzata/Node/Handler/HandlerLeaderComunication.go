/****************************************************************************************
* Questo handler gestisce la comunicazione con tra nodo corrente e il leader			*
* il quale gestisce l'accesso alla sezione critica.										*
* In particolare l'Handler implementa funzionalità per:									*
*	- richiedere accesso alla sezione critica al leader									*
*	- inviare un ack in seguito alla ricezione dell'autorizzazione del leader			*
*	- inviare messaggio di rilascio quando il nodo corrente ha intezione di uscire		*
*	  dalla sezione critica.															*
*****************************************************************************************/

package handler

import (
	"context"
	"encoding/json"
	"time"

	Log "fiscariello/luca/node/Logger"

	"github.com/segmentio/kafka-go"
)

type LeaderComunicationHandler struct {
	Url       string
	ID_LEADER string
	ID_NODE   string
}

var lastID = 0
var reader *kafka.Reader = nil

/*
 * Con questa funzione il nodo corrente richiede l'autorizzazione per entrare nella sezione critica e attende una
 * risposta dal leader. La risposta deve arrivare in un certo intervallo temporale. Allo scadere di un timeout
 * il nodo corrente assumerà che il leader sia down.
 */
func (lh *LeaderComunicationHandler) CanExecute() bool {

	lh.singletonReader()

	config := kafka.WriterConfig{
		Brokers: []string{lh.Url},
		Topic:   lh.ID_LEADER}

	writer := kafka.NewWriter(config)

	lastID = lastID + 1

	//invio autorizzazione
	message := Message{TypeMessage: RequestAutorization, Id_node: lh.ID_NODE, Id_message: lastID}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

	Log.Println("Il node corrente richiede di entrare in sezione critica.")

	//setto tempo di attesa massimo
	contextTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//attendo risposta
	for {
		var messageReceved Message

		messageKafka, err := reader.ReadMessage(contextTimeout)
		err = json.Unmarshal(messageKafka.Value, &messageReceved)

		if err != nil {
			Log.Println("Il leader non è contattabile.")
			return false
		}

		if messageReceved.TypeMessage == AutorizationOK && messageReceved.Id_message == lastID {
			lh.SendAck(writer)

			Log.Println("Il nodo corrente ottiene autorizzazione dal leader.")
			return true
		}

	}
}

/*
 * Funzione usata per inviare un messaggio di ack al leader dopo aver ottenuto l'autorizzazione ad accedere alla sezione critica
 */
func (lh *LeaderComunicationHandler) SendAck(writer *kafka.Writer) {
	message := Message{TypeMessage: AutorizationACK, Id_node: lh.ID_NODE, Id_message: lastID}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

	Log.Println("Il nodo corrente invia un ack al leader.")

}

/*
 * Funzione usata per indicare l'intezione di uscire dalla sezione critica. Questo messaggio è inviato solo al leader.
 */
func (lh *LeaderComunicationHandler) Release() {

	config := kafka.WriterConfig{
		Brokers: []string{lh.Url},
		Topic:   lh.ID_LEADER}

	writer := kafka.NewWriter(config)

	message := Message{TypeMessage: AutorizationRelease, Id_node: lh.ID_NODE, Id_message: lastID}
	messageByte, _ := json.Marshal(message)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})
	checkErr(err)

	Log.Println("Il nodo corrente rilascia l'autorizzazione")

}

func (lh *LeaderComunicationHandler) singletonReader() {

	if reader == nil {
		configRead := kafka.ReaderConfig{
			Brokers:  []string{lh.Url},
			Topic:    lh.ID_NODE,
			MaxBytes: 10e6}

		reader = kafka.NewReader(configRead)
	}

}
