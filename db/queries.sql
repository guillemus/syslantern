-- name: GetUserByEmailQuery :one
SELECT sqlc.embed(users)
FROM users
WHERE email = @email;

-- name: GetUserByIDQuery :one
SELECT sqlc.embed(users)
FROM users
WHERE id = @id;

-- name: CreateUserQuery :one
INSERT INTO users (email, password_hash)
VALUES (@email, @password_hash)
RETURNING *;

-- name: CreateWorkspaceQuery :one
INSERT INTO workspaces (name)
VALUES (@name)
RETURNING *;

-- name: AddWorkspaceMemberQuery :exec
INSERT INTO workspace_members (workspace_id, user_id, role)
VALUES (@workspace_id, @user_id, @role);

-- name: GetUserDefaultWorkspaceQuery :one
SELECT sqlc.embed(workspaces)
FROM workspaces
JOIN workspace_members ON workspace_members.workspace_id = workspaces.id
WHERE workspace_members.user_id = @user_id
ORDER BY workspace_members.created_at, workspaces.id
LIMIT 1;

-- name: GetAgentQuery :one
SELECT sqlc.embed(agents)
FROM agents
WHERE workspace_id = @workspace_id AND id = @agent_id;

-- name: CreateAgentQuery :one
INSERT INTO agents (workspace_id, widget_id, name)
VALUES (@workspace_id, @widget_id, @name)
RETURNING *;

-- name: WidgetExistsQuery :one
SELECT 1
FROM agents
WHERE widget_id = @widget_id;

-- name: UpsertVisitorQuery :exec
INSERT INTO visitors (id, widget_id)
VALUES (@visitor_id, @widget_id)
ON CONFLICT(id) DO UPDATE SET last_seen_at = CURRENT_TIMESTAMP;

-- name: UpsertConversationQuery :one
INSERT INTO conversations (visitor_id)
VALUES (@visitor_id)
ON CONFLICT(visitor_id) DO UPDATE SET updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: CreateMessageQuery :exec
INSERT INTO messages (conversation_id, role, text)
VALUES (@conversation_id, @role, @text);

-- name: TouchConversationQuery :exec
UPDATE conversations
SET updated_at = CURRENT_TIMESTAMP
WHERE id = @conversation_id;

-- name: ListMessagesQuery :many
SELECT sqlc.embed(messages)
FROM messages
WHERE conversation_id = @conversation_id
ORDER BY created_at, id;

-- name: DeleteSessionQuery :exec
DELETE FROM sessions
WHERE token = @token;

-- name: FindSessionQuery :one
SELECT data
FROM sessions
WHERE token = @token
AND expiry > @now;

-- name: CommitSessionQuery :exec
INSERT INTO sessions (token, data, expiry)
VALUES (@token, @data, @expiry)
ON CONFLICT(token) DO UPDATE SET data = excluded.data, expiry = excluded.expiry;

-- name: GetSourceWebsiteQuery :one
SELECT sqlc.embed(source_websites)
FROM source_websites
WHERE workspace_id = @workspace_id AND root_url = @root_url;

-- name: CreateSourceWebsiteQuery :one
INSERT INTO source_websites (workspace_id, root_url)
VALUES (@workspace_id, @root_url)
RETURNING *;

-- name: UpsertSourceWebsitePageQuery :exec
INSERT INTO source_website_pages (
    source_website_id,
    url,
    title,
    content,
    byte_count
)
VALUES (
    @source_website_id,
    @url,
    @title,
    @content,
    @byte_count
)
ON CONFLICT(source_website_id, url) DO UPDATE SET
    title = excluded.title,
    content = excluded.content,
    byte_count = excluded.byte_count,
    updated_at = CURRENT_TIMESTAMP;
