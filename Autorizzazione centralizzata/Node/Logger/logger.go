package logger

import (
	"log"
	"os"
)

var Logger *log.Logger = nil

func Inizialize() {

	if Logger == nil {
		file, _ := os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		Logger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	}
}

func Println(text string) {
	Logger.Println(text)
	log.Println(text)
}
