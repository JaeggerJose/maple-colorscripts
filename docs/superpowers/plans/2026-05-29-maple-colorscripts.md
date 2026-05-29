# maple-colorscripts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** A single-binary Go CLI that prints MapleStory monsters as truecolor ANSI art, with sprites pre-rendered from maplestory.io and embedded at compile time.

**Architecture:** Two layers in one Go module. A build tool (`cmd/build`) fetches monster GIFs from maplestory.io and converts the first frame to ANSI via `chafa`, writing `colorscripts/large/<slug>` + `maple.json`. The runtime CLI (`main.go` at repo root) embeds those assets with `go:embed` and only selects + prints. Selection/printing logic lives in `internal/sprite` and operates on an injected `fs.FS` so it is unit-testable without the real embed.

**Tech Stack:** Go 1.22, `go:embed`, `chafa` 1.18 (build-time only), maplestory.io API.

---

## File Structure

```
maple-colorscripts/
├── go.mod                          # module: maple-colorscripts
├── mobs.list                       # editable curation list (one mob id per line)
├── maple.json                      # build-generated metadata of embedded mobs
├── colorscripts/large/<slug>       # build-generated truecolor ANSI sprites
├── main.go                         # runtime CLI (package main, embeds assets)
├── internal/
│   ├── sprite/
│   │   ├── catalog.go              # Mob, Catalog, Load, Random, ByName, ByID, List, Render
│   │   └── catalog_test.go
│   └── maplestoryio/
│       ├── client.go               # Client, URL builders, FetchMeta, FetchRender
│       └── client_test.go
├── cmd/build/
│   ├── main.go                     # orchestration (network + chafa + write)
│   ├── render.go                   # pure helpers: Slug, ParseMobList, StripCursor
│   └── render_test.go
├── scripts/install.sh
└── README.md
```

**Module path note:** uses plain `maple-colorscripts` so it builds locally with no
GitHub username guess. When publishing, run
`go mod edit -module github.com/<you>/maple-colorscripts` and update imports.

---

## Task 1: Project scaffold

**Files:**
- Create: `go.mod`
- Create: `mobs.list`

- [ ] **Step 1: Initialize the Go module**

Run from repo root (`~/Downloads/maple-colorscripts`):
```bash
go mod init maple-colorscripts
```
Expected: creates `go.mod` containing `module maple-colorscripts` and `go 1.22`.

- [ ] **Step 2: Create the curation list with real classic mob ids**

Create `mobs.list`:
```
# maple-colorscripts curation list — one maplestory.io mob id per line.
# Blank lines and lines starting with # are ignored. Add ids and re-run the build.
100000   # Snail Lv1
100001   # Blue Snail Lv2
100002   # Red Snail Lv4
100004   # Orange Mushroom Lv6
100005   # Stump Lv4
100006   # Slime Lv7
100007   # Pig Lv7
1110100  # Green Mushroom Lv10
1110101  # Dark Stump Lv20
1120100  # Octopus Lv10
1130100  # Axe Stump Lv21
1210101  # Ribbon Pig Lv10
1210103  # Bubbling Lv10
2110200  # Horny Mushroom Lv12
2220000  # Mano Lv10 (BOSS)
2230100  # Evil Eye Lv26
2230101  # Zombie Mushroom Lv65
2230102  # Wild Boar Lv55
2230106  # Cico Lv76
2230110  # Wooden Mask Lv60
2300100  # Stirge Lv43
2600208  # Mushmom Lv122 (BOSS)
2700311  # Lupin Lv23
3210100  # Fire Boar Lv58
3230100  # Curse Eye Lv27
3230101  # Jr. Wraith Lv44
4090000  # Iron Hog Lv42
4130101  # Tortie Lv28
4230102  # Wraith Lv45
5200000  # Jr. Sentinel Lv76
```

- [ ] **Step 3: Commit**

```bash
git add go.mod mobs.list
git commit -m "chore: scaffold go module and curation list"
```

