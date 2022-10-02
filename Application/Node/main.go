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

var ALL_ARTICLE H.Data
var ALL_NODE map[string]bool = make(map[string]bool)
var serverWSIsReady bool = false

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

func main() {

	timeout, err := strconv.Atoi(TIMEOUT)
	checkErr(err)
	time.Sleep(time.Duration(timeout) * time.Second)

	createNewTopicKafka()
	createPresentation()
	retriveArticle()

	go updateNode()
	go startChannelHandler()
	go startHeartBeatHandler()

	waitConnectionWS()

	for _, article := range ALL_ARTICLE.Articles {

		if TOKEN == "YES" {
			sendMessage(article)
		}

		time.Sleep(4 * time.Second)
	}

}

func createNewTopicKafka() {
	handler := H.KafkatHandler{ID_NODE: ID_NODE, Url: URLKAFKA}
	handler.CreateNewTopicKafka()
}

func createPresentation() {
	handler := H.KafkatHandler{ID_NODE: ID_NODE, Url: URLKAFKA}
	handler.CreatePresentation(PRESENTATION_TOPIC)
}

func retriveArticle() {
	url := fmt.Sprintf(URLTEMPLATE, TEMA, APIKEY)
	handler := H.ApiHandler{Url: url}
	ALL_ARTICLE = handler.RetriveArticle()
}

func updateNode() {

	config := kafka.ReaderConfig{
		Brokers:  []string{URLKAFKA},
		Topic:    PRESENTATION_TOPIC,
		MaxBytes: 10e6}
	reader := kafka.NewReader(config)

	for {
		message, err := reader.ReadMessage(context.Background())
		checkErr(err)

		id_node := string(message.Value[:])

		if !contains(ALL_NODE, id_node) {
			ALL_NODE[id_node] = false
		}
	}

}

func startHeartBeatHandler() {

	configReadNode := kafka.ReaderConfig{
		Brokers:  []string{URLKAFKA},
		Topic:    ID_NODE,
		MaxBytes: 10e6}

	configWrite := kafka.WriterConfig{
		Brokers: []string{URLKAFKA},
		Topic:   HEARTBEAT_TOPIC}

	reader := kafka.NewReader(configReadNode)
	writer := kafka.NewWriter(configWrite)
	handler := H.HeartBeatHandler{ID_NODE: ID_NODE}

	for {

		handler.SendHeart(writer)
		handler.Wait(reader, ALL_NODE)

		fmt.Println("fine primo ciclo")
		fmt.Println(ALL_NODE)
		time.Sleep(5 * time.Second)

	}

}

func startChannelHandler() {
	handler := H.ChanneltHandler{Url: URLKAFKA, ID_NODE: ID_NODE, HEARTBEAT_TOPIC: HEARTBEAT_TOPIC}
	handler.SendBeat()

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

	r, err := server.SendArticle(ctx, &pb.Article{Title: article.Title, Description: article.Description, UrlToImage: article.UrlToImage})
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
