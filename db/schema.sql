----- USERS -----

-- Users are human support or admin members.
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX users_email_lower_idx ON users (lower(email));

-- Workspaces are customer accounts or teams using the app.
CREATE TABLE workspaces (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Workspace members are users with access to a workspace.
CREATE TABLE workspace_members (
    workspace_id INTEGER NOT NULL REFERENCES workspaces(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    role TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (workspace_id, user_id)
);

CREATE INDEX workspace_members_user_idx ON workspace_members (user_id);

-- Sessions store authenticated browser sessions.
CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    data BLOB NOT NULL,
    expiry DATETIME NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions (expiry);

-- Visitors are end-users opening the widget from a customer app.
CREATE TABLE visitors (
    id TEXT PRIMARY KEY,
    widget_id TEXT NOT NULL REFERENCES agents(widget_id),
    country_code TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Conversations are support threads between a visitor, human support, and agents.
CREATE TABLE conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    visitor_id TEXT NOT NULL REFERENCES visitors(id),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX conversations_visitor_id_idx ON conversations (visitor_id);

-- Messages are single chat entries inside a conversation.
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL REFERENCES conversations(id),
    role TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX messages_conversation_created_idx ON messages (conversation_id, created_at);


----- AGENTS -----

-- Agents are automated support agents configured inside a workspace.
CREATE TABLE agents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workspace_id INTEGER NOT NULL REFERENCES workspaces(id),
    widget_id TEXT NOT NULL,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX agents_workspace_idx ON agents (workspace_id);
CREATE UNIQUE INDEX agents_id_workspace_idx ON agents (id, workspace_id);
CREATE UNIQUE INDEX agents_widget_id_idx ON agents (widget_id);

----- SOURCES -----

-- Source QAs are question-answer sources.
CREATE TABLE source_qas (
    source_id INTEGER PRIMARY KEY AUTOINCREMENT,
    workspace_id INTEGER NOT NULL REFERENCES workspaces(id),
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Source websites are website crawl sources.
CREATE TABLE source_websites (
    source_id INTEGER PRIMARY KEY AUTOINCREMENT,
    workspace_id INTEGER NOT NULL REFERENCES workspaces(id),
    root_url TEXT NOT NULL,
    name TEXT,
    favicon_url TEXT,
    theme_color TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX source_websites_workspace_root_url_idx
ON source_websites (workspace_id, root_url);

-- Source website pages are crawled pages inside a website source.
CREATE TABLE source_website_pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_website_id INTEGER NOT NULL REFERENCES source_websites(source_id),
    url TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    byte_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX source_website_pages_website_url_idx
ON source_website_pages (source_website_id, url);

-- Source files are uploaded file sources.
CREATE TABLE source_files (
    source_id INTEGER PRIMARY KEY AUTOINCREMENT,
    workspace_id INTEGER NOT NULL REFERENCES workspaces(id),
    file_name TEXT NOT NULL,
    mime_type TEXT,
    content TEXT NOT NULL,
    byte_count INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
