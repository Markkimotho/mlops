package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestLocalConsoleRequiresLogin(t *testing.T) {
	manager := NewLocalSessionManager("admin", "correct-password")
	handler := manager.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	anonymous := httptest.NewRecorder()
	handler.ServeHTTP(anonymous, httptest.NewRequest(http.MethodGet, "/console.html", nil))
	if anonymous.Code != http.StatusFound || !strings.HasPrefix(anonymous.Header().Get("Location"), "/login.html") {
		t.Fatalf("anonymous console response: %d %s", anonymous.Code, anonymous.Header().Get("Location"))
	}

	form := url.Values{"username": {"admin"}, "password": {"correct-password"}, "return_to": {"/console.html"}}
	login := httptest.NewRecorder()
	loginRequest := httptest.NewRequest(http.MethodPost, "/auth/local/login", strings.NewReader(form.Encode()))
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	handler.ServeHTTP(login, loginRequest)
	if login.Code != http.StatusFound || login.Header().Get("Location") != "/console.html" {
		t.Fatalf("login response: %d %s", login.Code, login.Header().Get("Location"))
	}
	request := httptest.NewRequest(http.MethodGet, "/console.html", nil)
	for _, cookie := range login.Result().Cookies() {
		request.AddCookie(cookie)
	}
	authenticated := httptest.NewRecorder()
	handler.ServeHTTP(authenticated, request)
	if authenticated.Code != http.StatusOK {
		t.Fatalf("authenticated console response: %d", authenticated.Code)
	}
}

func TestLocalLoginRejectsInvalidCredentials(t *testing.T) {
	manager := NewLocalSessionManager("admin", "correct-password")
	form := url.Values{"username": {"admin"}, "password": {"wrong"}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/auth/local/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	manager.Handler(http.NotFoundHandler()).ServeHTTP(response, request)
	if response.Code != http.StatusFound || response.Header().Get("Location") != "/login.html?error=invalid" {
		t.Fatalf("invalid login response: %d %s", response.Code, response.Header().Get("Location"))
	}
}
