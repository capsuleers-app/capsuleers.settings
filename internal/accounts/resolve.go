package accounts

import "time"

// Evidence says that a character belongs to an account. At is when it was observed (zero = undated).
type Evidence struct {
	CharID string
	UserID string
	Source string
	At     time.Time
}

// tier ranks the sources: launcher log and guided identification (certain, newest wins) over
// strictly paired esi lines over file content.
func tier(source string) int {
	switch source {
	case SourceLauncher, SourceGuided:
		return 3
	case SourceLauncherESI:
		return 2
	}
	return 1
}

// Link is the resolved account of one character.
type Link struct {
	UserID string
	Source string
	At     time.Time
}

// Resolve picks one account per character: the highest tier present, then the newest evidence
// (character transfers). Two accounts with the same newest stamp = unknown (omitted).
// users and chars filter the evidence when not nil.
func Resolve(ev []Evidence, users, chars map[string]bool) map[string]Link {
	best := map[string]Link{}
	tied := map[string]bool{}
	for _, e := range ev {
		if e.CharID == "" || e.UserID == "" || (chars != nil && !chars[e.CharID]) || (users != nil && !users[e.UserID]) {
			continue
		}
		cur, ok := best[e.CharID]
		switch {
		case !ok || tier(e.Source) > tier(cur.Source) || (tier(e.Source) == tier(cur.Source) && e.At.After(cur.At)):
			best[e.CharID] = Link{e.UserID, e.Source, e.At}
			delete(tied, e.CharID)
		case tier(e.Source) == tier(cur.Source) && e.At.Equal(cur.At) && e.UserID != cur.UserID:
			tied[e.CharID] = true
		}
	}
	for c := range tied {
		delete(best, c)
	}
	return best
}
