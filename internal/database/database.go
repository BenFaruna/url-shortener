package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var DB = &sql.DB{}

type User struct {
	ID                            int64
	Username, Password, SessionID string
}

type ShortUrls struct {
	ID, UserID       int64
	URL, ShortString string
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

func (s ShortUrls) Get(shortUrl string) (string, bool) {
	var userId sql.NullInt64
	row := DB.QueryRow("SELECT * FROM shorturls WHERE short_string=?;", shortUrl)
	err := row.Scan(&s.ID, &s.URL, &s.ShortString, &userId)
	if err != nil {
		return "", false
	}
	return s.URL, true
}

func (s ShortUrls) GetAll() []URLInfo {
	urls := make([]URLInfo, 0)
	rows, err := DB.Query("SELECT * FROM shorturls;")
	if err != nil {
		return urls
	}

	for rows.Next() {
		id, userId := sql.NullInt64{}, sql.NullInt64{}
		url := URLInfo{}
		err := rows.Scan(&id, &url.URL, &url.ShortAddress, &userId)
		if err != nil {
			log.Printf("ShortUrls.GetAll: %v\n", err)
			continue
		}
		urls = append(urls, url)
	}
	return urls
}

func (s ShortUrls) Add(url, shortLink string) (string, error) {
	if url == "" || shortLink == "" {
		return "", ErrorEmptyString
	}

	if s.IsExists(shortLink) {
		return "", ErrorDuplicateShortString
	}

	var shortUrlTemp string

	existingLink, exists := s.SearchURL(url)
	if exists {
		shortUrlTemp = existingLink
	} else {
		shortUrlTemp = shortLink
	}
	_, err := DB.Exec("INSERT INTO shorturls (url, short_string, user_id) VALUES(?, ?, ?);",
		url, shortUrlTemp, nil)
	if err != nil {
		return "", err
	}
	return shortUrlTemp, nil
}

func (s ShortUrls) IsExists(shortLink string) bool {
	var userId sql.NullInt64
	row := DB.QueryRow("SELECT * FROM shorturls WHERE short_string LIKE ? LIMIT 1", shortLink)
	err := row.Scan(&s.ID, &s.URL, &s.ShortString, &userId)
	if err != nil {
		return false
	}
	return true
}

func (s ShortUrls) SearchURL(url string) (string, bool) {
	var userId sql.NullInt64
	row := DB.QueryRow("SELECT * FROM shorturls WHERE url LIKE ? LIMIT 1", url)
	err := row.Scan(&s.ID, &s.URL, &s.ShortString, &userId)
	if err != nil {
		return "", false
	}
	return s.ShortString, true
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
	PRIMARY KEY("id" AUTOINCREMENT)
	FOREIGN KEY ("user_id") REFERENCES users("id")
);`)
	errorPanic(err)
}

func errorPanic(err error) {
	if err != nil {
		panic(err)
	}
}
