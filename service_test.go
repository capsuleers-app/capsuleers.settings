package main

import (
	"fmt"
	"testing"

	"github.com/capsuleers-app/capsuleers.settings/internal/backup"
)

func TestResultMapsErrors(t *testing.T) {
	if r := result(&backup.ApplyResult{Updated: nil}, nil); !r.OK || r.Updated == nil {
		t.Fatalf("ok result: %+v", r)
	}
	if r := result(nil, backup.ErrRunning); r.Code != "client_running" || r.OK {
		t.Fatalf("running: %+v", r)
	}
	r := result(nil, fmt.Errorf("%w: target 999 missing", backup.ErrBadRequest))
	if r.Code != "bad_request" || r.Message != "target 999 missing" {
		t.Fatalf("bad_request: %+v", r)
	}
	if r := result(nil, backup.ErrBadRequest); r.Message != "" {
		t.Fatalf("bare bad_request leaks code: %q", r.Message)
	}
}
