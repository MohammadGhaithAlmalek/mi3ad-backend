package ws

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
)

var upgrader = websocket.Upgrader{
	// TODO tighten this before going to production - allows any origin for now
	CheckOrigin: func(r *http.Request) bool { return true },
}

// RegisterRoutes wires up the GET /ws upgrade endpoint.
func RegisterRoutes(e *echo.Echo, hub *Hub) {
	e.GET("/ws", func(c *echo.Context) error {
		userID := c.QueryParam("userId") // replace with the authenticated user id once JWT auth is wired in
		if userID == "" {
			return c.String(http.StatusBadRequest, "userId is required")
		}

		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		client := newClient(userID, conn, hub)
		hub.register(userID, client)
		go client.readLoop()

		return nil
	})
}
