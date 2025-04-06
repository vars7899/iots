package ws

import "github.com/gorilla/websocket"

type Client struct {
	ID     string
	conn   *websocket.Conn
	SendCh chan []byte
}
