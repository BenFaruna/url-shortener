package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BenFaruna/url-shortener/internal/api"
	_ "github.com/BenFaruna/url-shortener/internal/controller"
	"os"
	"strings"

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

	t.Run("cookie is created after signup", func(t *testing.T) {
		formInput := &bytes.Buffer{}
		input := AuthForm{Username: "devfaruna1", Password: "1xshbixen6svub"}
		if data, err := json.Marshal(input); err != nil {
			t.Fatal(err)
		} else {
			formInput.WriteString(string(data))
		}
		handler := api.AuthMux()
		request := httptest.NewRequest(http.MethodPost, "/signup", formInput)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		cookie := response.Header().Get("Set-Cookie")
		cookie = strings.TrimPrefix(cookie, "[")
		cookie = strings.TrimSuffix(cookie, "]")
		if !strings.HasPrefix(cookie, "gosessionid=") {
			t.Errorf("expected cookie %q, got %q", "gosessionid", cookie)
		}
		if !strings.HasSuffix(cookie, "HttpOnly") {
			t.Errorf("expected cookie %q, got %q", "HttpOnly", cookie)
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

	t.Cleanup(func() {
		os.RemoveAll("app.db")
	})
}
