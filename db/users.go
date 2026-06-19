package db

import (
	"context"
	"errors"
	"strings"
)

var ErrDuplicateEmail = errors.New("duplicate email")

func (c *Conn) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row, err := c.GetUserByEmailQuery(ctx, email)
	return row.User, err
}

func (c *Conn) GetUserByID(ctx context.Context, id int64) (User, error) {
	row, err := c.GetUserByIDQuery(ctx, id)
	return row.User, err
}

func (c *Conn) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	user, err := c.CreateUserQuery(ctx, CreateUserQueryParams{Email: email, PasswordHash: passwordHash})
	if err != nil {
		if strings.Contains(err.Error(), "users_email_lower_idx") || strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return User{}, ErrDuplicateEmail
		}
		return User{}, err
	}
	return user, nil
}
