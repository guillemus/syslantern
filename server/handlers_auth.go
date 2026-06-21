package server

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"time"

	"syslantern/db"
	"syslantern/views"

	"github.com/starfederation/datastar-go/datastar"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) HandleSignInPage(w http.ResponseWriter, r *http.Request) {
	s.Renderer.RenderSignIn(w)
}

func (s *Server) HandleSignIn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload views.SignInSignals
	if err := datastar.ReadSignals(r, &payload); err != nil {
		s.Logger.Warn("sign in: read signals", "err", err)
		s.Renderer.RenderSignInGenericAuthErr(w, payload.Email)
		return
	}

	user, err := s.DB.GetUserByEmail(ctx, payload.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Logger.Warn("sign in: user not found", "email", payload.Email)
			s.Renderer.RenderSignInGenericAuthErr(w, payload.Email)
			return
		}

		s.Logger.Error("sign in: lookup user", "email", payload.Email, "err", err)
		s.Renderer.RenderSignInGenericAuthErr(w, payload.Email)
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash.String), []byte(payload.Password),
	); err != nil {
		s.Logger.Warn("sign in: bad password", "email", payload.Email)
		s.Renderer.RenderSignInInvalidCredsErr(w, payload.Email)
		return
	}

	if err := s.Sessions.RenewToken(ctx); err != nil {
		s.Logger.Error("sign in: renew token", "err", err)
		s.Renderer.RenderSignInInvalidCredsErr(w, payload.Email)
		return
	}
	s.Sessions.Put(ctx, "user_id", user.ID)

	if err := s.WriteSessionCookie(ctx, w); err != nil {
		s.Logger.Error("sign in: commit session", "err", err)
		s.Renderer.RenderSignInInvalidCredsErr(w, payload.Email)
		return
	}

	sse := datastar.NewSSE(w, r)
	sse.Redirect("/")
}

// WriteSessionCookie writes the session cookie immediately.
// We need this, otherwise datastar.NewSSE will flush response headers immediately,
// before SCS LoadAndSave
func (s *Server) WriteSessionCookie(ctx context.Context, w http.ResponseWriter) error {
	token, expiry, err := s.Sessions.Commit(ctx)
	if err != nil {
		return err
	}
	s.Sessions.WriteSessionCookie(ctx, w, token, expiry)

	return nil
}

func (s *Server) HandleSignUpPage(w http.ResponseWriter, r *http.Request) {
	s.Renderer.RenderSignUp(w)
}

func (s *Server) HandleSignUp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload views.SignUpSignals
	if err := datastar.ReadSignals(r, &payload); err != nil {
		s.Logger.Warn("sign up: read signals", "err", err)
		s.Renderer.RenderSignUpGenericAuthErr(w, payload.Email)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		s.Logger.Error("sign up: hash password", "err", err)
		s.Renderer.RenderSignUpGenericAuthErr(w, payload.Email)
		return
	}

	user, err := s.DB.CreateUserAndTeam(ctx, payload.Email, string(hash))
	if err != nil {
		if errors.Is(err, db.ErrDuplicateEmail) {
			s.Logger.Warn("sign up: duplicate email", "email", payload.Email)
			s.Renderer.RenderSignUpGenericAuthErr(w, payload.Email)
			return
		}

		s.Logger.Error("sign up: create user", "email", payload.Email, "err", err)
		s.Renderer.RenderSignUpGenericAuthErr(w, payload.Email)
		return
	}

	if err := s.Sessions.RenewToken(ctx); err != nil {
		s.Logger.Error("sign up: renew token", "err", err)
		s.Renderer.RenderSignUpGenericAuthErr(w, payload.Email)
		return
	}
	s.Sessions.Put(ctx, "user_id", user.ID)

	if err := s.WriteSessionCookie(ctx, w); err != nil {
		s.Logger.Error("sign up: commit session", "err", err)
		s.Renderer.RenderSignUpGenericAuthErr(w, payload.Email)
		return
	}

	sse := datastar.NewSSE(w, r)
	sse.Redirect("/")
}

func (s *Server) HandleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := s.Sessions.RenewToken(ctx); err != nil {
		s.Logger.Error("logout: renew token", "err", err)
	}
	if err := s.Sessions.Destroy(ctx); err != nil {
		s.Logger.Error("logout: destroy session", "err", err)
	}
	s.Sessions.WriteSessionCookie(ctx, w, "", time.Time{})
	sse := datastar.NewSSE(w, r)
	sse.Redirect("/sign-in")
}

func (s *Server) GetAuthenticatedUser(w http.ResponseWriter, r *http.Request) (user db.User, exists bool) {
	userID := s.Sessions.GetInt64(r.Context(), "user_id")
	if userID == 0 {
		return db.User{}, false
	}

	user, err := s.DB.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.Logger.Warn("auth: session user not found", "user_id", userID)
		} else {
			s.Logger.Error("auth: load session user", "user_id", userID, "err", err)
		}
		s.Sessions.Remove(r.Context(), "user_id")
		return db.User{}, false
	}

	return user, true
}
