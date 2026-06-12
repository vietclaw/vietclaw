package agent

import "strings"

func buildChildSessionIDs(parentSessionID, agentID string) (spawnPrefix, childSessionID string) {
	subID := newID("sub")
	spawnPrefix = parentSessionID + ":spawn:" + agentID + ":" + subID
	childSessionID = spawnPrefix + ":delegate:" + agentID
	return spawnPrefix, childSessionID
}

func ParentSessionID(sessionID string) string {
	idx := strings.Index(sessionID, ":spawn:")
	if idx < 0 {
		return sessionID
	}
	return sessionID[:idx]
}

func SpawnAgentID(sessionID string) string {
	idx := strings.Index(sessionID, ":spawn:")
	if idx < 0 {
		return ""
	}
	rest := sessionID[idx+len(":spawn:"):]
	end := strings.Index(rest, ":")
	if end <= 0 {
		return ""
	}
	return rest[:end]
}

func IsSpawnChildSession(sessionID string) bool {
	return strings.Contains(sessionID, ":spawn:")
}
