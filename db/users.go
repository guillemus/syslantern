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

func (c *Conn) GetUserDefaultWorkspace(ctx context.Context, userID int64) (Workspace, error) {
	row, err := c.GetUserDefaultWorkspaceQuery(ctx, userID)
	return row.Workspace, err
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

func (c *Conn) CreateUserWithWorkspace(ctx context.Context, email, passwordHash string) (User, error) {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()
	qtx := c.WithTx(tx)

	user, err := qtx.CreateUserQuery(ctx, CreateUserQueryParams{Email: email, PasswordHash: passwordHash})
	if err != nil {
		return User{}, err
	}

	workspace, err := qtx.CreateWorkspaceQuery(ctx, "My workspace")
	if err != nil {
		return User{}, err
	}

	err = qtx.AddWorkspaceMemberQuery(ctx, AddWorkspaceMemberQueryParams{
		WorkspaceID: workspace.ID,
		UserID:      user.ID,
		Role:        "owner",
	})
	if err != nil {
		return User{}, err
	}

	return user, tx.Commit()
}
