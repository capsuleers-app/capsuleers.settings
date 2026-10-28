package main

import (
	"errors"
	"sync"
	"time"

	"github.com/capsuleers-app/capsuleers.settings/internal/accounts"
	"github.com/capsuleers-app/capsuleers.settings/internal/config"
)

const (
	identInterval = time.Second     // how often the profile files are looked at
	identSettle   = 3 * time.Second // changes this old without news are considered complete
)

// IdentFound is a character identified by the guided procedure.
type IdentFound struct {
	CharID   string `json:"charId"`
	UserID   string `json:"userId"`
	Remapped bool   `json:"remapped"` // differs from the account saved before
}

// IdentifyState is polled by the frontend while the identification panel is open.
type IdentifyState struct {
	Active    bool         `json:"active"`
	Seq       int          `json:"seq"`       // grows each time a mapping was saved
	Ambiguous int          `json:"ambiguous"` // batches discarded because several accounts changed
	Found     []IdentFound `json:"found"`
}

type identSession struct {
	dir  string
	stop chan struct{}
	done chan struct{}
	mu   sync.Mutex
	st   IdentifyState
}

// IdentifyStart begins the guided identification of a profile: it snapshots the character and
// account files (read-only) and watches them; when the changes settle, the characters that
// changed together with exactly one account file are mapped to it and saved in the app config.
func (s *Service) IdentifyStart(profileDir string) error {
	s.IdentifyStop()
	base, err := accounts.ReadSnapshot(profileDir)
	if err != nil {
		return errors.New("cannot read the profile folder")
	}
	sess := &identSession{dir: profileDir, stop: make(chan struct{}), done: make(chan struct{})}
	sess.st = IdentifyState{Active: true, Found: []IdentFound{}}
	s.mu.Lock()
	s.ident = sess
	s.mu.Unlock()
	go s.watch(sess, accounts.NewTracker(base, identSettle))
	return nil
}

func (s *Service) watch(sess *identSession, tr *accounts.Tracker) {
	defer close(sess.done)
	tick := time.NewTicker(identInterval)
	defer tick.Stop()
	for {
		select {
		case <-sess.stop:
			return
		case now := <-tick.C:
			cur, err := accounts.ReadSnapshot(sess.dir)
			if err != nil {
				continue
			}
			if out, ok := tr.Observe(cur, now); ok {
				s.record(sess, out, now)
			}
		}
	}
}

func (s *Service) record(sess *identSession, out accounts.Outcome, now time.Time) {
	if out.Kind == accounts.PairAmbiguous {
		sess.mu.Lock()
		sess.st.Ambiguous++
		sess.mu.Unlock()
		return
	}
	prev, _ := s.effective(sess.dir)
	s.mu.Lock()
	var found []IdentFound
	for _, c := range out.Chars {
		remapped := false
		if l, ok := prev[c]; ok && l.UserID != out.UserID {
			remapped = true
		}
		if err := s.cfg.SetAccountGuided(sess.dir, c, config.Link{UserID: out.UserID, At: now, Source: accounts.SourceGuided}); err != nil {
			continue
		}
		found = append(found, IdentFound{CharID: c, UserID: out.UserID, Remapped: remapped})
	}
	s.mu.Unlock()
	sess.mu.Lock()
	defer sess.mu.Unlock()
	for _, f := range found {
		replaced := false
		for i := range sess.st.Found {
			if sess.st.Found[i].CharID == f.CharID {
				sess.st.Found[i], replaced = f, true
			}
		}
		if !replaced {
			sess.st.Found = append(sess.st.Found, f)
		}
	}
	if len(found) > 0 {
		sess.st.Seq++
	}
}

// IdentifyPoll returns what the running identification found so far (inactive when none runs).
func (s *Service) IdentifyPoll() IdentifyState {
	s.mu.Lock()
	sess := s.ident
	s.mu.Unlock()
	if sess == nil {
		return IdentifyState{Found: []IdentFound{}}
	}
	sess.mu.Lock()
	defer sess.mu.Unlock()
	st := sess.st
	st.Found = append([]IdentFound{}, st.Found...)
	return st
}

// IdentifyStop ends the identification (panel closed, FINE pressed, view or profile changed).
func (s *Service) IdentifyStop() {
	s.mu.Lock()
	sess := s.ident
	s.ident = nil
	s.mu.Unlock()
	if sess != nil {
		close(sess.stop)
		<-sess.done
	}
}
