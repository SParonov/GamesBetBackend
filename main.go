package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/gorilla/context"
	"github.com/sparonov/GamesBetBackend/config"
	cors "github.com/sparonov/GamesBetBackend/cors"
	"github.com/sparonov/GamesBetBackend/database"
	"github.com/sparonov/GamesBetBackend/handlers"
)




func main() {
	config := config.Config()

	db := database.Connect()

	http.HandleFunc("/signup", cors.CORSMiddleware(config.CORSAllowedOrigins,handlers.SignupHandler(db)))
	http.HandleFunc("/login", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.LoginHandler(db)))
	http.HandleFunc("/games_hub", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.Auth(handlers.LoginHandler(db))))



	log.Printf("started listening on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), context.ClearHandler(http.DefaultServeMux)))

	defer db.Close()
}
