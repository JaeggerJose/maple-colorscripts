package maplestoryio

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestURLBuilders(t *testing.T) {
	c := New("GMS", "255")
	if got := c.MetaURL(100000); got != "https://maplestory.io/api/GMS/255/mob/100000" {
		t.Errorf("MetaURL = %q", got)
	}
	if got := c.RenderURL(100000); got != "https://maplestory.io/api/GMS/255/mob/100000/render/stand" {
		t.Errorf("RenderURL = %q", got)
	}
}

func TestFetchMetaHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/mob/100004") {
			w.Write([]byte(`{"id":100004,"name":"Orange Mushroom","meta":{"level":6}}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	c := New("GMS", "255")
	c.BaseURL = srv.URL
	m, err := c.FetchMeta(100004)
	if err != nil {
		t.Fatal(err)
	}
	if m.Name != "Orange Mushroom" || m.Level != 6 {
		t.Errorf("FetchMeta = %+v", m)
	}
	if _, err := c.FetchMeta(999999); err == nil {
		t.Error("expected error for 404")
	}
}

func TestParseMetaBadJSON(t *testing.T) {
	if _, err := parseMeta([]byte("not json")); err == nil {
		t.Error("expected error for invalid json")
	}
}

func TestParseMeta(t *testing.T) {
	body := []byte(`{"id":100004,"name":"Orange Mushroom","meta":{"level":6}}`)
	m, err := parseMeta(body)
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != 100004 || m.Name != "Orange Mushroom" || m.Level != 6 || m.IsBoss {
		t.Errorf("parseMeta = %+v", m)
	}
	boss := []byte(`{"id":2220000,"name":"Mano","meta":{"level":10,"isBoss":true}}`)
	b, err := parseMeta(boss)
	if err != nil {
		t.Fatal(err)
	}
	if !b.IsBoss || b.Level != 10 {
		t.Errorf("boss parse = %+v", b)
	}
}
