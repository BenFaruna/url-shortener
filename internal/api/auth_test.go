package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BenFaruna/url-shortener/internal/api"
	_ "github.com/BenFaruna/url-shortener/internal/controller"
	"os"
	//_ "github.com/BenFaruna/url-shortener/internal/database"
	"net/http"
	"net/http/httptest"
	"testing"
)

type AuthForm struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestSignup(t *testing.T) {
	t.Run("user can signup using username and password", func(t *testing.T) {
		formInput := &bytes.Buffer{}
		input := AuthForm{Username: "devfaruna", Password: "1xshbixen6svub"}
		output := &api.Response{}
		if data, err := json.Marshal(input); err != nil {
			t.Fatal(err)
		} else {
			formInput.WriteString(string(data))
		}
		handler := api.AuthMux()
		request := httptest.NewRequest(http.MethodPost, "/signup", formInput)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		err := json.NewDecoder(response.Body).Decode(output)
		if err != nil {
			t.Fatal(err)
		}

		if output.StatusCode != http.StatusCreated {
			t.Errorf("expected status code %d, got %d", http.StatusCreated, output.StatusCode)
		}
		if output.Message != fmt.Sprintf("User added: %q", input.Username) {
			t.Errorf("expected message %q, got %q", fmt.Sprintf("User added: %q", input.Username), output.Message)
		}
	})
}

func TestSignin(t *testing.T) {
	t.Run("user can signin using username and password", func(t *testing.T) {
		formInput := &bytes.Buffer{}
		input := AuthForm{Username: "devfaruna", Password: "1xshbixen6svub"}

		output := &api.Response{}
		if data, err := json.Marshal(input); err != nil {
			t.Fatal(err)
		} else {
			formInput.WriteString(string(data))
		}
		handler := api.AuthMux()
		request := httptest.NewRequest(http.MethodPost, "/signin", formInput)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		err := json.NewDecoder(response.Body).Decode(output)
		if err != nil {
			t.Fatal(err)
		}

		if output.StatusCode != http.StatusOK {
			t.Errorf("expected status code %d, got %d", http.StatusOK, output.StatusCode)
		}
		if output.Message != fmt.Sprintf("%q logged in", input.Username) {
			t.Errorf("expected message %q, got %q", fmt.Sprintf("%q logged in", input.Username), output.Message)
		}
	})

	t.Run("cookie is created after signin", func(t *testing.T) {
		signup(t)
		formInput := &bytes.Buffer{}
		input := AuthForm{Username: "devfaruna1", Password: "1xshbixen6svub"}
		if data, err := json.Marshal(input); err != nil {
			t.Fatal(err)
		} else {
			formInput.WriteString(string(data))
		}
		handler := api.AuthMux()
		request := httptest.NewRequest(http.MethodPost, "/signin", formInput)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		resp := response.Result()
		cookies := resp.Cookies()
		if len(cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(cookies))
		}
	})
}

func TestLogout(t *testing.T) {
	formInput := &bytes.Buffer{}
	input := AuthForm{Username: "devfaruna1", Password: "1xshbixen6svub"}
	if data, err := json.Marshal(input); err != nil {
		t.Fatal(err)
	} else {
		formInput.WriteString(string(data))
	}

	handler := api.AuthMux()
	signInRequest := httptest.NewRequest(http.MethodPost, "/signin", formInput)
	signInResponse := httptest.NewRecorder()
	handler.ServeHTTP(signInResponse, signInRequest)

	resp := signInResponse.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("expected cookie to be set")
	}

	//sess := controller.GlobalSessions.SessionStart(signInResponse, signInRequest)
	//user := sess.Get("user")
	//if user != nil {
	//	t.Errorf("expect nil, got %v", user)
	//}

	signOutRequest := httptest.NewRequest(http.MethodPost, "/signout", nil)
	signOutResponse := httptest.NewRecorder()
	signOutRequest.Header.Set("Set-Cookie", signInResponse.Header().Get("Set-Cookie"))

	handler.ServeHTTP(signOutResponse, signOutRequest)
	t.Logf("%v", signOutResponse.Header().Get("Set-Cookie"))
	//sess := controller.GlobalSessions.SessionStart(signOutResponse, signOutRequest)
	//user := sess.Get("user")
	//if user != nil {
	//	t.Errorf("expect nil, got %v", user.(controller.UserInfo))
	//}

	t.Cleanup(func() { os.RemoveAll("app.db") })
}

func signup(t testing.TB) (http.ResponseWriter, *http.Request) {
	t.Helper()
	formInput := &bytes.Buffer{}
	input := AuthForm{Username: "devfaruna", Password: "1xshbixen6svub"}
	if data, err := json.Marshal(input); err != nil {
		t.Fatal(err)
	} else {
		formInput.WriteString(string(data))
	}
	handler := api.AuthMux()
	request := httptest.NewRequest(http.MethodPost, "/signup", formInput)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)
	return response, request
}
