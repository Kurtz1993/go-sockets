package main

import (
	"fmt"
	"log"

	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// The key for the map is message.to
	clients := make(map[string]string)

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", c.Query("user_id"))
		return c.Next()
	})

	// Multiple event handling supported
	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		log.Println("fired connect 1")
	})

	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		log.Println("fired connect 2")
	})

	ikisocket.On(ikisocket.EventMessage, func(ep *ikisocket.EventPayload) {
		log.Println("fired message: " + string(ep.Data))
		// Emit the message directly to specified user
		err := ikisocket.EmitTo(ep.Kws.UUID, ep.Data)
		if err != nil {
			log.Println(err)
		}
	})

	ikisocket.On(ikisocket.EventDisconnect, func(ep *ikisocket.EventPayload) {
		log.Println("fired disconnect" + ep.Error.Error())
	})

	app.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		// Retrieve user id from the middleware (optional)
		userId := fmt.Sprintf("%v", kws.Locals("user_id"))

		// Every websocket connection has an optional session key => value storage
		kws.SetAttribute("user_id", userId)

		clients[userId] = kws.UUID
	}))

	ikisocket.On("close", func(payload *ikisocket.EventPayload) {
		log.Printf("fired close %s", payload.SocketAttributes["user_id"])
	})

	// app.Use("/ws", func(c *fiber.Ctx) error {
	// 	// IsWebSocketUpgrade returns true if the client
	// 	// requested upgrade to the WebSocket protocol.
	// 	if websocket.IsWebSocketUpgrade(c) {
	// 		c.Locals("allowed", true)
	// 		return c.Next()
	// 	}
	// 	return fiber.ErrUpgradeRequired
	// })

	// app.Get("/ws/:id", websocket.New(func(c *websocket.Conn) {
	// 	// c.Locals is added to the *websocket.Conn
	// 	log.Println(c.Locals("allowed"))  // true
	// 	log.Println(c.Params("id"))       // 123
	// 	log.Println(c.Query("v"))         // 1.0
	// 	log.Println(c.Cookies("session")) // ""

	// 	// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
	// 	var (
	// 		mt  int
	// 		msg []byte
	// 		err error
	// 	)
	// 	for {
	// 		if mt, msg, err = c.ReadMessage(); err != nil {
	// 			log.Println("read:", err)
	// 			break
	// 		}
	// 		log.Printf("recv: %s", msg)

	// 		if err = c.WriteMessage(mt, msg); err != nil {
	// 			log.Println("write:", err)
	// 			break
	// 		}
	// 	}

	// }))

	log.Fatal(app.Listen(":3000"))
	// Access the websocket server: ws://localhost:3000/ws/123?v=1.0
	// https://www.websocket.org/echo.html
}
