package activities

import (
	"database/sql"
	"math/rand"
	"time"
)

type Activity struct{
	RequiredCoins int
	Game string
	Reward int
}

func AddActivityForAGivenUser(db *sql.DB, email string){
	var a Activity;
	// range in [0, 10]
	a.RequiredCoins = 10 + rand.Intn(11)
	game:=rand.Intn(3)
	switch game{
	case 0:
		a.Game="Game1"
	case 1:
		a.Game="Game3"
	case 2:
		a.Game="Game4"
	}

	a.Reward = 50 + rand.Intn(51)

	db.Exec("INSERT INTO Activities VALUES (?, ?, ?, ?)", email, a.RequiredCoins, a.Game, a.Reward)
}

func AddActivityForEveryUser(db *sql.DB){
	row, _:=db.Query("SELECT Email FROM UserRegisterInfo")

	for row.Next(){
		var temp string
		row.Scan(&temp)
		go AddActivityForAGivenUser(db, temp)
	}

	time.Sleep(24*time.Hour)
	AddActivityForEveryUser(db)
}