package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func (c *Conn) Delete(token string) error {
	return c.DeleteCtx(context.Background(), token)
}

func (c *Conn) DeleteCtx(ctx context.Context, token string) error {
	return c.DeleteSessionQuery(ctx, token)
}

func (c *Conn) Find(token string) ([]byte, bool, error) {
	return c.FindCtx(context.Background(), token)
}

func (c *Conn) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	data, err := c.FindSessionQuery(ctx, FindSessionQueryParams{Token: token, Now: time.Now().UTC()})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

func (c *Conn) Commit(token string, data []byte, expiry time.Time) error {
	return c.CommitCtx(context.Background(), token, data, expiry)
}

func (c *Conn) CommitCtx(ctx context.Context, token string, data []byte, expiry time.Time) error {
	return c.CommitSessionQuery(ctx, CommitSessionQueryParams{Token: token, Data: data, Expiry: expiry.UTC()})
}
