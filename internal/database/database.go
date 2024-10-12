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

func (u *User) Get(uid int64) error {
	row := DB.QueryRow("SELECT * FROM users WHERE uid=?", uid)
	err := row.Scan(&u.ID, &u.Username, &u.Password, &u.SessionID)
	return err
}

func (u *User) GetUserInfo(username string) error {
	row := DB.QueryRow("SELECT * FROM users WHERE username=?", username)
	err := row.Scan(&u.ID, &u.Username, &u.Password)
	return err
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
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS "users" (
	"id"	INTEGER NOT NULL,
	"username"	string NOT NULL UNIQUE,
	"password"	string NOT NULL,
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE TABLE IF NOT EXISTS "shorturls" (
	"id" INTEGER NOT NULL,
	"url" string NOT NULL,
	"short_string" string NOT NULL,
	"user_id" INTEGER,
	FOREIGN KEY ("user_id") REFERENCES users("id")
);`)
	errorPanic(err)
}

func errorPanic(err error) {
	if err != nil {
		panic(err)
	}
}
