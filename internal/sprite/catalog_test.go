package sprite

import (
	"strings"
	"testing"
	"testing/fstest"
)

func testFS() fstest.MapFS {
	return fstest.MapFS{
		"maple.json": {Data: []byte(`[
			{"id":100000,"name":"Snail","level":1,"isBoss":false,"slug":"snail"},
			{"id":100004,"name":"Orange Mushroom","level":6,"isBoss":false,"slug":"orange-mushroom"}
		]`)},
		"colorscripts/large/snail":           {Data: []byte("SNAIL_ART")},
		"colorscripts/large/orange-mushroom": {Data: []byte("MUSH_ART")},
	}
}

func TestLoadAndList(t *testing.T) {
	c, err := Load(testFS())
	if err != nil {
		t.Fatal(err)
	}
	if len(c.List()) != 2 {
		t.Fatalf("List len = %d, want 2", len(c.List()))
	}
}

func TestByNameCaseInsensitive(t *testing.T) {
	c, _ := Load(testFS())
	m, err := c.ByName("orange MUSHROOM")
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != 100004 {
		t.Errorf("ByName id = %d, want 100004", m.ID)
	}
}

func TestByNameUnknown(t *testing.T) {
	c, _ := Load(testFS())
	if _, err := c.ByName("nonexistent"); err == nil {
		t.Error("expected error for unknown name")
	}
}

func TestByID(t *testing.T) {
	c, _ := Load(testFS())
	m, err := c.ByID(100000)
	if err != nil || m.Name != "Snail" {
		t.Errorf("ByID = %+v, err=%v", m, err)
	}
}

func TestRenderWithTitle(t *testing.T) {
	c, _ := Load(testFS())
	m, _ := c.ByID(100000)
	out, err := c.Render(m, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "SNAIL_ART") || !strings.Contains(out, "Snail") {
		t.Errorf("Render with title missing art or name: %q", out)
	}
}

func TestRenderNoTitle(t *testing.T) {
	c, _ := Load(testFS())
	m, _ := c.ByID(100000)
	out, _ := c.Render(m, false)
	if strings.Contains(out, "Lv") {
		t.Errorf("Render no-title should omit title line: %q", out)
	}
}

func TestRandomNonEmpty(t *testing.T) {
	c, _ := Load(testFS())
	m := c.Random()
	if m.Slug == "" {
		t.Errorf("Random = %+v", m)
	}
}