---

## Task 2: maplestory.io client (`internal/maplestoryio`)

**Files:**
- Create: `internal/maplestoryio/client.go`
- Test: `internal/maplestoryio/client_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/maplestoryio/client_test.go`:
```go
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
	body := []byte(`{"id":100004,"name":"Orange Mushroom","level":6,"isBoss":false}`)
	m, err := parseMeta(body)
	if err != nil {
		t.Fatal(err)
	}
	if m.ID != 100004 || m.Name != "Orange Mushroom" || m.Level != 6 || m.IsBoss {
		t.Errorf("parseMeta = %+v", m)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/maplestoryio/`
Expected: FAIL — `undefined: New`, `undefined: parseMeta`.

- [ ] **Step 3: Write the implementation**

Create `internal/maplestoryio/client.go`:
```go
// Package maplestoryio is a thin client for the maplestory.io mob API.
package maplestoryio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Meta is the subset of mob metadata this project uses.
type Meta struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Level  int    `json:"level"`
	IsBoss bool   `json:"isBoss"`
}

// Client targets a specific region+version of the API.
type Client struct {
	Region  string
	Version string
	HTTP    *http.Client
}

func New(region, version string) *Client {
	return &Client{Region: region, Version: version, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) base() string {
	return fmt.Sprintf("https://maplestory.io/api/%s/%s/mob", c.Region, c.Version)
}

func (c *Client) MetaURL(id int) string   { return fmt.Sprintf("%s/%d", c.base(), id) }
func (c *Client) RenderURL(id int) string { return fmt.Sprintf("%s/%d/render/stand", c.base(), id) }

func parseMeta(body []byte) (Meta, error) {
	var m Meta
	err := json.Unmarshal(body, &m)
	return m, err
}

// FetchMeta returns the mob's metadata.
func (c *Client) FetchMeta(id int) (Meta, error) {
	body, err := c.get(c.MetaURL(id))
	if err != nil {
		return Meta{}, err
	}
	return parseMeta(body)
}

// FetchRender returns the raw bytes of the mob's standing render (a GIF).
func (c *Client) FetchRender(id int) ([]byte, error) {
	return c.get(c.RenderURL(id))
}

func (c *Client) get(url string) ([]byte, error) {
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/maplestoryio/`
Expected: PASS (ok).

- [ ] **Step 5: Commit**

```bash
git add internal/maplestoryio/
git commit -m "feat: add maplestory.io api client"
```

---

## Task 3: Build helpers (`cmd/build/render.go`)

Pure, testable helpers used by the build orchestration.

**Files:**
- Create: `cmd/build/render.go`
- Test: `cmd/build/render_test.go`

- [ ] **Step 1: Write the failing tests**

Create `cmd/build/render_test.go`:
```go
package main

import (
	"reflect"
	"testing"
)

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Orange Mushroom": "orange-mushroom",
		"Jr. Necki":       "jr-necki",
		"  Red  Snail  ":  "red-snail",
		"Pig!!!":          "pig",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseMobList(t *testing.T) {
	in := "# header\n100000  # Snail\n\n  1110100\n# comment\n100004\n"
	got := ParseMobList(in)
	want := []int{100000, 1110100, 100004}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseMobList = %v, want %v", got, want)
	}
}

func TestStripCursor(t *testing.T) {
	in := "\x1b[?25l\x1b[38;2;1;2;3mX\x1b[0m\x1b[?25h"
	want := "\x1b[38;2;1;2;3mX\x1b[0m"
	if got := StripCursor(in); got != want {
		t.Errorf("StripCursor = %q, want %q", got, want)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./cmd/build/`
Expected: FAIL — `undefined: Slug`, `ParseMobList`, `StripCursor`.

- [ ] **Step 3: Write the implementation**

