package handler

type Message struct {
	TypeMessage string `json:"TypeMessage"`
	Id_node     string `json:"Id_node"`
	Id_message  int    `json:"Id_message"`
}

const (
	messageType         string = "messageType"
	HeartBeat                  = "HeartBeat"
	Token                      = "Token"
	TokenRequest               = "TokenRequest"
	TokenCheck                 = "TokenCheck"
	RequestAutorization        = "RequestAutorization"
	AutorizationOK             = "AutorizationOK"
	AutorizationRelease        = "AutorizationRelease"
	AutorizationACK            = "AutorizationACK"
	RequestLamport             = "RequestLamport"
	ACKLamport                 = "ACKLamport"
	ReleaseLamport             = "ReleaseLamport"
)
