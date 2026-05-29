# maple-colorscripts — Design Spec

**Date:** 2026-05-29
**Status:** Approved direction, pending final spec review

## Summary

A `pokemon-colorscripts`-style CLI that prints MapleStory ("新楓之谷") monsters
as truecolor ANSI art in the terminal. Sprites are sourced from the
[maplestory.io](https://maplestory.io) API, pre-rendered to ANSI text at build
time, embedded into a single Go binary, and printed instantly at runtime.

## Goals

- Print a MapleStory monster sprite in the terminal (random or by name/id).
- Instant startup (suitable for `.zshrc` greeting and statusline use).
- Single self-contained binary, zero runtime dependencies.
- Curated starter set of classic monsters, trivially extensible via one list file.

## Non-Goals (YAGNI)

- No animation playback (static first frame only).
- No `small` size variant (single `large` size).
- No runtime network access (all sprites baked in at build time).
- Not the full 11,497-mob catalog — a curated, extensible subset.

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Data source | maplestory.io API | 11,497 mobs with `{id,name,level,isBoss,mobType}`; `render/stand` returns a small GIF sprite. Verified working (GMS/255). |
| Sprite storage | Curated, pre-rendered ANSI, committed | Offline, zero runtime deps — the pokemon-colorscripts model. |
| Curation | Starter classic set in one editable list file | Add an id, rebuild, commit. Low maintenance cost. |
| Output | Static (first frame), print-once | Pipe-able, usable as a startup banner. |
| Language | **Go** | Single static binary, ~ms startup (critical for terminal-launch). `go:embed` bakes sprites into the binary. |
| Distribution | Single binary via `go:embed` | No install-path/dependency concerns. |

## Architecture (two layers)

```
BUILD layer (dev-time: needs chafa + network)      RUNTIME layer (zero-dep)
  mobs.list ──► cmd/build ──► maplestory.io           cmd/maple
   (maintained)     │           (fetch GIF + name)       │  select → print
                    └──► chafa (first frame → ANSI)      │
                            └──► colorscripts/large/<slug>
                                   + maple.json  ──(go:embed)──► binary
```

- **Build layer** (`cmd/build`): reads `mobs.list`, fetches each mob's metadata
  and `render/stand` GIF, runs `chafa` to convert the first frame to truecolor
  ANSI (default symbol set — the recipe validated in the statusline work), writes
  `colorscripts/large/<slug>` and regenerates `maple.json`.
- **Runtime layer** (`cmd/maple`): the user-facing CLI. Sprites + `maple.json`
  are embedded via `go:embed`; the CLI only selects and prints. No filesystem
  lookup, no network.

## Repo Structure

```
maple-colorscripts/
├── README.md
├── LICENSE
├── go.mod
├── mobs.list                     # ★ editable curation list: one mob id per line (# comments ok)
├── maple.json                    # build-generated: [{id,name,level,isBoss,slug}] of embedded mobs
├── colorscripts/
│   └── large/<slug>              # pre-rendered truecolor ANSI text, one file per mob
├── cmd/
│   ├── maple/main.go             # runtime CLI (embeds colorscripts/ + maple.json)
│   └── build/main.go             # dev build pipeline (fetch + chafa + write)
├── internal/
│   ├── sprite/                   # selection + printing logic (testable)
│   └── maplestoryio/             # API client (list, metadata, render) for build
├── scripts/install.sh            # go build + drop binary on PATH
└── docs/superpowers/specs/       # this spec
```

## Build Pipeline Detail (`cmd/build`)

1. Read `mobs.list` → list of mob ids (skip blank/`#` lines).
2. For each id:
   - GET `https://maplestory.io/api/{REGION}/{VERSION}/mob/{id}` → name, level, isBoss.
   - GET `.../mob/{id}/render/stand` → GIF bytes; save to a temp file.
   - Run `chafa --format symbols --animate off -c full --dither none -w 9 --size {W}x{H} <tmp>`
     → capture stdout (truecolor ANSI).
   - Strip the `\e[?25l`/`\e[?25h` cursor sequences chafa emits.
   - Write to `colorscripts/large/<slug>` where `slug = kebab(name)`, with `-<id>`
     suffix appended on slug collision.
3. Write `maple.json` with the embedded set's metadata.
4. `REGION` (e.g. `GMS`) and `VERSION` (e.g. `255`) and `SIZE` are top-level
   constants/flags so they can be re-pinned later.

Notes:
- `chafa` is a **build-time** dependency only; end users never need it.
- No U+2800 trick here — that is a Claude Code statusline-specific anti-trim hack;
  ordinary terminals render leading spaces fine, so sprites use clean ANSI.

## Runtime CLI Detail (`cmd/maple`)

| Invocation | Behavior |
|------------|----------|
| `maple` or `maple -r` | Print a random embedded mob. |
| `maple -n, --name <name>` | Print a specific mob by (case-insensitive) name. |
| `maple -i, --id <id>` | Print a specific mob by id. |
| `maple -l, --list` | List all embedded mobs (name + id + level). |
| `maple --no-title` | Suppress the name line (for statusline/banner use). |

- Default output: the ANSI sprite followed by a title line (`Name  Lv.N`), unless
  `--no-title`.
- Unknown name/id → friendly error to stderr, exit non-zero.
- Reads nothing from disk or network; everything is embedded.

## Starter Curation (`mobs.list`)

~30 classic early-game monsters by id, e.g. Snail (100100), Blue/Red Snail,
Orange/Green/Horny Mushroom, Slime, Pig, Ribbon Pig, Stump, Octopus, Jr. Necki,
Mano, Stirge, etc. Exact ids are confirmed against the API during build; any id
that fails to fetch is logged and skipped (build does not abort).

## Error Handling

- **Build:** per-mob failures (404, timeout, non-image response) are logged and
  skipped; the build continues and reports a summary (N succeeded / M skipped).
  Network/API base errors abort early with a clear message.
- **Runtime:** missing name/id → stderr message + non-zero exit. Embedded data is
  validated at startup (parse `maple.json`); a corrupt embed is a fatal build-time
  problem, not a runtime guess.

## Testing

- **`internal/sprite`** (unit): selection logic — random pick is non-empty, name
  lookup is case-insensitive, unknown name returns an error, `--no-title` omits
  the title. Uses a small fixture set, not the real embed.
- **`internal/maplestoryio`** (unit): URL construction; response parsing against
  a recorded sample payload (no live network in tests).
- **`cmd/build`** (integration, optional/manual): a `-dry-run` that fetches 1–2
  ids and prints sizes, run manually since it needs network + chafa.
- Coverage target: 80% on `internal/` packages (the logic that matters); the
  thin `main.go` wrappers are exercised via a couple of CLI smoke tests.

## Open Questions / Risks

- maplestory.io version pinning: `255` works today; if a version is retired the
  build must be re-pinned. Mitigated by making `VERSION` a single constant.
- Sprite cell size (`--size`): start at a tuned default (e.g. `40x40`, similar to
  the statusline work); revisit per how mobs look. Configurable in build.
- Name collisions across ids handled by `-<id>` slug suffix.
