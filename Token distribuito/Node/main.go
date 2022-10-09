package main

import (
	"context"
	H "fiscariello/luca/node/Handler"
	pb "fiscariello/luca/node/stub"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

var ALL_ARTICLE H.Data
var ALL_NODE map[string]bool = make(map[string]bool)
var isLeader bool
var leaderID string
var checkerActive = false

var TIMEOUT = os.Getenv("TIMEOUT")
var IP_KAFKA = os.Getenv("IP_KAFKA")
var PORT_KAFKA = os.Getenv("PORT_KAFKA")
var PRESENTATION_TOPIC = os.Getenv("PRESENTATION_TOPIC")
var HEARTBEAT_TOPIC = os.Getenv("HEARTBEAT_TOPIC")
var ID_NODE = os.Getenv("ID_NODE")
var TOKEN = os.Getenv("HAVE_TOKEN")
var TEMA = os.Getenv("TEMA")
var APIKEY = os.Getenv("APIKEY")
var URLTEMPLATE = os.Getenv("URLTEMPLATE")
var ID_WS = os.Getenv("ID_WS")
var START = os.Getenv("START_TOPIC")
var URLKAFKA = IP_KAFKA + PORT_KAFKA
var URLAPI = fmt.Sprintf(URLTEMPLATE, TEMA, APIKEY)

func main() {

	// Attendo lo scadere di un timeout prima di avviare il nodo
	waitStartNode()

	configRead := kafka.ReaderConfig{
		Brokers:  []string{URLKAFKA},
		Topic:    ID_NODE,
		MaxBytes: 10e6}

	//Reader e writer configurati per interagire con il generatore del token
	reader := kafka.NewReader(configRead)

	//Handler per gestire la logica del nodo
	handlerKafka := H.KafkatHandler{ID_NODE: ID_NODE, Url: URLKAFKA}
	handlerAPI := H.ApiHandler{Url: URLAPI}
	handlerHeartBeat := H.HeartBeatHandler{ID_NODE: ID_NODE, Url: URLKAFKA}
	handlerToken := H.TokenHandler{ID_NODE: ID_NODE, Url: URLKAFKA, ReaderTokenLeader: reader}
	handlerNode := H.NodeActiveHandler{ID_NODE: ID_NODE, Url: URLKAFKA, PRESENTATION_TOPIC: PRESENTATION_TOPIC, ALL_NODE: ALL_NODE}

	//Creazione canale di comunicazione "privato" del nodo in cui potrà ricevere token o messaggi di heart beat
	handlerKafka.CreateNewTopicKafka()

	/*
	 *Pubblicazione del messaggio di presentazione del nodo.
	 * Con questo messaggio il nodo si presenta agli altri nodi della rete.
	 */
	handlerKafka.CreatePresentation(PRESENTATION_TOPIC)

	//Contatto l'API per scaricare tutti gli articoli di un determinato tema
	ALL_ARTICLE = handlerAPI.RetriveArticle()

	go handlerNode.ListenNewNode()                            // Creo un goroutine che si mette in ascolto di nuovi messaggi di presentazione
	go startHeartBeatHandler(&handlerHeartBeat, &handlerNode) // Avvio il gestore dell'heart beat
	go listenToken(&handlerToken)                             // Avvio il listner che si mette in ascolto per l'arrivo del token

	fmt.Println("Nodo avviato")

	/*
	 * Attendo venga aperta la connessione Web socket.
	 * Questa connessione si concretizza quando viene aperta la pagina web tramite browser
	 */
	waitConnectionWS()

	for i := 0; i < len(ALL_ARTICLE.Articles); {

		//Ricerco leader. Se il nodo corrente è il leader avvia il gestore del token.
		isLeader, leaderID = serchLeader(handlerNode.GetNode())

		if isLeader && !checkerActive {
			fmt.Println("eletto nuovo leader")
			checkerToken := H.TokenChecker{Url: URLKAFKA, ID_NODE: leaderID, WAIT_TIME: TIMEOUT}
			checkerActive = true
			go checkerToken.Start()
		}

		//Se il nodo possiede il token puo accedere alla sezione critica
		haveToken := handlerToken.RequestToken(leaderID)
		if haveToken {
			fmt.Println("entro sezione critica")
			sendMessage(ALL_ARTICLE.Articles[i])

			if handlerNode.GetNumberNode() > 0 {

				/*Questo metodo contiene un invocazione ad una rpc.
				 *la procedura a chiamata remota invia l'articolo ad un nodo che lo pubblicherà sulla pagina web
				 *tramite la connessione con Web Socket
				 */
				handlerToken.SendToken(handlerNode.GetNode())
			}

			handlerToken.TOKEN = false
			i++
		}

		time.Sleep(4 * time.Second)
	}

}

func waitStartNode() {
	timeout, err := strconv.Atoi(TIMEOUT)
	checkErr(err)
	time.Sleep(time.Duration(timeout) * time.Second)
}

func startHeartBeatHandler(handler *H.HeartBeatHandler, handlerNode *H.NodeActiveHandler) {

	configReadNodeBeat := kafka.ReaderConfig{
		Brokers:  []string{URLKAFKA},
		Topic:    ID_NODE,
		MaxBytes: 10e6}

	configReadNodeHeart := kafka.ReaderConfig{
		Brokers:  []string{URLKAFKA},
		Topic:    HEARTBEAT_TOPIC,
		MaxBytes: 10e6}

	configWriteHeart := kafka.WriterConfig{
		Brokers: []string{URLKAFKA},
		Topic:   HEARTBEAT_TOPIC}

	readerBeat := kafka.NewReader(configReadNodeBeat)   //Reader che legge messaggi di beat sul canale "privato" del nodo
	readerHeart := kafka.NewReader(configReadNodeHeart) // Reader che legge messagi di heart sul canale condiviso tra tutti i nodi
	writerHeart := kafka.NewWriter(configWriteHeart)    // Writer che pubblica messaggi di heart sul canale condiviso

	/*
	 * Avvio goroutine che si mette in ascolto dei messaggi di heart sul canale comune
	 * e risponde al nodo che ha inviato la richiesta con un messaggio di beat
	 */
	go handler.SendBeat(readerHeart)

	for {

		/*
		 *Periodicamente invio messaggi di heart e attendo che arrivino tutti i messaggi di beat.
		 *Tutti i nodi che risultano essere attivi vengono memorizzatti in una mappa
		 */
		handler.SendHeart(writerHeart)
		handler.Wait(readerBeat, handlerNode)
		time.Sleep(5 * time.Second)

	}

}

func listenToken(handlerToken *H.TokenHandler) {
	channel := make(chan bool)
	go handlerToken.ListenReceveToken(channel)

	for {
		token := <-channel
		fmt.Println("token arrivato")
		handlerToken.TOKEN = token
		handlerToken.SendHackToken(leaderID)
	}

}
func waitConnectionWS() {
	configReadNode := kafka.ReaderConfig{
		Brokers:  []string{URLKAFKA},
		Topic:    START,
		MaxBytes: 10e6}

	/*
	 *Mi metto in ascolto su un topic in cui verrà pubblicato un messaggio di "start" nel momento in cui
	 *verrà aperta la connesione Web Socket
	 */
	reader := kafka.NewReader(configReadNode)
	reader.ReadMessage(context.Background())
}

func sendMessage(article H.Article) {

	conn, err := grpc.Dial(ID_WS, grpc.WithInsecure(), grpc.WithBlock())
	checkErr(err)
	defer conn.Close()

	server := pb.NewServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	//Sfrutto una rpc per contattare il nodo incaricato di scrivere l'articolo sul sito tramite web socket
	r, err := server.SendArticle(ctx, &pb.Article{Title: article.Title, Description: article.Description, UrlToImage: article.UrlToImage, UrlSite: article.Url})
	checkErr(err)
	fmt.Print(r)
}

func serchLeader(allNode map[string]bool) (bool, string) {

	minID := math.MaxInt64
	var minIDStr string
	var isLeader = false

	for nodeId, isActive := range allNode {

		if isActive {
			regexprNumber := regexp.MustCompile("[^0-9]+")
			idNumberSTR := regexprNumber.ReplaceAllString(nodeId, "")

			idNumber, err := strconv.Atoi(idNumberSTR)
			checkErr(err)

			if idNumber < minID {
				minID = idNumber
				minIDStr = nodeId
			}

		}

	}

	if ID_NODE < minIDStr {
		minIDStr = ID_NODE
		isLeader = true
	}

	return isLeader, minIDStr
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func contains(allNode map[string]bool, id_search string) bool {
	if id_search == ID_NODE {
		return true
	}

	for id := range allNode {

		if id == id_search {
			return true
		}
	}
	return false
}
