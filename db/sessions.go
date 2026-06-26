package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type SessionStore struct {
	conn *Conn
}

func NewSessionStore(conn *Conn) *SessionStore {
	return &SessionStore{conn: conn}
}

func (s *SessionStore) Delete(token string) error {
	return s.DeleteCtx(context.Background(), token)
}

func (s *SessionStore) DeleteCtx(ctx context.Context, token string) error {
	return s.conn.deleteSession(ctx, SessionToken(token))
}

func (s *SessionStore) Find(token string) ([]byte, bool, error) {
	return s.FindCtx(context.Background(), token)
}

func (s *SessionStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	data, err := s.conn.findSession(ctx, findSessionParams{Token: SessionToken(token), Now: time.Now().UTC()})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func (s *SessionStore) Commit(token string, data []byte, expiry time.Time) error {
	return s.CommitCtx(context.Background(), token, data, expiry)
}

func (s *SessionStore) CommitCtx(ctx context.Context, token string, data []byte, expiry time.Time) error {
	return s.conn.commitSession(ctx, commitSessionParams{Token: SessionToken(token), Data: data, Expiry: expiry.UTC()})
}
