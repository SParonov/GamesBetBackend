package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	emailverifier "github.com/AfterShip/email-verifier"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"github.com/sparonov/GamesBetBackend/user"
)

var store = sessions.NewCookieStore([]byte("session-key"))

func SignupHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	var Id int32
	Id = rand.Int31()
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
		//guarantee the id is unique
		checkForDublicateId := "SELECT * FROM UserData.UserRegisterInfo WHERE Id=?;"
		row := db.QueryRow(checkForDublicateId, Id)
		err = row.Scan(&Id)
		for err != sql.ErrNoRows {
			Id = rand.Int31()
			row = db.QueryRow(checkForDublicateId, Id)
			err = row.Scan(&Id)
		}
		var insertStatement *sql.Stmt

		insertStatement, err = db.Prepare("INSERT INTO UserRegisterInfo (Id, Username, Email, Password) VALUES (?, ?, ?, ?);")

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer insertStatement.Close()
		insertStatement.Exec(Id, user.Username, user.Email, user.Password)

		fmt.Println("New user registered:")
		fmt.Printf("Username %v, Password: %v, Email: %v\n", user.Username, user.Password, user.Email)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func LoginHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
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
		checkIfAnAccountExists := "SELECT * FROM UserData.UserRegisterInfo WHERE Email=? AND Password=?;"
		row := db.QueryRow(checkIfAnAccountExists, user.Email, user.Password)
		err = row.Scan(&user)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email and password combination", http.StatusBadRequest)
			return
		}
		session, _ := store.Get(r, "session")
		session.Values["userEmail"] = user.Email
		err = session.Save(r, w)
		if err != nil {
			fmt.Print(err)
		}

		w.WriteHeader(http.StatusOK)
	}
}

func Auth(HandlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		_, ok := session.Values["userEmail"]
		if !ok {
			http.Error(w, "Not logged in", http.StatusBadRequest)
			return
		}
		HandlerFunc.ServeHTTP(w, r)
	}
}

func Games_hubHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("logged in")
}
