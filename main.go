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
	"github.com/sparonov/GamesBetBackend/websocket"
)

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
	http.HandleFunc("/getChatHistory", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.GetChatHistoryHandler(db)))
	http.HandleFunc("/saveMessToChatHistory", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.SaveMessToChatHistoryHandler(db)))
	http.HandleFunc("/updateCoins", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.UpdateCoinsHandler(db)))
	// http.HandleFunc("/getCoins/{userEmail}", ...) (will be used in every game to show all coins of the user)
	http.HandleFunc("/getFriends/{userEmail}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.GetFriendsHandler(db)))
	http.HandleFunc("/getPotentialNewFriends/{userEmail}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.PotentialNewFriendsHandler(db)))
	http.HandleFunc("/getFriendInvites/{userEmail}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.GetFriendInvitesHandler(db)))
	http.HandleFunc("/inviteFriend/{userEmail}/{friendEmail}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.InviteFriendHandler(db)))
	http.HandleFunc("/handleInvite/{userEmail}/{friendEmail}/{type}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.CompleteInviteHandler(db)))
	http.HandleFunc("/addToScheduler", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.ScheduleGameHandler(db)))
	http.HandleFunc("/getAllScheduledGames/{userEmail}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.GetScheduledGamesHandler(db))) //gets all of the scheduled games of a user, that haven't expired (startDate + 1(tommorow day) > time.Now()), the expired ones get removed
	http.HandleFunc("/removeFromScheduler/{gameID}", cors.CORSMiddleware(config.CORSAllowedOrigins, handlers.RemoveScheduledGameHandler(db)))
	http.HandleFunc("/ws", websocket.WebSocketHandler)

	go websocket.HandleMessages()

	log.Printf("started listening on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), context.ClearHandler(http.DefaultServeMux)))

	defer db.Close()
}
