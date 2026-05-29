// Package sprite loads embedded mob metadata + ANSI art and selects/renders it.
package sprite

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand"
	"strings"
)

// Mob is one embedded monster.
type Mob struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Level  int    `json:"level"`
	IsBoss bool   `json:"isBoss"`
	Slug   string `json:"slug"`
}

// Catalog holds the loaded mobs and the filesystem their art lives in.
type Catalog struct {
	fsys fs.FS
	mobs []Mob
}

// Load parses maple.json from fsys.
func Load(fsys fs.FS) (*Catalog, error) {
	data, err := fs.ReadFile(fsys, "maple.json")
	if err != nil {
		return nil, err
	}
	var mobs []Mob
	if err := json.Unmarshal(data, &mobs); err != nil {
		return nil, err
	}
	if len(mobs) == 0 {
		return nil, fmt.Errorf("no mobs in maple.json")
	}
	return &Catalog{fsys: fsys, mobs: mobs}, nil
}

func (c *Catalog) List() []Mob { return c.mobs }

func (c *Catalog) Random() Mob {
	return c.mobs[rand.Intn(len(c.mobs))]
}

func (c *Catalog) ByName(name string) (Mob, error) {
	want := strings.ToLower(strings.TrimSpace(name))
	for _, m := range c.mobs {
		if strings.ToLower(m.Name) == want {
			return m, nil
		}
	}
	return Mob{}, fmt.Errorf("no mob named %q", name)
}

func (c *Catalog) ByID(id int) (Mob, error) {
	for _, m := range c.mobs {
		if m.ID == id {
			return m, nil
		}
	}
	return Mob{}, fmt.Errorf("no mob with id %d", id)
}

// Render returns the ANSI art for a mob, with an optional trailing title line.
func (c *Catalog) Render(m Mob, showTitle bool) (string, error) {
	art, err := fs.ReadFile(c.fsys, "colorscripts/large/"+m.Slug)
	if err != nil {
		return "", err
	}
	out := string(art)
	if showTitle {
		boss := ""
		if m.IsBoss {
			boss = " (Boss)"
		}
		// Level <= 0 means no level (e.g. NPCs) — show just the name.
		if m.Level > 0 {
			out += fmt.Sprintf("\n%s  Lv.%d%s\n", m.Name, m.Level, boss)
		} else {
			out += fmt.Sprintf("\n%s%s\n", m.Name, boss)
		}
	}
	return out, nil
}
