package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	emailverifier "github.com/AfterShip/email-verifier"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sparonov/GamesBetBackend/sessionmanager"
	"github.com/sparonov/GamesBetBackend/user"
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
