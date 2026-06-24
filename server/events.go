package server

import (
	"syslantern/db"
)

type AgentCreatedEvent struct {
	TeamID  db.TeamID
	AgentID db.AgentID
}
