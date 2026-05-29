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
