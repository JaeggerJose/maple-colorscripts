package maplestoryio

import "testing"

func TestURLBuilders(t *testing.T) {
	c := New("GMS", "255")
	if got := c.MetaURL(100000); got != "https://maplestory.io/api/GMS/255/mob/100000" {
		t.Errorf("MetaURL = %q", got)
	}
	if got := c.RenderURL(100000); got != "https://maplestory.io/api/GMS/255/mob/100000/render/stand" {
		t.Errorf("RenderURL = %q", got)
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
