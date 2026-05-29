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
