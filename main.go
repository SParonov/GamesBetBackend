package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"math/rand"
	_ "github.com/go-sql-driver/mysql"
	cors "github.com/sparonov/GamesBetBackend/cors"
	emailverifier "github.com/AfterShip/email-verifier"
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
	var Id int32
	Id=rand.Int31()
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
		verifier:=emailverifier.NewVerifier()
		res, _:= verifier.Verify(user.Email)
		if !res.Syntax.Valid{
			http.Error(w, "Invalid email syntax", http.StatusBadRequest)
			return
		}
		//guarantee the id is unique
		checkForDublicateId:="SELECT * FROM UserData.UserRegisterInfo WHERE Id=?;"
		row:=db.QueryRow(checkForDublicateId, Id)
		err=row.Scan(&Id)
		for err!=sql.ErrNoRows{
			Id=rand.Int31()
			row=db.QueryRow(checkForDublicateId, Id)
			err=row.Scan(&Id)
		}
		var insertStatement *sql.Stmt
	
		insertStatement, err= db.Prepare("INSERT INTO UserRegisterInfo (Id, Username, Email, Password) VALUES (?, ?, ?, ?);")

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

func main() {
	config := configuration{}
	err := gonfig.GetConf("config.json", &config)

	if err != nil {
		log.Fatalf("cannot read the config file. GetConf returned: %s", err.Error())
	}

	db, err := sql.Open("mysql", config.ConnectionString)

	if err != nil {
		log.Fatalf("cannot open db engine %v", err)
	}

	http.HandleFunc("/signup", cors.CORSMiddleware(config.CORSAllowedOrigins, signupHandler(db)))

	log.Printf("started listening on port %d\n", config.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))

	defer db.Close()
}
