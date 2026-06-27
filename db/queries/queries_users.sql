-- name: GetUserByEmail :one
SELECT users.*
FROM users
WHERE email = @email;

-- name: GetUserByID :one
SELECT users.*
FROM users
WHERE id = @id;

-- name: createUser :one
INSERT INTO users (team_id, email, password_hash)
VALUES (@team_id, @email, @password_hash)
RETURNING *;
