package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/BenFaruna/url-shortener/internal/controller"
	"github.com/BenFaruna/url-shortener/internal/database"
	"golang.org/x/crypto/bcrypt"
	"html"
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
	authMux.Handle("/signout", signout())

	//return http.StripPrefix("/", authMux)
	return authMux
}

// signup godoc
//
//	@Summary		signup user
//	@Description	register user details to database
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			username	body		string	true	"username"
//	@Param			password	body		string	true	"password"
//	@Success		201			{object}	Response
//
//	@Failure		400			{object}	Response
//	@Failure		500			{object}	Response
//
//	@Router			/signup [post]
func signup() http.Handler {
	return controller.Post(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		formInput := &controller.AuthFormInput{}
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(formInput)
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, "bad json input")
			return
		}
		// TODO: validate token
		//token := r.Form.Get("token")

		if formInput.Username == "" || formInput.Password == "" {
			errorHandler(w, r, http.StatusBadRequest, "username or password missing")
			return
		}

		username := html.EscapeString(formInput.Username)
		password := html.EscapeString(formInput.Password)

		password, err = generatePasswordHash(password)
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, "invalid password")
			return
		}

		user := database.User{Username: username, Password: password}
		err = user.Add()
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, fmt.Sprintf("User not added: username already exists"))
			return
		}
		//userInfo := controller.UserInfo{ID: user.ID, Username: user.Username}
		controller.GlobalSessions.SessionDestroy(w, r)
		//err = registerUserSession(w, r, userInfo)
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

// signin godoc
//
//	@Summary		signin user
//	@Description	login user and create user session
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			username	body		string	true	"username"
//	@Param			password	body		string	true	"password"
//	@Success		200			{object}	Response
//
//	@Failure		400			{object}	Response
//	@Failure		500			{object}	Response
//
//	@Router			/signin [post]
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

		username := html.EscapeString(formInput.Username)
		password := html.EscapeString(formInput.Password)

		u := &database.User{}
		err = u.GetUserInfo(username)
		if err != nil {
			errorHandler(w, r, http.StatusBadRequest, "invalid username")
			return
		}
		if err := comparePassword(u.Password, password); err != nil {
			errorHandler(w, r, http.StatusBadRequest, "invalid password")
			return
		}
		userInfo := controller.UserInfo{ID: u.ID, Username: u.Username}
		//controller.GlobalSessions.SessionDestroy(w, r)
		if err := registerUserSession(w, r, userInfo); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "Session error")
			return
		}

		data := []byte(fmt.Sprintf("%s:%s", username, password))
		dst := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
		base64.StdEncoding.Encode(dst, data)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Authorization", fmt.Sprintf("Basic %s", string(dst)))
		err = json.NewEncoder(w).Encode(&Response{StatusCode: http.StatusOK, Message: fmt.Sprintf("%q logged in", u.Username)})
		if err != nil {
			errorHandler(w, r, http.StatusInternalServerError, fmt.Sprintf("Error writing response: %v", err))
			return
		}
	}))
}

// signout godoc
//
//	@Summary					signout user
//	@Description				logout user and destroy user session
//	@Tags						auth
//	@Accept						json
//	@Produce					json
//	@securitydefinitions.basic	BasicAuth
//	@Success					200	{object}	Response
//
//	@Failure					401	{object}	Response
//	@Failure					500	{object}	Response
//
//	@Router						/signout [post]
func signout() http.Handler {
	return controller.Post(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := controller.GlobalSessions.SessionStart(w, r)
		user := sess.Get("user")
		if user == nil {
			errorHandler(w, r, http.StatusUnauthorized, "user not authenticated")
			return
		}
		if err := sess.Delete("user"); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "Session error")
			return
		}
		if err := removeUserSession(w, r); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "Session error")
			return
		}
		w.Header().Del("Authorization")
		w.Header().Set("Content-Type", "application/json")
		resp := &Response{StatusCode: http.StatusOK, Message: fmt.Sprintf("%q logged out", user.(controller.UserInfo).Username)}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			errorHandler(w, r, http.StatusInternalServerError, "Error writing response")
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

func removeUserSession(w http.ResponseWriter, r *http.Request) error {
	sess := controller.GlobalSessions.SessionStart(w, r)
	if err := sess.Delete("user"); err != nil {
		return fmt.Errorf("removeUserSession: unable to remove session\n%v", err)
	}
	if err := sess.Delete("accesstime"); err != nil {
		return fmt.Errorf("removeUserSession: unable to remove session\n%v", err)
	}
	controller.GlobalSessions.SessionDestroy(w, r)
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
