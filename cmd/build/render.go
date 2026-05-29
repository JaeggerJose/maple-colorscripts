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
