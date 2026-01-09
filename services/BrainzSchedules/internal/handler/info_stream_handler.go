package handler

import (
	"brainz-api/internal/models"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
)

var upgrader = websocket.HertzUpgrader{
	CheckOrigin: func(r *app.RequestContext) bool {
		return true
	},
}

func InfoStreamHandler(ctx context.Context, c *app.RequestContext) {
	institutionID := string(c.QueryArgs().Peek("institution_id"))
	if institutionID == "" {
		c.String(http.StatusBadRequest, "missing institution_id")
		return
	}

	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		client := &models.Client{
			Conn: conn,
			Send: make(chan []byte, 10),
		}

		channel := fmt.Sprintf("info_stream:%s", institutionID)
		models.MainHub.AddClient(channel, client)
		defer models.MainHub.RemoveClient(channel, client)
		defer conn.Close()
		defer close(client.Send)

		log.Printf("Client connected to %s", channel)

		go func() {
			for msg := range client.Send {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.Println("write error:", err)
					return
				}
			}
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				log.Println("client disconnected:", err)
				return
			}
		}
	})

	if err != nil {
		log.Println("upgrade error:", err)
	}
}
