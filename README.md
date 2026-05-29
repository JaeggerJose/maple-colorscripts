# maple-colorscripts

Print MapleStory (新楓之谷) monsters as **truecolor ANSI art** in your terminal —
a [pokemon-colorscripts](https://gitlab.com/phoneybadger/pokemon-colorscripts)
for MapleStory mobs.

Sprites are sourced from the [maplestory.io](https://maplestory.io) API,
pre-rendered to ANSI at build time, and **embedded into a single Go binary** — so
the installed CLI has zero runtime dependencies and starts instantly.

```
$ maple -n "Orange Mushroom"
   <colorful ANSI mushroom>
Orange Mushroom  Lv.6
```

## Requirements

- **To run:** nothing — the binary is self-contained.
- **To build/install from source:** [Go](https://go.dev/dl/) 1.22+.
- **To regenerate sprites** (only if you add monsters): `Go` + [`chafa`](https://hpjansson.org/chafa/) + network.
- **Best display:** a terminal with truecolor (e.g. iTerm2, Kitty, WezTerm,
  modern GNOME Terminal). For the sharpest output, use a font that includes the
  "Symbols for Legacy Computing" block (e.g. Cascadia Code, Iosevka, JuliaMono).

## Install

### Option A — `go install` (recommended)

```bash
go install github.com/JaeggerJose/maple-colorscripts@latest
```

This builds and drops the `maple-colorscripts` binary in `$(go env GOPATH)/bin`.
Make sure that directory is on your `PATH`. Optionally shorten the name:

```bash
ln -sf "$(go env GOPATH)/bin/maple-colorscripts" "$(go env GOPATH)/bin/maple"
```

### Option B — clone + install script

```bash
git clone https://github.com/JaeggerJose/maple-colorscripts.git
cd maple-colorscripts
./scripts/install.sh          # builds `maple` into ~/.local/bin (override with PREFIX=...)
```

Ensure your install dir is on `PATH`, e.g. add to `~/.zshrc`:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### Option C — clone + build manually

```bash
git clone https://github.com/JaeggerJose/maple-colorscripts.git
cd maple-colorscripts
go build -o maple .
./maple
```

## Usage

```bash
maple                  # random monster
maple -n "Snail"       # by name (case-insensitive)
maple -i 100004        # by id
maple --list           # list all embedded monsters (name, id, level)
maple --no-title       # omit the name line (handy for prompts / statuslines)
```

### Show one on every new shell

Add to `~/.zshrc` (or `~/.bashrc`):

```bash
maple
```

## Add more monsters

The embedded set is driven by `mobs.list` — one maplestory.io mob id per line
(`#` comments allowed). To add monsters, append their ids and regenerate:

```bash
echo "2300100  # Stirge" >> mobs.list
go run ./cmd/build         # re-fetch + re-render (needs chafa + network)
./scripts/install.sh       # rebuild + reinstall the binary
```

Browse mob ids in the API: <https://maplestory.io/api/GMS/255/mob>

## How it works

Two layers in one Go module:

- **Build layer** (`cmd/build`) — reads `mobs.list`, fetches each mob's standing
  GIF + metadata from maplestory.io, converts the first frame to truecolor ANSI
  with `chafa`, and writes `colorscripts/large/<slug>` + `maple.json`.
- **Runtime layer** (`main.go`) — embeds those assets via `go:embed`. The CLI
  only selects a mob and prints it: no filesystem lookups, no network, no deps.

Selection/rendering logic lives in `internal/sprite`; the API client in
`internal/maplestoryio`.

## License

See [LICENSE](LICENSE) if present; otherwise all rights reserved by the author.
