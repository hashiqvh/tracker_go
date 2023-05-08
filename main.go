package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

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
	// create log file
	logFile, err := os.Create("socket.log")
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	// create logger instance
	logger := log.New(logFile, "", log.Ldate|log.Ltime)

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Server is running")
	})

	// handle WebSocket connections
	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// add new client to the clients map
		clients[c] = true
		defer delete(clients, c)

		logger.Printf("New client connected: %s", c.RemoteAddr())

		// listen for incoming messages from the client
		for {
			// read incoming WebSocket messages
			_, message, err := c.ReadMessage()
			if err != nil {
				logger.Println(err)
				break
			}

			// log incoming message
			logger.Printf("Received message from client %s: %s", c.RemoteAddr(), string(message))

			// handle tracking data
			var pos Position
			err = json.Unmarshal(message, &pos)
			if err != nil {
				logger.Printf("Client error: %s", err)
				continue
			}

			// broadcast tracking data to all connected clients
			broadcast <- pos
		}

		logger.Printf("Client disconnected: %s", c.RemoteAddr())
	}))

	// broadcast tracking data to all connected clients
	go func() {
		for {
			pos := <-broadcast
			data, err := json.Marshal(pos)
			if err != nil {
				logger.Println(err)
				continue
			}

			for client := range clients {
				err = client.WriteMessage(websocket.TextMessage, data)
				if err != nil {
					logger.Println(err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}()

	app.Listen(":3000")

	// wait for server to shut down before closing log file
	time.Sleep(time.Second)
}
