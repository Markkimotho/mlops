package auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSessionLoginCreatesStateAndRedirects(t *testing.T) {
	manager, err := NewSessionManager(SessionConfig{
		ClientID: "console", ClientSecret: "secret",
		AuthURL: "https://identity.example/auth", TokenURL: "https://identity.example/token",
		RedirectURL: "https://nexus.example/auth/callback", Secure: true,
	}, &Verifier{})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/auth/login?return_to=%2Fconsole.html%3Fview%3Dmodels", nil)
	response := httptest.NewRecorder()
	manager.Handler(http.NotFoundHandler()).ServeHTTP(response, request)
	if response.Code != http.StatusFound {
		t.Fatalf("login returned %d", response.Code)
	}
	location, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	if location.Host != "identity.example" || location.Query().Get("state") == "" ||
		location.Query().Get("redirect_uri") != "https://nexus.example/auth/callback" {
		t.Fatalf("unexpected authorization redirect: %s", location)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 || !cookies[0].HttpOnly || !cookies[0].Secure {
		t.Fatalf("state cookies are not hardened: %+v", cookies)
	}
}

func TestSafeReturnToPreventsExternalRedirect(t *testing.T) {
	for _, value := range []string{"https://evil.example", "//evil.example/path", "javascript:alert(1)", ""} {
		if got := safeReturnTo(value); got != "/console.html" {
			t.Errorf("safeReturnTo(%q) = %q", value, got)
		}
	}
	if got := safeReturnTo("/console.html?view=models"); got != "/console.html?view=models" {
		t.Fatalf("valid return path changed: %s", got)
	}
}

func TestVerifierRedirectsBrowserButNotAPI(t *testing.T) {
	verifier := New(Config{})
	handler := verifier.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	console := httptest.NewRecorder()
	handler.ServeHTTP(console, httptest.NewRequest(http.MethodGet, "/console.html", nil))
	if console.Code != http.StatusFound || !strings.HasPrefix(console.Header().Get("Location"), "/auth/login?") {
		t.Fatalf("console must redirect to login: %d %s", console.Code, console.Header().Get("Location"))
	}
	api := httptest.NewRecorder()
	handler.ServeHTTP(api, httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil))
	if api.Code != http.StatusUnauthorized {
		t.Fatalf("API must return 401, got %d", api.Code)
	}
	landing := httptest.NewRecorder()
	handler.ServeHTTP(landing, httptest.NewRequest(http.MethodGet, "/", nil))
	if landing.Code != http.StatusOK {
		t.Fatalf("landing must remain public, got %d", landing.Code)
	}
}
