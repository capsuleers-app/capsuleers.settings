// Package esi resolves character ids to names through the public, unauthenticated
// POST /universe/names/ endpoint, with an on-disk cache.
package esi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const (
	AppName    = "capsuleers.settings"
	AppVersion = "0.1.0"
	RepoURL    = "https://github.com/WilliamFalci/capsuleers.settings"
	// UserAgent is the only User-Agent this app ever sends.
	UserAgent = AppName + "/" + AppVersion + " (+" + RepoURL + ")"
	// CompatibilityDate pins the ESI route set (the /latest/ alias hides modern routes).
	CompatibilityDate = "2026-10-28"

	namesURL  = "https://esi.evetech.net/universe/names/"
	batchSize = 500
	cacheTTL  = 30 * 24 * time.Hour
)

type entry struct {
	Name string `json:"name"`
	At   int64  `json:"at"`
}

// Resolver caches names in <dir>/names.json.
type Resolver struct {
	mu     sync.Mutex
	path   string
	url    string
	client *http.Client
	cache  map[string]entry
}

func NewResolver(dir string) *Resolver {
	r := &Resolver{
		path:   filepath.Join(dir, "names.json"),
		url:    namesURL,
		client: &http.Client{Timeout: 10 * time.Second},
		cache:  map[string]entry{},
	}
	if b, err := os.ReadFile(r.path); err == nil {
		_ = json.Unmarshal(b, &r.cache)
	}
	return r
}

// Cached returns names known without touching the network (stale entries included).
func (r *Resolver) Cached(ids []string) map[string]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[string]string{}
	for _, id := range ids {
		if e, ok := r.cache[id]; ok {
			out[id] = e.Name
		}
	}
	return out
}

// Resolve returns names for ids, fetching missing or stale ones. Offline or failed lookups
// simply leave ids out of the result.
func (r *Resolver) Resolve(ctx context.Context, ids []string) map[string]string {
	now := time.Now()
	var need []int64
	r.mu.Lock()
	for _, id := range ids {
		n, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			continue
		}
		if e, ok := r.cache[id]; !ok || now.Sub(time.Unix(e.At, 0)) > cacheTTL {
			need = append(need, n)
		}
	}
	r.mu.Unlock()
	for i := 0; i < len(need); i += batchSize {
		end := min(i+batchSize, len(need))
		r.fetch(ctx, need[i:end])
	}
	r.mu.Lock()
	r.save()
	r.mu.Unlock()
	return r.Cached(ids)
}

// fetch posts one batch. ESI answers 404 for the whole batch when one id is unknown, so a
// failing batch is split until the bad id is isolated.
func (r *Resolver) fetch(ctx context.Context, ids []int64) {
	if len(ids) == 0 {
		return
	}
	body, _ := json.Marshal(ids)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-Compatibility-Date", CompatibilityDate)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := r.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound && len(ids) > 1 {
		h := len(ids) / 2
		r.fetch(ctx, ids[:h])
		r.fetch(ctx, ids[h:])
		return
	}
	if resp.StatusCode != http.StatusOK {
		return
	}
	var res []struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		Category string `json:"category"`
	}
	if json.NewDecoder(resp.Body).Decode(&res) != nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range res {
		if x.Category == "character" && x.Name != "" {
			r.cache[strconv.FormatInt(x.ID, 10)] = entry{Name: x.Name, At: time.Now().Unix()}
		}
	}
}

func (r *Resolver) save() {
	b, err := json.Marshal(r.cache)
	if err != nil || os.MkdirAll(filepath.Dir(r.path), 0o700) != nil {
		return
	}
	tmp := r.path + ".tmp"
	if os.WriteFile(tmp, b, 0o600) == nil {
		_ = os.Rename(tmp, r.path)
	}
}
