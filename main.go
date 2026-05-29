// Command maple prints MapleStory monsters as ANSI art in the terminal.
package main

import (
	"embed"
	"flag"
	"fmt"
	"os"

	"github.com/JaeggerJose/maple-colorscripts/internal/sprite"
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
		if err != nil {
			fatal(err)
		}
	case *id != 0:
		mob, err = cat.ByID(*id)
		if err != nil {
			fatal(err)
		}
	default:
		mob = cat.Random()
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
