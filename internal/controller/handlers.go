package controller

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BenFaruna/url-shortener/internal/database"
	"github.com/BenFaruna/url-shortener/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type FormToken struct {
	Token string
}

// HomeHandler accept requests to the home route and provide responses are redirection for short routes
func HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderer, err := NewRenderer()
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("route %q does not exists", r.URL.Path))
		}
		switch r.URL.Path {
		case "/":
			data := model.Db.GetAll()
			renderer.Render(w, "index.gohtml", data)
			// fmt.Fprint(w, "Hello World")
			return
		default:
			shortID := strings.TrimPrefix(r.URL.Path, "/")
			url, ok := model.Db.Get(shortID)
			if !ok {
				errorHandler(w, r, http.StatusNotFound, fmt.Sprintf("route %q does not exists", r.URL.Path))
				return
			}

			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}
	}
}

func ShortenHandler(shortStringFunc func() string) http.Handler {
	return Post(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/shorten" {
			errorHandler(w, r, http.StatusNotFound, fmt.Sprintf("route %q does not exists", r.URL.Path))
			return
		}
		var data model.Body
		json.NewDecoder(r.Body).Decode(&data)

		shortenedURL := shortStringFunc()

		shortenedURL, err := model.Db.Add(data.URL, shortenedURL)

		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, err.Error())
			return
		}

		w.WriteHeader(201)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&model.StatusMessage{
			Message: "url shortened",
			Data:    r.URL.Hostname() + shortenedURL,
		})
	}))
}

func GetFullAddressHandler() http.Handler {

	return Get(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortAddress := strings.TrimPrefix(r.URL.Path, "/address/")

		url, ok := model.Db.Get(shortAddress)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, "address does not exist")
			return
		}

		json.NewEncoder(w).Encode(model.StatusMessage{
			Data:    url,
			Message: "address found",
		})
	}))
}

func LoginHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderer, err := NewRenderer()
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
		}
		if r.Method == http.MethodGet {
			curtime := time.Now().Unix()
			h := md5.New()
			io.WriteString(h, strconv.FormatInt(curtime, 10))
			token := fmt.Sprintf("%x", h.Sum(nil))
			if err = renderer.Render(w, "login.gohtml", FormToken{token}); err != nil {
				errorHandler(w, r, http.StatusInternalServerError, "")
				return
			}
		} else if r.Method == http.MethodPost {
			r.ParseForm()

			if token := r.Form.Get("token"); token != "" {
				// fmt.Println("Token", token)
			} else {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}

			username := template.HTMLEscapeString(r.Form.Get("username"))
			password := template.HTMLEscapeString(r.Form.Get("password"))

			fmt.Println("Username", username)
			fmt.Println("Password", password)

			if !validateUsername(username) {
				errorHandler(w, r, http.StatusBadRequest, "invalid username")
				return
			}
			if !validatePassword(password) {
				errorHandler(w, r, http.StatusBadRequest, "invalid password")
				return
			}
			// TODO: check password before login
			http.Redirect(w, r, "/", http.StatusPermanentRedirect)

		} else {
			errorHandler(w, r, http.StatusMethodNotAllowed, fmt.Sprintf("%s method not allowed", r.Method))
		}
	})
}

func SignupHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			curtime := time.Now().Unix()
			h := md5.New()
			io.WriteString(h, strconv.FormatInt(curtime, 10))
			token := fmt.Sprintf("%x", h.Sum(nil))

			renderer, err := NewRenderer()
			if err != nil {
				errorHandler(w, r, http.StatusInternalServerError, "")
			}

			renderer.Render(w, "signup.gohtml", FormToken{token})
			return
		} else if r.Method == http.MethodPost {
			r.ParseForm()

			if token := r.Form.Get("token"); token != "" {
				// fmt.Println("Token", token)
			} else {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}

			username := r.Form.Get("username")
			password := r.Form.Get("password")

			if username == "" || password == "" {
				errorHandler(w, r, http.StatusBadRequest, "")
				return
			}

			password, err := generatePasswordHash(password)
			if err != nil {
				errorHandler(w, r, http.StatusBadRequest, "invalid password")
				return
			}

			user := database.User{Username: username, Password: password}
			if err := user.Add(); err != nil {
				errorHandler(w, r, http.StatusInternalServerError, "User not added")
				return
			}

			// TODO: ensure username is unique and add entry to database
			http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
			return
		}
	})
}

func validateUsername(s string) bool {
	match, _ := regexp.MatchString("^[a-z0-_.]{3,15}$", s)
	// fmt.Println("Username", match)
	return match
}

func validatePassword(s string) bool {
	match, _ := regexp.MatchString("^[A-z0-9_!@#$_%^&*.?()-=+ ]*$", s)
	// fmt.Println("password", match)
	return match
}

func generatePasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", fmt.Errorf("invalid password")
	}
	return string(hashedPassword), nil
}

func errorHandler(w http.ResponseWriter, _ *http.Request, status int, errorMessage string) {
	w.WriteHeader(status)
	// w.Header().Set("Content-Type", "application/json")
	if status == http.StatusNotFound {
		fmt.Fprint(w, errorMessage)
	}
}

func GenerateShortString() string {
	output := ""

	for i := 0; i < 6; i++ {
		n := rand.Intn(51)
		output += string(model.Characters[n])
	}

	return output
}
