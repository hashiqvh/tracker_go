package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Position struct {
	Latitude  float64 `json:"lat"`
	Longitude float64 `json:"lng"`
	BusNumber int     `json:"bus_number"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Position)

func main() {
	app := fiber.New()

app.Get("/", func(c *fiber.Ctx) error {
    return c.Send("<h1>Server is running</h1>")
})

	// handle WebSocket connections
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// add new client to the clients map
		clients[c] = true
		defer delete(clients, c)

		fmt.Println("New client connected:", c.RemoteAddr())

		// listen for incoming messages from the client
		for {
			// read incoming WebSocket messages
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}

			// log incoming message
			log.Printf("Received message from client %s: %s", c.RemoteAddr(), string(message))

			// handle tracking data
			var pos Position
			err = json.Unmarshal(message, &pos)
			if err != nil {
				fmt.Println("Client error:", err)
				log.Println(err)
				continue
			}

			// broadcast tracking data to all connected clients
			broadcast <- pos
		}

		fmt.Println("Client disconnected:", c.RemoteAddr())
	}))

	// broadcast tracking data to all connected clients
	go func() {
		for {
			pos := <-broadcast
			data, err := json.Marshal(pos)
			if err != nil {
				log.Println(err)
				continue
			}

			for client := range clients {
				err = client.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					log.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}()

	app.Listen(":3000")
}
