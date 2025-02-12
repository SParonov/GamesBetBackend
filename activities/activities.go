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

	var hasGame3 bool
	var hasGame4 bool

	row:=db.QueryRow("SELECT * FROM UserGamesInfo WHERE Email=? AND Game3_Unlocked=1", email)
	if(row.Scan()==sql.ErrNoRows){
		hasGame3=false
	}else{
		hasGame3=true
	}

	row=db.QueryRow("SELECT * FROM UserGamesInfo WHERE Email=? AND Game4_Unlocked=1", email)
	if(row.Scan()==sql.ErrNoRows){
		hasGame4=false
	}else{
		hasGame4=true
	}

	if(!hasGame3&&!hasGame4){
		a.Game="Game1"
	}else if(hasGame3&&!hasGame4){
		game:=rand.Intn(2)
		switch game{
		case 0:
			a.Game="Game1"
		case 1:
			a.Game="Game3"
		}
	}else if(!hasGame3&&hasGame4){
		game:=rand.Intn(2)
		switch game{
		case 0:
			a.Game="Game1"
		case 1:
			a.Game="Game4"
		}
	}else{
		game:=rand.Intn(3)

		switch game{
		case 0:
			a.Game="Game1"
		case 1:
			a.Game="Game3"
		case 2:
			a.Game="Game4"
		}
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

	time.Sleep(1*time.Minute)
	AddActivityForEveryUser(db)
}