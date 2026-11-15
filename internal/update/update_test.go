package update

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestVersions(t *testing.T) {
	cases := []struct {
		cur, lat string
		want     bool
	}{
		{"1.2.9", "v1.2.10", true}, {"1.2.10", "1.2.9", false}, {"1.2.3", "1.2.3", false},
		{"1.9.1", "1.10.0", true}, {"dev", "1.0.1", false}, {"0.1.0", "x", false}, {"1.2", "1.2.3", false},
	}
	for _, c := range cases {
		if got := Newer(c.cur, c.lat); got != c.want {
			t.Errorf("Newer(%q,%q)=%v", c.cur, c.lat, got)
		}
	}
}

func TestTarget(t *testing.T) {
	if _, _, ok := Target("linux", "", "/x"); ok {
		t.Error("linux without APPIMAGE must be skipped")
	}
	if a, f, ok := Target("linux", "/a/b.AppImage", "/x"); !ok || a != AssetLinux || f != "/a/b.AppImage" {
		t.Error("linux appimage")
	}
	if a, _, ok := Target("windows", "", "/x.exe"); !ok || a != AssetWindows {
		t.Error("windows")
	}
	if _, _, ok := Target("darwin", "", "/x"); ok {
		t.Error("darwin skipped")
	}
}

func server(t *testing.T, tag, digest string, body []byte) *httptest.Server {
	var srv *httptest.Server
	mux := http.NewServeMux()
	mux.HandleFunc("/latest", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "ua-test" {
			http.Error(w, "ua", 400)
			return
		}
		fmt.Fprintf(w, `{"tag_name":%q,"assets":[{"name":%q,"browser_download_url":%q,"digest":%q,"size":%d}]}`,
			tag, AssetLinux, srv.URL+"/dl", digest, len(body))
	})
	mux.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) { w.Write(body) })
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func sum(b []byte) string { s := sha256.Sum256(b); return "sha256:" + hex.EncodeToString(s[:]) }

func updater(srv *httptest.Server, cur string) *Updater {
	return &Updater{Current: cur, UserAgent: "ua-test", APIURL: srv.URL + "/latest", Client: srv.Client()}
}

func TestCheckDownloadsAndSwaps(t *testing.T) {
	body := []byte("new binary")
	srv := server(t, "v1.2.4", sum(body), body)
	dir := t.TempDir()
	app := filepath.Join(dir, "app.AppImage")
	os.WriteFile(app, []byte("old"), 0o755)
	p, err := updater(srv, "1.2.3").Check(context.Background(), "linux", app, "")
	if err != nil || p == nil || p.Version != "1.2.4" {
		t.Fatalf("check: %v %v", p, err)
	}
	if filepath.Dir(p.Tmp) != dir {
		t.Error("temp must sit next to target")
	}
	if err := p.Apply(); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(app)
	if string(got) != "new binary" {
		t.Errorf("content %q", got)
	}
	if st, _ := os.Stat(app); st.Mode().Perm() != 0o755 {
		t.Error("mode")
	}
}

func TestCheckSilentCases(t *testing.T) {
	body := []byte("x")
	dir := t.TempDir()
	app := filepath.Join(dir, "a.AppImage")
	srv := server(t, "v1.2.3", sum(body), body)
	if p, err := updater(srv, "1.2.3").Check(context.Background(), "linux", app, ""); p != nil || err != nil {
		t.Error("same version")
	}
	bad := server(t, "v1.2.4", sum([]byte("other")), body)
	if p, err := updater(bad, "1.2.3").Check(context.Background(), "linux", app, ""); p != nil || err == nil {
		t.Error("digest mismatch must fail")
	}
	none := server(t, "v1.2.4", "", body)
	if p, err := updater(none, "1.2.3").Check(context.Background(), "linux", app, ""); p != nil || err == nil {
		t.Error("missing digest must fail")
	}
	if es, _ := os.ReadDir(dir); len(es) != 0 {
		t.Errorf("temp files left: %v", es)
	}
}

func TestSizeCap(t *testing.T) {
	body := []byte("0123456789")
	srv := server(t, "v1.2.4", sum(body), body)
	u := updater(srv, "1.2.3")
	u.MaxSize = 5
	app := filepath.Join(t.TempDir(), "a.AppImage")
	if p, err := u.Check(context.Background(), "linux", app, ""); p != nil || err == nil {
		t.Error("oversize must fail")
	}
}

func TestSwapWindowsAndCleanup(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "app.exe")
	tmp := filepath.Join(dir, tmpPrefix+"1"+tmpSuffix)
	os.WriteFile(exe, []byte("old"), 0o755)
	os.WriteFile(tmp, []byte("new"), 0o644)
	if err := Swap(tmp, exe, true); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "new" {
		t.Error("new not in place")
	}
	if b, _ := os.ReadFile(exe + ".old"); string(b) != "old" {
		t.Error(".old missing")
	}
	orphan := filepath.Join(dir, tmpPrefix+"2"+tmpSuffix)
	os.WriteFile(orphan, nil, 0o644)
	Cleanup(exe)
	if es, _ := os.ReadDir(dir); len(es) != 1 {
		t.Errorf("cleanup left %v", es)
	}
}

func TestTextsAndLang(t *testing.T) {
	if Lang("it_IT.UTF-8") != "it" || Lang("en_US") != "en" || Lang("") != "en" {
		t.Error("lang")
	}
	if Texts("it", "1.2.3").Restart != "Riavvia e installa" || Texts("en", "1.2.3").Later != "Later" {
		t.Error("texts")
	}
}
