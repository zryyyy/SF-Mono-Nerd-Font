# Nerd Font Patcher

A small personal build script for patching selected fonts with the official Nerd Fonts Docker image.

The repository intentionally does not store font binaries. Fonts are downloaded from `font.json`, patched with `nerdfonts/patcher`, and packaged into release zip files.

## Requirements

- Go
- Docker
- 7-Zip (`7z` or `7zz`) for DMG and ZIP source

## Usage

Review `font.json`, then run:

```sh
go run . -config font.json
```

To validate the configuration and check required commands without downloading or patching:

```sh
go run . -config font.json -dry-run
```

Outputs are written to `dist/`:

- `<font-group>.zip` for each enabled group
- `all_fonts.zip` containing all patched fonts

## Configuration

`font.json` groups fonts by output archive. Each group has a safe `name`, one or more remote `sources`, and optional `include` patterns.

Supported source types:

- `font`: direct `.ttf` or `.otf` URL
- `zip`: `.zip` URL containing `.ttf` or `.otf` files
- `dmg`: DMG source extracted through external 7-Zip, used for SF Mono

The default patcher settings run:

```sh
docker run --rm -v <input>:/in -v <output>:/out -e PN=10 nerdfonts/patcher --complete
```

## Project layout

- `main.go`: thin CLI entry point
- `internal/app/`: build orchestration
- `internal/config/`: `font.json` loading and validation
- `internal/source/`: font downloads and source extraction
- `internal/docker/`: Nerd Fonts Docker invocation
- `internal/ziputil/`: output archive helpers
- `.github/workflows/patch.yml`: manual release workflow
