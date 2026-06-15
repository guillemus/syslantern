package server

import (
	"errors"
	"net/http"
	"strings"

	"app/db"
	"app/views"

	"golang.org/x/crypto/bcrypt"
)

const (
	genericAuthErrorMessage        = "Something went wrong. Please try again."
	invalidCredentialsErrorMessage = "Invalid email or password."
	emailTakenErrorMessage         = "An account with that email already exists."
)

func (s *Server) RegisterAuthRoutes() {
}

func (s *Server) HandleSignInPage(w http.ResponseWriter, r *http.Request) {
	s.Renderer.RenderSignIn(w, views.SignInData{})
}

func (s *Server) HandleSignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")

	user, err := s.DB.GetUserByEmail(ctx, email)
	if err != nil {
		if db.IsNotFound(err) {
			s.Logger.Warn("sign in: user not found", "email", email)
			s.Renderer.RenderSignIn(w, views.SignInData{
				Email: email,
				Error: invalidCredentialsErrorMessage,
			})
			return
		}

		s.Logger.Error("sign in: lookup user", "email", email, "err", err)
		s.Renderer.RenderSignIn(w, views.SignInData{
			Email: email,
			Error: genericAuthErrorMessage,
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.Logger.Warn("sign in: bad password", "email", email)
		s.Renderer.RenderSignIn(w, views.SignInData{
			Email: email,
			Error: invalidCredentialsErrorMessage,
		})
		return
	}

	if err := s.Sessions.RenewToken(ctx); err != nil {
		s.Logger.Error("sign in: renew token", "err", err)
		s.Renderer.RenderSignIn(w, views.SignInData{
			Email: email,
			Error: genericAuthErrorMessage,
		})
		return
	}
	s.Sessions.Put(ctx, "user_id", user.ID)

	http.Redirect(w, r, "/dash", http.StatusSeeOther)
}

func (s *Server) HandleSignUpPage(w http.ResponseWriter, r *http.Request) {
	s.Renderer.RenderSignUp(w, views.SignUpData{})
}

func (s *Server) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.Logger.Error("sign up: hash password", "err", err)
		s.Renderer.RenderSignUp(w, views.SignUpData{
			Email: email,
			Error: genericAuthErrorMessage,
		})
		return
	}

	user, err := s.DB.CreateUser(ctx, email, string(hash))
	if err != nil {
		if errors.Is(err, db.ErrDuplicateEmail) {
			s.Logger.Warn("sign up: duplicate email", "email", email)
			s.Renderer.RenderSignUp(w, views.SignUpData{
				Email: email,
				Error: emailTakenErrorMessage,
			})
			return
		}

		s.Logger.Error("sign up: create user", "email", email, "err", err)
		s.Renderer.RenderSignUp(w, views.SignUpData{
			Email: email,
			Error: genericAuthErrorMessage,
		})
		return
	}

	if err := s.Sessions.RenewToken(ctx); err != nil {
		s.Logger.Error("sign up: renew token", "err", err)
		s.Renderer.RenderSignUp(w, views.SignUpData{
			Email: email,
			Error: genericAuthErrorMessage,
		})
		return
	}
	s.Sessions.Put(r.Context(), "user_id", user.ID)

	http.Redirect(w, r, "/dash", http.StatusSeeOther)
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.Sessions.RenewToken(ctx); err != nil {
		s.Logger.Error("logout: renew token", "err", err)
	}
	if err := s.Sessions.Destroy(ctx); err != nil {
		s.Logger.Error("logout: destroy session", "err", err)
	}
	http.Redirect(w, r, "/sign-in", http.StatusSeeOther)
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
