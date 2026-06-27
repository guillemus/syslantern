-- name: GetTeamByID :one
SELECT teams.*  
FROM teams
WHERE id = @id;

-- name: GetTeamByAgentAPIKey :one
SELECT teams.* 
FROM teams
JOIN agents ON agents.team_id = teams.id
WHERE agents.api_key = @agent_api_key;

-- name: createTeam :one
INSERT INTO teams (name)
VALUES (@name)
RETURNING *;