Create `cmd/build/render.go`:
```go
package main

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	nonAlnum   = regexp.MustCompile(`[^a-z0-9]+`)
	cursorCode = regexp.MustCompile(`\x1b\[\?25[lh]`)
)

// Slug converts a mob name into a filesystem-safe slug.
func Slug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonAlnum.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ParseMobList extracts mob ids from the curation file contents.
// Blank lines, # comment lines, and trailing # comments are ignored.
func ParseMobList(contents string) []int {
	var ids []int
	for _, line := range strings.Split(contents, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if id, err := strconv.Atoi(line); err == nil {
			ids = append(ids, id)
		}
	}
	return ids
}

// StripCursor removes the cursor show/hide escapes chafa emits.
func StripCursor(s string) string {
	return cursorCode.ReplaceAllString(s, "")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./cmd/build/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add cmd/build/render.go cmd/build/render_test.go
git commit -m "feat: add build helper functions"
```

---

## Task 4: Build orchestration (`cmd/build/main.go`)

Wires the client + chafa + helpers to produce `colorscripts/` and `maple.json`.
Network + chafa dependent, so it has no unit test; it is run manually in Task 6.

**Files:**
- Create: `cmd/build/main.go`

- [ ] **Step 1: Write the implementation**

Create `cmd/build/main.go`:
```go
// Command build fetches mobs from maplestory.io, renders the first frame of each
// to truecolor ANSI via chafa, and writes colorscripts/large/<slug> + maple.json.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"maple-colorscripts/internal/maplestoryio"
)

const (
	region  = "GMS"
	version = "255"
	size    = "40x40"
)

// embedMeta mirrors the maple.json schema consumed by internal/sprite.
type embedMeta struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Level  int    `json:"level"`
	IsBoss bool   `json:"isBoss"`
	Slug   string `json:"slug"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "build:", err)
		os.Exit(1)
	}
}

func run() error {
	listBytes, err := os.ReadFile("mobs.list")
	if err != nil {
		return err
	}
	ids := ParseMobList(string(listBytes))
	if len(ids) == 0 {
		return fmt.Errorf("mobs.list has no ids")
	}

	outDir := filepath.Join("colorscripts", "large")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	client := maplestoryio.New(region, version)
	usedSlug := map[string]bool{}
	var metas []embedMeta
	ok, skipped := 0, 0

	for _, id := range ids {
		meta, err := client.FetchMeta(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %d: meta: %v\n", id, err)
			skipped++
			continue
		}
		gif, err := client.FetchRender(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %d (%s): render: %v\n", id, meta.Name, err)
			skipped++
			continue
		}
		ansi, err := renderANSI(gif)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skip %d (%s): chafa: %v\n", id, meta.Name, err)
			skipped++
			continue
		}

		slug := Slug(meta.Name)
		if slug == "" {
			slug = fmt.Sprintf("mob-%d", id)
		}
		if usedSlug[slug] {
			slug = fmt.Sprintf("%s-%d", slug, id)
		}
		usedSlug[slug] = true

		if err := os.WriteFile(filepath.Join(outDir, slug), []byte(ansi), 0o644); err != nil {
			return err
		}
		metas = append(metas, embedMeta{ID: meta.ID, Name: meta.Name, Level: meta.Level, IsBoss: meta.IsBoss, Slug: slug})
		ok++
		fmt.Printf("ok %d %s -> %s\n", id, meta.Name, slug)
	}

	out, err := json.MarshalIndent(metas, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile("maple.json", append(out, '\n'), 0o644); err != nil {
		return err
	}
	fmt.Printf("done: %d ok, %d skipped\n", ok, skipped)
	return nil
}

