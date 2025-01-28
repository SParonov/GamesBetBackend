package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	emailverifier "github.com/AfterShip/email-verifier"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/sparonov/GamesBetBackend/sessionmanager"
	"github.com/sparonov/GamesBetBackend/user"
	"github.com/sparonov/GamesBetBackend/websocket"
)

func SignupHandler(db *sql.DB, sessionManager *sessionmanager.SessionManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var user user.User

		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if user.Password != user.PasswordConfirm {
			http.Error(w, "Different password added on confirm password", http.StatusBadRequest)
			return
		}

		verifier := emailverifier.NewVerifier()
		res, _ := verifier.Verify(user.Email)
		if !res.Syntax.Valid {
			http.Error(w, "Invalid email syntax", http.StatusBadRequest)
			return
		}
		var Id int32
		Id = rand.Int31()
		//guarantee the id is unique
		checkForDublicateId := "SELECT * FROM UserData.UserRegisterInfo WHERE Id=?;"
		row := db.QueryRow(checkForDublicateId, Id)
		err = row.Scan(&Id)
		for err != sql.ErrNoRows {
			Id = rand.Int31()
			row = db.QueryRow(checkForDublicateId, Id)
			err = row.Scan(&Id)
		}

		query := "INSERT INTO UserRegisterInfo (Id, Username, Email, Password) VALUES (?, ?, ?, ?)"
		err = db.QueryRow(query, Id, user.Username, user.Email, user.Password).Err()

		if err != nil {
			fmt.Print(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		query = "INSERT INTO UserGamesInfo (Email, Coins, Game1_Unlocked, Game1_Highscore, Game2_Unlocked, Game2_Highscore, Game3_Unlocked, Game3_Highscore, Game4_Unlocked, Game4_Highscore, Game5_Unlocked, Game5_Highscore, Game6_Unlocked, Game6_Highscore, Game7_Unlocked, Game7_Highscore, Game8_Unlocked, Game8_Highscore) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
		err = db.QueryRow(query, user.Email, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0).Err()

		if err != nil {
			fmt.Print(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("New user registered:")
		fmt.Printf("Username %v, Password: %v, Email: %v\n", user.Username, user.Password, user.Email)

		ctx := context.Background()
		sessionID, err := sessionManager.CreateSession(ctx, user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sessionData, err := sessionManager.GetSessionData(sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sessionDataJSON, err := json.Marshal(sessionData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sessionDataJSON))
	}
}

func LoginHandler(db *sql.DB, sessionManager *sessionmanager.SessionManager) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var user user.User

		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		checkIfAnAccountExists := "SELECT * FROM Userdata.userRegisterInfo WHERE Email=? AND Password=?;"
		row := db.QueryRow(checkIfAnAccountExists, user.Email, user.Password)
		err = row.Scan(&user)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email and password combination", http.StatusBadRequest)
			return
		}

		fmt.Println("User logged:")
		fmt.Printf("Email: %v, Password: %v\n", user.Email, user.Password)

		ctx := context.Background()
		sessionID, err := sessionManager.CreateSession(ctx, user.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sessionData, err := sessionManager.GetSessionData(sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		sessionDataJSON, err := json.Marshal(sessionData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sessionDataJSON))
	}
}

type GameData struct {
	Coins     int32 `json:"coins"`
	Highscore int32 `json:"highscore"`
}

func UpdateUserGameDataHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		gameHigscore := r.PathValue("gameID") + "_Highscore"
		userEmail := r.PathValue("userEmail")

		getPrevCoinsAndHighscore := fmt.Sprintf("SELECT Coins, %s FROM UserGamesInfo WHERE Email=?", gameHigscore)

		var prevCoins int32
		var prevHighscore int32

		row := db.QueryRow(getPrevCoinsAndHighscore, userEmail)
		if err := row.Scan(&prevCoins, &prevHighscore); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No previous data found for the user", http.StatusNotFound)
			} else {
				http.Error(w, "Error retrieving previous game data: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		var newGameData GameData

		err := json.NewDecoder(r.Body).Decode(&newGameData)

		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		query := fmt.Sprintf("UPDATE UserGamesInfo SET coins=?, %s=? WHERE Email=?", gameHigscore)

		_, err = db.Exec(query, prevCoins+newGameData.Coins, int32(math.Max(float64(prevHighscore), float64(newGameData.Highscore))), userEmail)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type HighscoreResponse struct {
	Highscore int32 `json:"highscore"`
}

func GetUserGameDataHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		gameHigscore := r.PathValue("gameID") + "_Highscore"
		userEmail := r.PathValue("userEmail")

		getPrevCoinsAndHighscore := fmt.Sprintf("SELECT %s FROM UserGamesInfo WHERE Email=?", gameHigscore)

		var highscore int32

		row := db.QueryRow(getPrevCoinsAndHighscore, userEmail)
		if err := row.Scan(&highscore); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No previous data found for the user", http.StatusNotFound)
			} else {
				http.Error(w, "Error retrieving previous game data: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := HighscoreResponse{
			Highscore: highscore,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type ChatHistoryResponse struct {
	ChatHistory []websocket.Message
}

func GetChatHistoryHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		defer r.Body.Close()

		rows, err := db.Query("SELECT id, email, timestamp, text FROM ChatHistory")

		if err != nil {
			fmt.Println(err)
			return
		}

		defer rows.Close()

		var message websocket.Message

		messages := []websocket.Message{}

		for rows.Next() {
			err = rows.Scan(&message.ID, &message.Email, &message.Timestamp, &message.Text)

			if err != nil {
				fmt.Println(err)
				return
			}

			messages = append(messages, message)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		response := ChatHistoryResponse{
			ChatHistory: messages,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func SaveMessToChatHistoryHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var message websocket.Message

		err := json.NewDecoder(r.Body).Decode(&message)

		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		query := "INSERT INTO ChatHistory (Id, Text, Email, Timestamp) VALUES (?, ?, ?, ?)"
		err = db.QueryRow(query, message.ID, message.Text, message.Email, message.Timestamp).Err()

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type CoinsQuery struct {
	Coins int32  `json:"coins"`
	Email string `json:"email"`
}

type CoinsQueryResponse struct {
	Coins       int32
	EnoughCoins bool
}

func UpdateCoinsHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var coinsQuery CoinsQuery

		err := json.NewDecoder(r.Body).Decode(&coinsQuery)

		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		query := "SELECT Coins FROM UserGamesInfo WHERE Email=?;"
		row := db.QueryRow(query, coinsQuery.Email)

		var availableCoins int32

		if err := row.Scan(&availableCoins); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "No previous data found for the user", http.StatusNotFound)
			} else {
				http.Error(w, "Error retrieving previous game data: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		enoughCoins := true
		var response CoinsQueryResponse

		if availableCoins < coinsQuery.Coins {
			enoughCoins = false
			response = CoinsQueryResponse{
				EnoughCoins: enoughCoins,
				Coins:       0,
			}
		} else {
			removeCoinsFromAccount := "UPDATE UserGamesInfo SET Coins=? WHERE Email=?"

			_, err = db.Exec(removeCoinsFromAccount, availableCoins-coinsQuery.Coins, coinsQuery.Email)

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			response = CoinsQueryResponse{
				EnoughCoins: enoughCoins,
				Coins:       coinsQuery.Coins,
			}
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response: "+err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
	}
}

func PotentialNewFriendsHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		userEmail := r.PathValue("userEmail")

		query := `
			SELECT Email
			FROM UserRegisterInfo
			LEFT JOIN Friendships f 
			ON (Email = f.FriendEmail AND f.UserEmail = ?) 
			WHERE (f.status IS NULL OR f.status = 'declined')
			AND Email != ?;`

		rows, err := db.Query(query, userEmail, userEmail)
		if err != nil {
			http.Error(w, "Failed to execute query: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var potentialFriends []string

		for rows.Next() {
			var email string
			if err := rows.Scan(&email); err != nil {
				http.Error(w, "Error reading data: "+err.Error(), http.StatusInternalServerError)
				return
			}
			potentialFriends = append(potentialFriends, email)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating over rows: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := struct {
			PotentialFriends []string `json:"potentialFriends"`
		}{
			PotentialFriends: potentialFriends,
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}
}

func InviteFriendHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		userEmail := r.PathValue("userEmail")
		friendEmail := r.PathValue("friendEmail")

		if userEmail == "" || friendEmail == "" {
			http.Error(w, "Missing userEmail or friendEmail in the URL", http.StatusBadRequest)
			return
		}

		query := `
			SELECT COUNT(*) 
			FROM userdata.friendships 
			WHERE (UserEmail = ? AND FriendEmail = ?) 
			   OR (UserEmail = ? AND FriendEmail = ?);`

		var count int
		err := db.QueryRow(query, userEmail, friendEmail, friendEmail, userEmail).Scan(&count)
		if err != nil {
			http.Error(w, "Error checking friendship status", http.StatusInternalServerError)
			return
		}

		if count > 0 {
			updateQuery := `
				UPDATE friendships 
				SET status = 'pending' 
				WHERE (UserEmail = ? AND FriendEmail = ?) 
				   OR (UserEmail = ? AND FriendEmail = ?);`

			_, err := db.Exec(updateQuery, userEmail, friendEmail, friendEmail, userEmail)
			if err != nil {
				http.Error(w, "Error updating friendship status", http.StatusInternalServerError)
				return
			}
		} else {
			insertQuery := `
				INSERT INTO userdata.friendships (UserEmail, FriendEmail, status)
				VALUES (?, ?, 'pending');`

			_, err := db.Exec(insertQuery, userEmail, friendEmail)
			if err != nil {
				http.Error(w, "Error inserting friendship", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func GetFriendsHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid HTTP method", http.StatusMethodNotAllowed)
			return
		}

		userEmail := r.PathValue("userEmail")

		if userEmail == "" {
			http.Error(w, "Missing userEmail in the URL", http.StatusBadRequest)
			return
		}

		query := `
			SELECT FriendEmail
			FROM Friendships 
			WHERE UserEmail = ?
			AND status = 'accepted';`

		rows, err := db.Query(query, userEmail)
		if err != nil {
			http.Error(w, "Error fetching friends", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var friends []string

		for rows.Next() {
			var friendEmail string
			if err := rows.Scan(&friendEmail); err != nil {
				http.Error(w, "Error scanning rows", http.StatusInternalServerError)
				return
			}
			friends = append(friends, friendEmail)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating over rows", http.StatusInternalServerError)
			return
		}

		response := struct {
			Friends []string `json:"friends"`
		}{
			Friends: friends,
		}

		responseJSON, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Failed to marshal response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(responseJSON)
	}
}

func GetFriendInvitesHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid HTTP method", http.StatusMethodNotAllowed)
			return
		}

		userEmail := r.PathValue("userEmail")
		if userEmail == "" {
			http.Error(w, "Missing userEmail in the URL", http.StatusBadRequest)
			return
		}

		query := `SELECT UserEmail FROM userdata.Friendships WHERE FriendEmail = ? AND Status = 'pending';`
		rows, err := db.Query(query, userEmail)
		if err != nil {
			http.Error(w, "Error retrieving friend invites", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var invites []string
		for rows.Next() {
			var inviterEmail string
			if err := rows.Scan(&inviterEmail); err != nil {
				http.Error(w, "Error scanning data", http.StatusInternalServerError)
				return
			}
			invites = append(invites, inviterEmail)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error processing rows", http.StatusInternalServerError)
			return
		}

		response := struct {
			Invites []string `json:"invites"`
		}{
			Invites: invites,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Error encoding response", http.StatusInternalServerError)
		}
	}
}

func CompleteInviteHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		userEmail := r.PathValue("userEmail")
		friendEmail := r.PathValue("friendEmail")
		action := r.PathValue("type")

		var newStatus string
		if action == "accept" {
			newStatus = "accepted"
		} else if action == "decline" {
			newStatus = "declined"
		} else {
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}

		updateQuery := `
			UPDATE Friendships 
			SET Status = ? 
			WHERE UserEmail = ? AND FriendEmail = ?
		`
		_, err := db.Exec(updateQuery, newStatus, friendEmail, userEmail)
		if err != nil {
			http.Error(w, "Error updating friendship status: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var count int
		checkQuery := `
			SELECT COUNT(*) 
			FROM userdata.Friendships 
			WHERE UserEmail = ? AND FriendEmail = ?
		`
		err = db.QueryRow(checkQuery, userEmail, friendEmail).Scan(&count)
		if err != nil {
			http.Error(w, "Error checking second row existence: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if count > 0 {
			updateQuerySecondRow := `
				UPDATE userdata.Friendships 
				SET Status = ? 
				WHERE UserEmail = ? AND FriendEmail = ?
			`
			_, err := db.Exec(updateQuerySecondRow, newStatus, userEmail, friendEmail)
			if err != nil {
				http.Error(w, "Error updating second row: "+err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			insertQuery := `
				INSERT INTO userdata.Friendships (UserEmail, FriendEmail, Status)
				VALUES (?, ?, ?)
			`
			_, err := db.Exec(insertQuery, userEmail, friendEmail, newStatus)
			if err != nil {
				http.Error(w, "Error inserting second row: "+err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

type ScheduleGameRequest struct {
	Player1   string `json:"player1"`
	Player2   string `json:"player2"`
	StartDate string `json:"startDate"`
	Game      string `json:"game"`
}

func ScheduleGameHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := uuid.New().String()

		var req ScheduleGameRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Failed to parse request", http.StatusBadRequest)
			return
		}

		query := `INSERT INTO scheduler (id, player1, player2, startDate, game) VALUES (?, ?, ?, ?, ?)`
		_, err = db.Exec(query, id, req.Player1, req.Player2, req.StartDate, req.Game)
		if err != nil {
			http.Error(w, "Failed to schedule game", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

type Game struct {
	ID        string `json:"id"`
	Player1   string `json:"player1"`
	Player2   string `json:"player2"`
	StartDate string `json:"startDate"`
	Game      string `json:"game"`
}

func removeScheduledGame(db *sql.DB, gameID string) error {
	query := `
		SET SQL_SAFE_UPDATES = 0;
		DELETE FROM scheduler WHERE id = ?;
	`

	_, err := db.Exec(query, gameID)
	return err
}

func GetScheduledGamesHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		userEmail := r.PathValue("userEmail")
		if userEmail == "" {
			http.Error(w, "User email is required", http.StatusBadRequest)
			return
		}

		query := `
			SELECT id, player1, player2, startDate, game 
			FROM scheduler 
			WHERE player1 = ? OR player2 = ?`

		rows, err := db.Query(query, userEmail, userEmail)
		if err != nil {
			http.Error(w, "Failed to fetch scheduled games", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var games []Game

		for rows.Next() {
			var game Game
			err := rows.Scan(&game.ID, &game.Player1, &game.Player2, &game.StartDate, &game.Game)
			if err != nil {
				http.Error(w, "Failed to read game data", http.StatusInternalServerError)
				return
			}

			gameStartDate, err := time.Parse("2006-01-02T15:04", game.StartDate)
			if err != nil {
				http.Error(w, "Invalid date format", http.StatusInternalServerError)
				return
			}

			if gameStartDate.Add(24 * time.Hour).Before(time.Now()) {
				err := removeScheduledGame(db, game.ID)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to remove expired game: %v", err), http.StatusInternalServerError)
					return
				}
				continue
			}

			games = append(games, game)
		}

		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating rows", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(struct {
			ScheduledGames []Game `json:"scheduledGames"`
		}{ScheduledGames: games}); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}

func RemoveScheduledGameHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		gameID := r.PathValue("gameID")

		err := removeScheduledGame(db, gameID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to remove scheduled game: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
