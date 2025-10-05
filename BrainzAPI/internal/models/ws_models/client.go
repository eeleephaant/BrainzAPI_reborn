package wsmodels

import "github.com/hertz-contrib/websocket"

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}
