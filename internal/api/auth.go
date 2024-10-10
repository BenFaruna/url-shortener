package api

import (
	"encoding/json"
	"fmt"
	"github.com/BenFaruna/url-shortener/internal/controller"
	"github.com/BenFaruna/url-shortener/internal/database"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type Response struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func AuthMux() http.Handler {
	authMux := http.NewServeMux()
	authMux.Handle("/signup", signup())
	authMux.Handle("/signin", signin())

	//return http.StripPrefix("/", authMux)
	return authMux
}

func signup() http.Handler {
	return controller.Post(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		formInput := &controller.AuthFormInput{}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(formInput)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// TODO: validate token
		//token := r.Form.Get("token")

		if formInput.Username == "" || formInput.Password == "" {
			errorHandler(w, r, http.StatusBadRequest, "username or password missing")
			return
		}

		password, err := generatePasswordHash(formInput.Password)
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, "invalid password")
			return
		}

		user := database.User{Username: formInput.Username, Password: password}
		err = user.Add()
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, fmt.Sprintf("User not added: username already exists"))
			return
		}
		userInfo := controller.UserInfo{ID: user.ID, Username: user.Username}
		controller.GlobalSessions.SessionDestroy(w, r)
		err = registerUserSession(w, r, userInfo)
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "Session error")
			return
		}
		w.WriteHeader(http.StatusCreated)
		if err = json.NewEncoder(w).Encode(&Response{StatusCode: http.StatusCreated, Message: fmt.Sprintf("User added: %q", user.Username)}); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
		}
	}))
}

func signin() http.Handler {
	return controller.Post(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		formInput := &controller.AuthFormInput{}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(formInput)
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, fmt.Sprintf("expected %q and %q", "username", "password"))
			return
		}

		if formInput.Username == "" || formInput.Password == "" {
			errorHandler(w, r, http.StatusBadRequest, "username or password missing")
			return
		}

		u := &database.User{}
		err = u.GetUserInfo(formInput.Username)
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, "invalid username")
			return
		}
		if err := comparePassword(u.Password, formInput.Password); err != nil {
			errorHandler(w, r, http.StatusBadRequest, "invalid password")
			return
		}
		userInfo := controller.UserInfo{ID: u.ID, Username: u.Username}
		controller.GlobalSessions.SessionDestroy(w, r)
		if err := registerUserSession(w, r, userInfo); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "Session error")
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(&Response{StatusCode: http.StatusOK, Message: fmt.Sprintf("%q logged in", u.Username)})
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
			return
		}
	}))
}

func generatePasswordHash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return "", fmt.Errorf("invalid password")
	}
	return string(hashedPassword), nil
}

func comparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func registerUserSession(w http.ResponseWriter, r *http.Request, u controller.UserInfo) error {
	sess := controller.GlobalSessions.SessionStart(w, r)
	if err := sess.Set("user", u); err != nil {
		return fmt.Errorf("registerUserSession: unable to register session\n%v", err)
	}
	if err := sess.Set("accesstime", time.Now().Unix()); err != nil {
		return fmt.Errorf("registerUserSession: unable to register session\n%v", err)
	}
	return nil
}

func errorHandler(w http.ResponseWriter, _ *http.Request, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	resp, err := json.Marshal(&Response{StatusCode: code, Message: msg})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.WriteHeader(code)
	_, err = w.Write(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}