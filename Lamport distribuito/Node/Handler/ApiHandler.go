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

func (ha ApiHandler) RetriveArticle() Data {
	var data Data

	resp, err := http.Get(ha.Url)
	checkErr(err)

	respByte, err := ioutil.ReadAll(resp.Body)
	checkErr(err)

	json.Unmarshal(respByte, &data)
	return data
}
