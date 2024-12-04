package database

import(
	"log"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sparonov/GamesBetBackend/config"
)

func Connect() *sql.DB{
	config:=config.Config()
	db, err := sql.Open("mysql", config.ConnectionString)

	if err != nil {
		log.Fatalf("cannot open db engine %v", err)
	}
	return db
}