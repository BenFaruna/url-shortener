package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
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
	if userId.Valid {
		if err = userId.Scan(&s.UserID); err != nil {
			log.Printf("ShortUrls.Get: %v\n", err)
		}
	}
	return s.URL, true
}

func (s ShortUrls) GetAll(userId int64) []URLInfo {
	var query string
	if userId != 0 {
		query = "SELECT * FROM shorturls WHERE user_id=?;"
	} else {
		query = "SELECT * FROM shorturls;"
	}
	urls := make([]URLInfo, 0)
	rows, err := DB.Query(query, userId)
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
		if id.Valid {
			url.UrlId = id.Int64
		}
		if userId.Valid {
			url.UserId = userId.Int64
		}
		urls = append(urls, url)
	}
	return urls
}

func (s ShortUrls) IsUserURL(userId int64, shortLink string) bool {
	if userId == 0 {
		return false
	}
	r, err := DB.Query("SELECT * FROM shorturls WHERE user_id=? AND short_string LIKE ?", userId, shortLink)
	if err != nil {
		log.Print(err)
		return false
	}
	ok := r.Next()
	return ok
}

func (s ShortUrls) Add(userId int64, url, shortLink string) (string, error) {
	if url == "" || shortLink == "" {
		return "", ErrorEmptyString
	}

	if s.IsExists(shortLink) {
		return "", ErrorDuplicateShortString
	}

	var shortUrlTemp string

	existingLink, urlUserId, exists := s.SearchURL(url)
	if exists {
		shortUrlTemp = existingLink
	} else {
		shortUrlTemp = shortLink
	}

	if userId != urlUserId || urlUserId != 0 {
		_, err := DB.Exec("INSERT INTO shorturls (url, short_string, user_id) VALUES(?, ?, ?);",
			url, shortUrlTemp, userId)
		if err != nil {
			return "", err
		}
	}

	return shortUrlTemp, nil
}

func (s ShortUrls) Delete(userId int64, shortLink string) error {
	_, err := DB.Exec("DELETE FROM shorturls WHERE short_string LIKE ? AND user_id = ?", shortLink, userId)
	if err != nil {
		return err
	}
	return nil
}

func (s ShortUrls) IsExists(shortLink string) bool {
	var userId sql.NullInt64
	row := DB.QueryRow("SELECT * FROM shorturls WHERE short_string LIKE ? LIMIT 1", shortLink)
	err := row.Scan(&s.ID, &s.URL, &s.ShortString, &userId)
	return err == nil
}

func (s ShortUrls) SearchURL(url string) (string, int64, bool) {
	var userId sql.NullInt64
	row := DB.QueryRow("SELECT * FROM shorturls WHERE url LIKE ? LIMIT 1", url)
	err := row.Scan(&s.ID, &s.URL, &s.ShortString, &userId)
	if err != nil {
		return "", userId.Int64, false
	}
	return s.ShortString, userId.Int64, true
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
