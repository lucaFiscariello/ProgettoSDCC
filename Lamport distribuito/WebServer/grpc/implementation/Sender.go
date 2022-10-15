/********************************************************************************
* Entità incaricata di scrivere articoli sulla pagina web tramite web socket.	*
* Il sender riceverà richieste di scrittura dai nodi della rete tramite grpc.	*
*********************************************************************************/

package impelementation

import (
	"context"
	"encoding/json"
	pb "fiscariello/luca/webserver/grpc/stub"

	"fmt"
	"log"
	"net"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

type Sender struct {
	pb.UnimplementedServiceServer
	connectionWS  *websocket.Conn
	NewConnection bool
}

func (sender *Sender) RegisterServiceServer(PORT string, connection *websocket.Conn) {
	sender.connectionWS = connection

	lis, err := net.Listen("tcp", PORT)
	checkErr(err)

	s := grpc.NewServer()
	pb.RegisterServiceServer(s, sender)
	err = s.Serve(lis)
	checkErr(err)

	log.Printf("server listening at %v", lis.Addr())

}

/*
 * Questo metodo deve essere invocato in maniera mutuamente esclisiva
 */
func (sender *Sender) SendArticle(ctx context.Context, article *pb.Article) (*pb.Response, error) {

	fmt.Println("server invia " + article.Title)

	//Scrittura dell'articolo tramite web socket
	byteMessage, err := json.Marshal(article)
	err = sender.connectionWS.WriteMessage(1, byteMessage)
	checkErr(err)

	return &pb.Response{}, nil
}

func (sender *Sender) RefreshConnection(connection *websocket.Conn) {
	sender.connectionWS = connection
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
