package main

import (
	"context"
	H "fiscariello/luca/node/Handler"
	pb "fiscariello/luca/node/stub"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc"
)

const (
	HaveToken string = "HaveToken"
	YES              = "YES"
	NO               = "NO"
)

var ALL_ARTICLE H.Data
var ALL_NODE map[string]bool = make(map[string]bool)

var TIMEOUT = os.Getenv("TIMEOUT")
var IP_KAFKA = os.Getenv("IP_KAFKA")
var PORT_KAFKA = os.Getenv("PORT_KAFKA")
var PRESENTATION_TOPIC = os.Getenv("PRESENTATION_TOPIC")
var HEARTBEAT_TOPIC = os.Getenv("HEARTBEAT_TOPIC")
var TOKENCHECK_TOPIC = os.Getenv("TOKENCHECK_TOPIC")
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

	waitStartNode()

	handlerKafka := H.KafkatHandler{ID_NODE: ID_NODE, Url: URLKAFKA}
	handlerAPI := H.ApiHandler{Url: URLAPI}
	handlerHeartBeat := H.HeartBeatHandler{ID_NODE: ID_NODE, Url: URLKAFKA}
	handlerToken := H.TokenHandler{ID_NODE: ID_NODE, Url: URLKAFKA, TOKEN: TOKEN, TOKEN_TOPIC: TOKENCHECK_TOPIC}
	handlerNode := H.NodeActiveHandler{ID_NODE: ID_NODE, Url: URLKAFKA, PRESENTATION_TOPIC: PRESENTATION_TOPIC, ALL_NODE: ALL_NODE}

	handlerKafka.CreateNewTopicKafka()
	handlerKafka.CreatePresentation(PRESENTATION_TOPIC)
	ALL_ARTICLE = handlerAPI.RetriveArticle()

	go handlerNode.ListenNewNode()
	go startHeartBeatHandler(&handlerHeartBeat, &handlerNode)
	go listenToken(&handlerToken)

	waitConnectionWS()

	for i := 0; i < len(ALL_ARTICLE.Articles); {

		if handlerToken.TOKEN == YES {
			sendMessage(ALL_ARTICLE.Articles[i])

			if handlerNode.GetNumberNode() > 0 {
				handlerToken.SendToken(handlerNode.GetNode())
				handlerToken.TOKEN = NO
				i++
			}
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

	readerBeat := kafka.NewReader(configReadNodeBeat)
	readerHeart := kafka.NewReader(configReadNodeHeart)
	writerHeart := kafka.NewWriter(configWriteHeart)

	go handler.SendBeat(readerHeart)

	for {

		handler.SendHeart(writerHeart)
		handler.Wait(readerBeat, handlerNode)

		fmt.Println(handlerNode.ALL_NODE)
		time.Sleep(5 * time.Second)

	}

}

func listenToken(handlerToken *H.TokenHandler) {
	channel := make(chan string)
	go handlerToken.ListenReceveToken(channel)

	for {
		token := <-channel
		fmt.Println("token arrivato")
		handlerToken.TOKEN = token
	}

}
func waitConnectionWS() {
	configReadNode := kafka.ReaderConfig{
		Brokers:  []string{URLKAFKA},
		Topic:    START,
		MaxBytes: 10e6}

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

	r, err := server.SendArticle(ctx, &pb.Article{Title: article.Title, Description: article.Description, UrlToImage: article.UrlToImage, UrlSite: article.Url})
	checkErr(err)
	fmt.Print(r)
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
