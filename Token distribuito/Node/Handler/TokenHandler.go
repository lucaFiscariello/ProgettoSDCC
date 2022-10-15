/****************************************************************************************
* Questo handler ha il compito di gestire la ricezione di l'invio del token dal nodo	*
* corrente a un nuovo nodo. Oltre che a richiedere la generazione di un nuovo token		*
* quando necessario.																	*
*****************************************************************************************/

package handler

import (
	"context"
	"encoding/json"
	Log "fiscariello/luca/node/Logger"
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

/*
 * Questa funzione selezione il prossimo nodo attivo a cui inviare il token. La scelta è fatta in maniera ciclica.
 * Il nodo1 invia il messaggio al nodo2 , il nodo2 al nodo3 e cosi via. Se un nodo risulta essere non attivo per la ricezione si passa
 * a quello successivo. Ad esempio da node1 a node3 se node2 è non attivo.
 */
func (h TokenHandler) SendToken(is_active map[string]bool) {

	idTemplate := h.ID_NODE
	idNumber := h.ID_NODE

	//scompongo l'id del nodo in due componenti: il prefisso "node" e il numero identificativo
	//Ad esempio node1 verrà scomposto in node e 1
	regexprTemplate := regexp.MustCompile(`[^a-zA-Z ]+`)
	idTemplatenew := regexprTemplate.ReplaceAllString(idTemplate, "")
	regexprNumber := regexp.MustCompile("[^0-9]+")
	idNumbernew := regexprNumber.ReplaceAllString(idNumber, "")

	id, err := strconv.Atoi(idNumbernew)
	checkErr(err)

	numberNode := len(is_active) + 1

	//scorro i nodi per trovarne uno attivo a cui inviare il token
	for i := 1; i < 2*numberNode; i++ {
		numId := (i + id) % numberNode
		idToSend := idTemplatenew + fmt.Sprint(numId)

		if is_active[idToSend] {
			configWrite := kafka.WriterConfig{
				Brokers: []string{h.Url},
				Topic:   idToSend}

			writer := kafka.NewWriter(configWrite)

			message := Message{TypeMessage: Token, Id_node: h.ID_NODE}
			messageByte, _ := json.Marshal(message)
			writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

			Log.Println("Il nodo corrente invia il token a: " + idToSend + ". Gli altri nodi della rete sono: " + fmt.Sprint(is_active))

			return
		}
	}
}

/*
 * Questa funzione si mette in ascolto di messaggi di consegna del token.
 */
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
			Log.Println("Il nodo corrente riceve token da: " + messageReceved.Id_node)
			channel <- true
		}
	}

}

/*
 * Questa funzione richiede la generazione del token quando necessario. L'output è booleano.
 * La generazione potrebbe non essere concessa.
 */
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

	Log.Println("Il nodo corrente invia controllo periodico per verificare se token checker è ancora attivo.")

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
			Log.Println("Leader non risulta attivo")
			return false
		}

		if messageReceved.TypeMessage == TokenResponse && messageReceved.Id_message == Id_message {
			Log.Println("Leader genera nuovo token?(true/false) " + strconv.FormatBool(messageReceved.ConcessToken))
			h.TOKEN = messageReceved.ConcessToken
			return messageReceved.ConcessToken
		}
	}

}

/*
 * Questa funzione invia un messaggio di ack al token checker al momento dell'arrivo del token.
 * In questo modo il token checker potrà sapere se un token è ancora in circolo nella rete
 */
func (h TokenHandler) SendHackToken(leaderID string) {

	configWrite := kafka.WriterConfig{
		Brokers: []string{h.Url},
		Topic:   leaderID}

	writerLeader := kafka.NewWriter(configWrite)

	message := Message{TypeMessage: TokenCheck, Id_node: h.ID_NODE}
	messageByte, _ := json.Marshal(message)
	writerLeader.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

	Log.Println("Il nodo corrente invia un ack dopo la ricezione del token al leader: " + leaderID)

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
