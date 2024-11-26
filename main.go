package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/denisenkom/go-mssqldb"
	cors "github.com/sparonov/GamesBetBackend/cors"
	"github.com/tkanos/gonfig"
)

type configuration struct {
	ConnectionString   string
	Port               int
	Url                string
	CORSAllowedOrigins []string
}

type user struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"passwordConfirm"`
	Email           string `json:"email"`
}

func signupHandler(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var user user

		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if user.Password != user.PasswordConfirm {
			http.Error(w, "Different password added on confirm password", http.StatusBadRequest)
			return
		}

		query := `USE GamesBet
				INSERT INTO Users (Username, Email, Password)
				VALUES 	(@p1, @p2, @p3)`

		err = db.QueryRow(query, user.Username, user.Email, user.Password).Err()

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("New user registered:")
		fmt.Printf("Username %v, Password: %v, Email: %v\n", user.Username, user.Password, user.Email)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	config := configuration{}
	err := gonfig.GetConf("config.json", &config)

	if err != nil {
		log.Fatalf("cannot read the config file. GetConf returned: %s", err.Error())
	}

	db, err := sql.Open("sqlserver", config.ConnectionString)

	if err != nil {
		log.Fatalf("cannot open db engine %v", err)
	}

	http.HandleFunc("/signup", cors.CORSMiddleware(config.CORSAllowedOrigins, signupHandler(db)))

	log.Printf("started listening on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))

	defer db.Close()
}