// renderANSI writes the gif to a temp file and runs chafa to produce ANSI art.
func renderANSI(gif []byte) (string, error) {
	tmp, err := os.CreateTemp("", "mob-*.gif")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(gif); err != nil {
		return "", err
	}
	tmp.Close()

	cmd := exec.Command("chafa",
		"--format", "symbols", "--animate", "off",
		"-c", "full", "--dither", "none", "-w", "9",
		"--size", size, tmp.Name())
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return StripCursor(string(out)), nil
}
```

- [ ] **Step 2: Verify it compiles**

Run: `go build ./cmd/build/`
Expected: no output, exit 0.

- [ ] **Step 3: Commit**

```bash
git add cmd/build/main.go
git commit -m "feat: add build orchestration pipeline"
```

---

## Task 5: Sprite catalog (`internal/sprite`)

Selection + rendering over an injected `fs.FS`, unit-tested with `fstest.MapFS`.

**Files:**
- Create: `internal/sprite/catalog.go`
- Test: `internal/sprite/catalog_test.go`

- [ ] **Step 1: Write the failing tests**

Create `internal/sprite/catalog_test.go`:
```go
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
	m, err := c.Random()
	if err != nil || m.Slug == "" {
		t.Errorf("Random = %+v, err=%v", m, err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/sprite/`
Expected: FAIL — `undefined: Load`.

- [ ] **Step 3: Write the implementation**

Create `internal/sprite/catalog.go`:
```go
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

func (c *Catalog) Random() (Mob, error) {
	return c.mobs[rand.Intn(len(c.mobs))], nil
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
		out += fmt.Sprintf("\n%s  Lv.%d%s\n", m.Name, m.Level, boss)
	}
	return out, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/sprite/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/sprite/
git commit -m "feat: add sprite catalog with selection and rendering"
```

---

## Task 6: Runtime CLI (`main.go`)

**Files:**
- Create: `main.go`

Note: this task's `go:embed` requires `maple.json` and `colorscripts/` to exist.
They are generated in Task 7. To let `main.go` compile before then, Step 1 creates
minimal placeholder assets that Task 7 overwrites with real data.

- [ ] **Step 1: Create placeholder assets so the embed compiles**

```bash
mkdir -p colorscripts/large
printf 'PLACEHOLDER\n' > colorscripts/large/placeholder
printf '[{"id":0,"name":"Placeholder","level":1,"isBoss":false,"slug":"placeholder"}]\n' > maple.json
```

- [ ] **Step 2: Write the implementation**

Create `main.go`:
```go
// Command maple prints MapleStory monsters as ANSI art in the terminal.
package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"maple-colorscripts/internal/sprite"
)

//go:embed maple.json
//go:embed colorscripts
var assets embed.FS

func main() {
	name := flag.String("name", "", "print mob by name")
	flag.StringVar(name, "n", "", "print mob by name (shorthand)")
	id := flag.Int("id", 0, "print mob by id")
	flag.IntVar(id, "i", 0, "print mob by id (shorthand)")
	list := flag.Bool("list", false, "list all mobs")
	flag.BoolVar(list, "l", false, "list all mobs (shorthand)")
	noTitle := flag.Bool("no-title", false, "do not print the name line")
	flag.Bool("r", false, "print a random mob (default behavior)")
	flag.Parse()

	cat, err := sprite.Load(assets)
	if err != nil {
		fatal(err)
	}

	if *list {
		for _, m := range cat.List() {
			boss := ""
			if m.IsBoss {
				boss = " (Boss)"
			}
			fmt.Printf("%-20s id=%-8d Lv.%d%s\n", m.Name, m.ID, m.Level, boss)
		}
		return
	}

	var mob sprite.Mob
	switch {
	case *name != "":
		mob, err = cat.ByName(*name)
	case *id != 0:
		mob, err = cat.ByID(*id)
	default:
		mob, err = cat.Random()
	}
	if err != nil {
		fatal(err)
	}

	art, err := cat.Render(mob, !*noTitle)
	if err != nil {
		fatal(err)
	}
	fmt.Print(art)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "maple:", err)
	os.Exit(1)
}
```

- [ ] **Step 3: Verify it builds and runs against placeholder data**

Run:
```bash
go build -o maple . && ./maple --list
```
Expected: builds, prints `Placeholder           id=0        Lv.1`.

- [ ] **Step 4: Commit**

```bash
git add main.go maple.json colorscripts/large/placeholder
git commit -m "feat: add runtime cli with embedded assets"
```

---

## Task 7: Generate real sprites and finalize

**Files:**
- Modify: `colorscripts/large/*` (generated), `maple.json` (generated)
- Create: `scripts/install.sh`, `README.md`

- [ ] **Step 1: Remove placeholder assets**

```bash
rm -f colorscripts/large/placeholder
```

- [ ] **Step 2: Run the build pipeline for real**

Run from repo root (needs network + chafa on PATH):
```bash
go run ./cmd/build
```
Expected: a series of `ok <id> <name> -> <slug>` lines and a final
`done: N ok, M skipped` (N should be ~30). `maple.json` and
`colorscripts/large/<slug>` files are populated.

- [ ] **Step 3: Verify the CLI renders real mobs**

```bash
go build -o maple . && ./maple -n "Orange Mushroom" && ./maple --list | head
```
Expected: a colorful ANSI mushroom + a title line `Orange Mushroom  Lv.6`, then a
list of ~30 mobs.

- [ ] **Step 4: Write the install script**

Create `scripts/install.sh`:
```bash
#!/bin/bash
# Build the maple binary and install it to ~/.local/bin (override with PREFIX).
set -euo pipefail
PREFIX="${PREFIX:-$HOME/.local/bin}"
mkdir -p "$PREFIX"
go build -o "$PREFIX/maple" .
echo "Installed maple to $PREFIX/maple"
echo "Ensure $PREFIX is on your PATH."
```
Then: `chmod +x scripts/install.sh`

- [ ] **Step 5: Write the README**

Create `README.md`:
```markdown
# maple-colorscripts

Print MapleStory monsters as truecolor ANSI art in your terminal.

## Install

    ./scripts/install.sh        # builds and installs `maple` to ~/.local/bin

## Usage

    maple                 # random monster
    maple -n "Snail"      # by name
    maple -i 100004       # by id
    maple --list          # list embedded monsters
    maple --no-title      # omit the name line (for prompts/statuslines)

## Add more monsters

Sprites are sourced from [maplestory.io](https://maplestory.io) and embedded at
build time. To add monsters, append their mob ids to `mobs.list`, then:

    go run ./cmd/build     # re-fetch + re-render (needs `chafa` + network)
    ./scripts/install.sh   # rebuild the binary

Find mob ids in the API: `https://maplestory.io/api/GMS/255/mob`.

## How it works

- `cmd/build` fetches each mob's standing GIF, converts the first frame to ANSI
  with `chafa`, and writes `colorscripts/large/<slug>` + `maple.json`.
- `main.go` embeds those assets with `go:embed`; the runtime CLI only selects and
  prints — zero dependencies, instant startup.
```

- [ ] **Step 6: Verify full coverage and build**

Run:
```bash
go test ./... && go vet ./... && go build -o maple .
```
Expected: all tests PASS, vet clean, binary builds.

- [ ] **Step 7: Commit**

```bash
git add -A
git commit -m "feat: generate sprites, add install script and readme"
```

---

## Self-Review Notes

- **Spec coverage:** data source (Task 2), curated pre-render (Tasks 4+7),
  extensible list file (Task 1 `mobs.list`), static output (chafa `--animate off`
  in Task 4), Go + go:embed (Task 6), CLI flags random/name/id/list/no-title
  (Task 6), build/runtime two-layer split (Tasks 2-4 vs 5-6), error handling
  (skip-and-continue in Task 4, stderr+exit in Task 6), tests (Tasks 2,3,5).
- **No live network in unit tests:** client tests parse a literal payload; sprite
  tests use `fstest.MapFS`; the only network step is the manual build in Task 7.
- **Type consistency:** `maple.json` schema (`id,name,level,isBoss,slug`) is
  identical in `cmd/build` `embedMeta` and `internal/sprite` `Mob`.
