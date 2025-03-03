package controller

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/BenFaruna/url-shortener/internal/database"

	"github.com/BenFaruna/url-shortener/internal/session"
	_ "github.com/BenFaruna/url-shortener/internal/session/providers/memory"
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
				URLs []database.URLInfo
				User UserInfo
			}{}
			data.User = getUserSignedIn(w, r)
			data.URLs = database.ShortUrls{}.GetAll(data.User.ID)
			renderer.Render(w, "index.gohtml", data)
			return
		default:
			shortID := strings.TrimPrefix(r.URL.Path, "/")
			url, ok := database.ShortUrls{}.Get(shortID)
			if !ok {
				errorHandler(w, r, http.StatusNotFound, fmt.Sprintf("route %q does not exists", r.URL.Path))
				return
			}

			http.Redirect(w, r, url, http.StatusMovedPermanently)
		}
	}
}

func AddressHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetAddressHandler(w, r)
			return

		case http.MethodPost:
			ShortenAddressHandler(w, r, GenerateShortString)
			return

		case http.MethodDelete:
			DeleteAddressHandler(w, r)
			return

		default:
			errorHandler(w, r, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}
}

// ShortenAddressHandler godoc
//
//	@Summary		Shortens a url
//	@Description	returns the short code of url shortened
//	@Tags			url
//	@Accept			json
//	@Produce		json
//	@Param			url	body		database.Body	true	"url to shorten"
//	@Success		201	{object}	database.StatusMessage
//
//	@Failure		403	{object}	database.StatusMessage
//
//	@Router			/address/shorten [post]
func ShortenAddressHandler(w http.ResponseWriter, r *http.Request, shortStringFunc func() string) {
	if r.URL.Path != "/address/shorten" {
		errorHandler(w, r, http.StatusNotFound, fmt.Sprintf("route %q does not exists", r.URL.Path))
		return
	}
	user := getUserSignedIn(w, r)
	var data database.Body
	json.NewDecoder(r.Body).Decode(&data)

	shortenedURL := shortStringFunc()

	shortenedURL, err := database.ShortUrls{}.Add(user.ID, data.URL, shortenedURL)

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(&database.StatusMessage{
			Message: "Error",
			Data:    err.Error(),
		})
		return
	}

	w.WriteHeader(201)
	json.NewEncoder(w).Encode(&database.StatusMessage{
		Message: "url shortened",
		Data:    r.URL.Hostname() + shortenedURL,
	})
}

// DeleteAddressHandler godoc
//
//	@Summary		delete address and short code from database
//	@Description	deletes short urls using the url id
//	@Tags			url
//	@Accept			json
//	@Produce		json
//	@Param			url-id	path		string	true	"url id in database"
//	@Success		200		{object}	database.StatusMessage
//
//	@Failure		404		{object}	database.StatusMessage
//	@Failure		403		{object}	database.StatusMessage
//
//	@Router			/address/{url-id} [delete]
func DeleteAddressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	shortAddress := strings.TrimPrefix(r.URL.Path, "/address/")

	user := getUserSignedIn(w, r)

	isUserUrl := database.ShortUrls{}.IsUserURL(user.ID, shortAddress)
	if !isUserUrl {
		errorHandler(w, r, http.StatusUnauthorized, "cannot delete url")
		return
	}

	err := database.ShortUrls{}.Delete(user.ID, shortAddress)
	if err != nil {
		log.Print(err)
		errorHandler(w, r, http.StatusNotFound, fmt.Sprintf("%q not found", shortAddress))
		return
	}
	res := database.StatusMessage{Message: "success"}
	if err := json.NewEncoder(w).Encode(&res); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorHandler(w, r, http.StatusInternalServerError, "error encoding response")
	}
}

// GetAddressHandler godoc
//
//	@Summary		Full address of short code
//	@Description	returns the full url of the short code
//	@Tags			url
//	@Accept			json
//	@Produce		json
//	@Param			url	path		string	true	"shortcode to url"
//	@Success		200	{object}	database.StatusMessage
//
//	@Failure		404	{object}	database.StatusMessage
//
//	@Router			/address/{url} [get]
func GetAddressHandler(w http.ResponseWriter, r *http.Request) {
	shortAddress := strings.TrimPrefix(r.URL.Path, "/address/")

	w.Header().Set("Content-Type", "application/json")
	url, ok := database.ShortUrls{}.Get(shortAddress)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(database.StatusMessage{
			Data:    "address does not exist",
			Message: "Error",
		})
		return
	}

	json.NewEncoder(w).Encode(database.StatusMessage{
		Data:    url,
		Message: "address found",
	})
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

func ProfileHandler() http.Handler {
	return Get(AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderer, err := NewRenderer()
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}
		var user UserInfo
		sess := GlobalSessions.SessionStart(w, r)
		u := sess.Get("user")
		if u != nil {
			user = u.(UserInfo)
		} else {
			http.Redirect(w, r, "/login", http.StatusPermanentRedirect)
			return
		}
		data := struct{ User UserInfo }{
			User: user,
		}
		if err := renderer.Render(w, "profile.gohtml", data); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "")
			return
		}
	})))
}

func errorHandler(w http.ResponseWriter, _ *http.Request, status int, errorMessage string) {
	w.WriteHeader(status)
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
		output += string(database.Characters[n])
	}

	return output
}

func init() {
	session.GlobalSession, _ = session.NewManager("memory", "gosessionid", 3600)
	GlobalSessions = session.GlobalSession
	go session.GlobalSession.GC()
}

func (u UserInfo) FirstLetter() string {
	return string(unicode.ToUpper(rune(u.Username[0])))
}
