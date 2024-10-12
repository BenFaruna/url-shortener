package controller

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	_ "github.com/BenFaruna/url-shortener/internal/session/providers/memory"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BenFaruna/url-shortener/internal/model"
	"github.com/BenFaruna/url-shortener/internal/session"
)

type FormToken struct {
	Token string
}

type FormPageData struct {
	FormToken
	User UserInfo
}

type UserInfo struct {
	ID       int64
	Username string
}

type AuthFormInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	//Token    string `json:"token,omitempty"`
}

var GlobalSessions *session.Manager

// HomeHandler accept requests to the home route and provide responses are redirection for short routes
func HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderer, err := NewRenderer()
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("route %q does not exists", r.URL.Path))
		}
		switch r.URL.Path {
		case "/":
			data := &struct {
				URLs []model.URLInfo
				User UserInfo
			}{}
			data.URLs = model.Db.GetAll()
			data.User = getUserSignedIn(w, r)
			renderer.Render(w, "index.gohtml", data)
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
	return Get(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderer, err := NewRenderer()
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}

		curtime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(curtime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))
		if err = renderer.Render(w, "login.gohtml", FormPageData{FormToken: FormToken{token}}); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}
	}))
}

func SignupHandler() http.Handler {
	return Get(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderer, err := NewRenderer()
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}
		curtime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(curtime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		if err = renderer.Render(w, "signup.gohtml", FormPageData{FormToken: FormToken{token}}); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}
	}))
}

func errorHandler(w http.ResponseWriter, _ *http.Request, status int, errorMessage string) {
	//w.WriteHeader(status)
	// w.Header().Set("Content-Type", "application/json")
	if status == http.StatusNotFound {
		fmt.Fprint(w, errorMessage)
	}
}

func getUserSignedIn(w http.ResponseWriter, r *http.Request) UserInfo {
	sess := GlobalSessions.SessionStart(w, r)
	user := sess.Get("user")
	if user == nil {
		return UserInfo{}
	}
	return user.(UserInfo)
}

func GenerateShortString() string {
	output := ""

	for i := 0; i < 6; i++ {
		n := rand.Intn(51)
		output += string(model.Characters[n])
	}

	return output
}

func init() {
	session.GlobalSession, _ = session.NewManager("memory", "gosessionid", 3600)
	GlobalSessions = session.GlobalSession
	go session.GlobalSession.GC()
}
