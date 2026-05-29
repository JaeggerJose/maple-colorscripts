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
	metas := []embedMeta{}
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

	if ok == 0 {
		return fmt.Errorf("all %d mobs skipped; refusing to write empty maple.json", skipped)
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
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

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
