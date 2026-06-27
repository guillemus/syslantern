-- name: deleteSession :exec
DELETE FROM sessions
WHERE token = @token;

-- name: findSession :one
SELECT data
FROM sessions
WHERE token = @token
AND expiry > @now;

-- name: commitSession :exec
INSERT INTO sessions (token, data, expiry)
VALUES (@token, @data, @expiry)
ON CONFLICT(token) DO UPDATE SET data = excluded.data, expiry = excluded.expiry;
