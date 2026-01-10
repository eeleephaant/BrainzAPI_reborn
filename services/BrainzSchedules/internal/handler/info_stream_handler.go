package handler

import (
	"brainz-api/internal/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.HertzUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(ctx *app.RequestContext) bool {
		origin := string(ctx.GetHeader("Origin"))
		allowedOrigins := map[string]bool{
			"https://pentapulse.ru": true,
			"":                      true,
		}
		return allowedOrigins[origin]
	},
	EnableCompression: true,
	HandshakeTimeout:  10 * time.Second,
}

func InfoStreamHandler(ctx context.Context, c *app.RequestContext) {
	instIDStr := string(c.QueryArgs().Peek("institution_id"))
	if instIDStr == "" {
		c.String(http.StatusBadRequest, "missing institution_id")
		return
	}

	instID, err := strconv.ParseInt(instIDStr, 10, 64)
	if err != nil || instID <= 0 {
		c.String(http.StatusBadRequest, "invalid institution_id")
		return
	}

	channel := fmt.Sprintf("info_stream:%d", instID)

	err = upgrader.Upgrade(c, func(conn *websocket.Conn) {
		client := &models.Client{
			ID:   fmt.Sprintf("%d-%s", time.Now().UnixNano(), conn.RemoteAddr().String()),
			Conn: conn,
			Send: make(chan []byte, 64),
		}

		models.MainHub.AddClient(channel, client)
		defer models.MainHub.RemoveClient(channel, client)
		defer conn.Close()
		defer close(client.Send)

		zap.L().Info("WebSocket client connected",
			zap.String("channel", channel),
			zap.String("client_id", client.ID),
			zap.String("remote_addr", conn.RemoteAddr().String()))

		// Отдельная горутина для отправки сообщений
		go func() {
			for msg := range client.Send {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.Println("write error:", err)
					return
				}
			}
		}()

		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		conn.SetPingHandler(func(string) error {
			conn.WriteMessage(websocket.PongMessage, nil)
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				log.Println("client disconnected:", err)
				return
			}
		}
	})

	if err != nil {
		zap.L().Error("websocket upgrade failed", zap.Error(err))
		c.String(http.StatusInternalServerError, "failed to upgrade connection")
	}
}
