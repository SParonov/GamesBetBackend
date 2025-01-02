package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"github.com/sparonov/GamesBetBackend/config"
	cors "github.com/sparonov/GamesBetBackend/cors"
	"github.com/sparonov/GamesBetBackend/database"
	"github.com/sparonov/GamesBetBackend/handlers"
	"github.com/sparonov/GamesBetBackend/sessionmanager"
)

// Upgrader to upgrade HTTP connections to WebSocket connections
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for simplicity
		return true
	},
}

// A simple message struct for the chat
type Message struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

// Store connected clients and broadcast channel
var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

// Handle WebSocket connections
func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatalf("Error upgrading connection: %v", err)
	}
	defer ws.Close()

	// Register the new client
	clients[ws] = true

	// Listen for messages from this client
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			delete(clients, ws)
			break
		}
		// Send the message to the broadcast channel
		broadcast <- msg
	}
}

// Handle broadcasting messages to all clients
func handleMessages() {
	for {
		// Get the next message from the broadcast channel
		msg := <-broadcast

		// Send the message to all connected clients
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	config := config.Config()

	db := database.Connect()

	sessionManager := sessionmanager.New(db, time.Minute*time.Duration(10))

	http.HandleFunc("/signup", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.SignupHandler(db, sessionManager)))
	http.HandleFunc("/login", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.LoginHandler(db, sessionManager)))
	// http.HandleFunc("/getGamesData", ...); returns data from all games from all users (will be used in scoreboard)
	// http.HandleFunc("/getGamesData/{gameID}", ...); returns data from a game from all users (will be used in scoreboard)
	http.HandleFunc("/updateGamesData/{gameID}/{userEmail}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.UpdateUserGameDataHandler(db)))
	http.HandleFunc("/getGamesData/{gameID}/{userEmail}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.GetUserGameDataHandler(db)))
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	log.Printf("started listening on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), context.ClearHandler(http.DefaultServeMux)))

	defer db.Close()
}
