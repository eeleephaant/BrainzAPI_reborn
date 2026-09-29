package models

import "github.com/hertz-contrib/websocket"

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}
