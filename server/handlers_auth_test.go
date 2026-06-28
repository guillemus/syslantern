package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"syslantern/config"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthSignUpThenSignIn(t *testing.T) {
	s := newTestServer(t)

	testEmail := "test@example.com"
	testPassword := "correct horse battery staple"
	testPayload := fmt.Sprintf(`{"email":%q,"password":%q}`, testEmail, testPassword)

	signUp := sendPostJSON(
		s, "/sign-up", testPayload)
	assertRedirectHome(t, signUp, "sign-up")
	require.NotEmpty(
		t, signUp.Result().Cookies(),
		"sign-up should create a session cookie")

	user, err := s.DB.GetUserByEmail(t.Context(), testEmail)
	require.NoError(t, err, "sign-up should create user")
	require.NoError(t, bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash), []byte(testPassword),
	), "sign-up should store hashed password")

	team, err := s.DB.GetTeamByID(t.Context(), user.TeamID)
	require.NoError(t, err, "sign-up should create team")
	require.Equal(t, "My Team", team.Name, "sign-up should create default team")

	assertCookiesAuthenticate(t, s, signUp.Result().Cookies(), testEmail)

	signIn := sendPostJSON(s, "/sign-in", testPayload, signUp.Result().Cookies()...)
	assertRedirectHome(t, signIn, "sign-in")
	require.NotEmpty(t, signIn.Result().Cookies(), "sign-in should set a session cookie")
	assertCookiesAuthenticate(t, s, signIn.Result().Cookies(), testEmail)
}

func TestHandleSignInRejectsBadPassword(t *testing.T) {
	s := newTestServer(t)
	testEmail := "test@example.com"
	wrongPassword := "wrong"
	_, err := s.DB.CreateUserAndTeam(
		t.Context(), testEmail, "$2a$10$seFT5QbguA5gFM1daVH0xec0GUnf31awNmVK89yjQ5A9vuwU6kyhu")
	require.NoError(t, err, "create existing user fixture")

	testPayload := fmt.Sprintf(`{"email":%q,"password":%q}`, testEmail, wrongPassword)
	rr := sendPostJSON(s, "/sign-in", testPayload)

	require.Equal(t, http.StatusOK, rr.Code, "bad sign-in should re-render the form")
	require.Contains(t, rr.Body.String(), "Invalid email or password.", "bad sign-in should show invalid credentials copy")
	require.Empty(
		t, rr.Result().Cookies(),
		"bad sign-in should not create a session cookie")
}

func sendPostJSON(s *Server, url string, json string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	req := newPostRequest(url, json)
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)
	return rr
}

func assertRedirectHome(t *testing.T, rr *httptest.ResponseRecorder, action string) {
	t.Helper()

	require.Equal(t, http.StatusOK, rr.Code, "%s should respond successfully", action)
	contentType := rr.Header().Get("Content-Type")
	require.Equal(t, "text/event-stream", contentType, "%s success should stream a Datastar redirect", action)
	require.Contains(t, rr.Body.String(), `window.location.href = "/"`, "%s success should redirect home", action)
}

func assertCookiesAuthenticate(
	t *testing.T, s *Server, cookies []*http.Cookie, email string,
) {
	t.Helper()

	req := newGetRequest("/is-authenticated")
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Equal(t, email, rr.Body.String(), "authenticated user should match email")
}

func newGetRequest(url string) *http.Request {
	return httptest.NewRequest(http.MethodGet, url, nil)
}

func newPostRequest(url string, json string) *http.Request {
	return httptest.NewRequest(http.MethodPost, url, strings.NewReader(json))
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	s := NewServerFromConfig(config.Config{
		DBPath:       fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()),
		Port:         "0",
		AssetVersion: "test",
		Dev:          true,
	})
	t.Cleanup(func() { require.NoError(t, s.DB.Close(), "close test DB") })

	require.NoError(t, s.DB.ExecSchemaScript(), "apply schema")

	return s
}
