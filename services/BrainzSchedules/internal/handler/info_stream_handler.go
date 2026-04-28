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

const (
	wsWriteWait  = 10 * time.Second
	wsPongWait   = 60 * time.Second
	wsPingPeriod = (wsPongWait * 9) / 10
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

		// Single writer goroutine: all writes (data + ping) must be serialized.
		go func() {
			ticker := time.NewTicker(wsPingPeriod)
			defer ticker.Stop()
			for {
				select {
				case msg, ok := <-client.Send:
					if !ok {
						return
					}
					_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
					if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
						log.Println("write error:", err)
						return
					}
				case <-ticker.C:
					_ = conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						log.Println("ping error:", err)
						return
					}
				}
			}
		}()

		_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
		conn.SetPongHandler(func(string) error {
			_ = conn.SetReadDeadline(time.Now().Add(wsPongWait))
			return nil
		})
		// Keep default ping handler; it replies with pong safely.

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
