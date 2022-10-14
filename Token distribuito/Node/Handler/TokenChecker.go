/*************************************************************************************
*	Questo handler ha il compito di supervisionare la comunicazione e assicurarsi che
* 	i nodi continuino a comunicare senza entrare in deadlock. Il suo compito è quello di
*	verificare che un token sia in circolo nella rete. Se il token è perso ne genera
*	uno nuovo.
**************************************************************************************/

package handler

import (
	"context"
	"encoding/json"
	Log "fiscariello/luca/node/Logger"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type TokenChecker struct {
	Url       string
	ID_NODE   string
	WAIT_TIME string
}

/*
 * Questa funzione avvia un listner che si mette in ascolto di richieste di generazione di token da parte degli
 * altri nodi della rete. Se il token è gia stato creato non viene rigenerato. Viene rigenerato solo se
 * allo scadere di un timeout nessun nodo dimostra di avere il token. In particolare il possessore del token
 * dovrà inviare periodicamente un messaggio di check al checker per informare che il token è ancora in circolo nella rete.
 */
func (tc TokenChecker) Start() {

	var messageReceved Message

	var timeout time.Time = time.Now()
	var TokenGenerated = false
	var concessToken = true

	configReader := kafka.ReaderConfig{
		Brokers:  []string{tc.Url},
		Topic:    tc.ID_NODE,
		MaxBytes: 10e6}

	reader := kafka.NewReader(configReader)

	for {
		messageKafka, err := reader.ReadMessage(context.Background())
		json.Unmarshal(messageKafka.Value, &messageReceved)
		checkErr(err)

		// quando arriva una richiesta di generazione di token, valuto se generarlo o no.
		if messageReceved.TypeMessage == TokenRequest {
			configWrite := kafka.WriterConfig{
				Brokers: []string{tc.Url},
				Topic:   messageReceved.Id_node}

			writer := kafka.NewWriter(configWrite)

			if !TokenGenerated {
				TokenGenerated = true
				timeout = time.Now()
			} else if time.Since(timeout) > 15*time.Second {
				concessToken = true
			} else {
				concessToken = false
			}

			Log.Println("Il token checker riceve richiesta token da " + messageReceved.Id_node + ". Token generato?(true/false) " + strconv.FormatBool(concessToken))

			message := Message{TypeMessage: TokenResponse, ConcessToken: concessToken, Id_message: messageReceved.Id_message}
			messageByte, _ := json.Marshal(message)
			writer.WriteMessages(context.Background(), kafka.Message{Value: messageByte})

		} else if messageReceved.TypeMessage == TokenCheck {

			//Azzero timeout. Il token è ancora in circolo nella rete
			timeout = time.Now()
			Log.Println("Il token checker riceve un check da: " + messageReceved.Id_node + ". Il token è correttamente in circolo.")
		}

	}

}
