package esi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveHeadersCacheAndSplit(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("User-Agent") != UserAgent || r.Header.Get("X-Compatibility-Date") != CompatibilityDate {
			t.Errorf("bad headers: %v", r.Header)
		}
		var ids []int64
		_ = json.NewDecoder(r.Body).Decode(&ids)
		for _, id := range ids {
			if id == 666 { // unknown id: ESI fails the whole batch
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
		}
		var out []map[string]any
		for _, id := range ids {
			out = append(out, map[string]any{"id": id, "name": "Pilot", "category": "character"})
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
	defer srv.Close()
	dir := t.TempDir()
	r := NewResolver(dir)
	r.url = srv.URL
	got := r.Resolve(context.Background(), []string{"1", "666", "3", "x"})
	if got["1"] != "Pilot" || got["3"] != "Pilot" || len(got) != 2 {
		t.Fatalf("got=%v", got)
	}
	before := calls
	if r2 := NewResolver(dir); len(r2.Cached([]string{"1", "3"})) != 2 {
		t.Fatal("cache not persisted")
	}
	r.Resolve(context.Background(), []string{"1", "3"})
	if calls != before {
		t.Fatal("cached ids refetched")
	}
}
