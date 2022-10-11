package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	"github.com/segmentio/kafka-go"

	S "fiscariello/luca/webserver/grpc/implementation"
)

var PORT = os.Getenv("PORT")
var START = os.Getenv("START_TOPIC")
var IP_KAFKA = os.Getenv("IP_KAFKA")
var PORT_KAFKA = os.Getenv("PORT_KAFKA")

var upgrader = websocket.Upgrader{}
var start = "start"
var sender S.Sender

func main() {

	sender = S.Sender{NewConnection: true}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./webSite/index.html")
	})

	http.HandleFunc("/Connection/WebSocket", func(w http.ResponseWriter, r *http.Request) {

		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(w, r, nil)

		for err != nil {
			conn, err = upgrader.Upgrade(w, r, nil)
		}

		if sender.NewConnection {
			notifyNode()
			sender.RegisterServiceServer(PORT, conn)
		} else {
			sender.RefreshConnection(conn)
		}

	})

	http.ListenAndServe(":8080", nil)

}

func notifyNode() {

	url := IP_KAFKA + PORT_KAFKA

	config := kafka.WriterConfig{
		Brokers: []string{url},
		Topic:   START}

	writer := kafka.NewWriter(config)
	err := writer.WriteMessages(context.Background(), kafka.Message{Value: []byte(start)})
	checkErr(err)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
