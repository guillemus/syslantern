package db

import "context"

func (c *Conn) CreateUserAndTeam(ctx context.Context, email, passwordHash string) (User, error) {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return User{}, err
	}
	defer tx.Rollback()

	q := c.Queries.WithTx(tx)

	team, err := q.createTeam(ctx, "My Team")
	if err != nil {
		return User{}, err
	}

	user, err := q.createUser(ctx, createUserParams{
		TeamID:       team.ID,
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return User{}, err
	}

	if err := tx.Commit(); err != nil {
		return User{}, err
	}

	return user, nil
}
