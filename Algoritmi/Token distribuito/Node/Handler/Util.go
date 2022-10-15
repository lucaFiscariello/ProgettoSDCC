package handler

type Message struct {
	TypeMessage  string `json:"TypeMessage"`
	Id_node      string `json:"Id_node"`
	Id_message   int    `json:"Id_message"`
	ConcessToken bool   `json:"ConcessToken"`
}

const (
	messageType   string = "messageType"
	HeartBeat            = "HeartBeat"
	Token                = "Token"
	TokenRequest         = "TokenRequest"
	TokenCheck           = "TokenCheck"
	TokenResponse        = "TokenResponse"
)
