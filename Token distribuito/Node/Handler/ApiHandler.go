/*************************************************************************************
*	Questo handler ha l'obiettivo di scaricare gli articoli di un determinato tema
*	tramite un API. Questi articoli verranno poi inviate dal programma principale
*	(main.go) al Sender, incaricato di pubblicare l'articolo sulla pagina web tramite
* 	web socket.
**************************************************************************************/

package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Article struct {
	Source      Source `json:"source"`
	Author      string `json:"author"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Url         string `json:"url"`
	UrlToImage  string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
	Content     string `json:"content"`
}

type Data struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

type ApiHandler struct {
	Url string
}

type HandlerAPI interface {
	RetriveArticle()
}

/*
 * La funzione scarica dati dall'api e li restituisce.
 */
func (ha ApiHandler) RetriveArticle() Data {
	var data Data

	resp, err := http.Get(ha.Url)
	checkErr(err)

	respByte, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	json.Unmarshal(respByte, &data)
	return data
}
