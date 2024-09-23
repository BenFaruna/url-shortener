package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

var DB = &sql.DB{}

type User struct {
	ID                            int64
	Username, Password, SessionID string
}

func (u *User) Get(uid int64) {
	row := DB.QueryRow("SELECT * FROM users WHERE uid=?", uid)
	row.Scan(&u.ID, &u.Username, &u.Password, &u.SessionID)
}

func (u *User) Add() error {
	if u.Username == "" || u.Password == "" {
		return fmt.Errorf("User.Add(): Username or Password empty")
	}
	r, err := DB.Exec("INSERT INTO users (username, password) VALUES(?, ?);", u.Username, u.Password)
	if err != nil {
		return err
	}
	id, _ := r.LastInsertId()
	u.ID = id
	return nil
}

func init() {
	db, err := sql.Open("sqlite3", "app.db")
	errorPanic(err)
	DB = db

	err = db.Ping()
	errorPanic(err)
	db.Exec("CREATE TABLE users (uid integer PRIMARY KEY AUTOINCREMENT, username string, password string, sessionid string);")
}

func errorPanic(err error) {
	if err != nil {
		panic(err)
	}
}
