package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/sparonov/GamesBetBackend/config"
	cors "github.com/sparonov/GamesBetBackend/cors"
	"github.com/sparonov/GamesBetBackend/database"
	"github.com/sparonov/GamesBetBackend/handlers"
	"github.com/sparonov/GamesBetBackend/sessionmanager"
)

func main() {
	config := config.Config()

	db := database.Connect()

	sessionManager := sessionmanager.New(db, time.Minute*time.Duration(10))

	http.HandleFunc("/signup", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.SignupHandler(db, sessionManager)))
	http.HandleFunc("/login", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.LoginHandler(db, sessionManager)))

	log.Printf("started listening on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), context.ClearHandler(http.DefaultServeMux)))

	defer db.Close()
}
