# text-swap

`text-swap` is a Go CLI tool for searching and replacing text in files.
It supports both one-off operations and rule-based batch processing via YAML/JSON config files.

Built with Cobra, it is designed for local automation, scripting, and release distribution with GoReleaser.

## Features

- Search occurrences of a target string in a file
- Replace text using a single rule (`--target` + `--replacement`) or multiple rules from config (`--config`)
- Optional case-insensitive matching (`--ignore-case`)
- Parallel chunk processing for large files (`--chunk-size`, or auto mode)
- Progress display for file processing
- Output to stdout or write to a separate output file (`--out`)

## Installation

### Option 1: Download prebuilt binaries (recommended)

For official releases, download binaries from GitHub Releases.

- Linux: `text-swap_*_Linux_*.tar.gz`
- macOS: `text-swap_*_Darwin_*.tar.gz`
- Windows: `text-swap_*_Windows_*.zip`

After extracting, place the executable in your `PATH`.

### Option 2: Build from source

Requirements:

- Go `1.26.5` or later

Build:

```bash
go build -o text-swap .
```

Or use the task:

```bash
task build
```

## Usage

General help:

```bash
text-swap --help
text-swap search --help
text-swap replace --help
```

### `search` command

Search for occurrences in a file.

```bash
text-swap search --file <path> --target <word> [--ignore-case] [--chunk-size <bytes>]
```

Or load rules from config:

```bash
text-swap search --file <path> --config <rules.yaml|rules.json> [--chunk-size <bytes>]
```

Flags:

- `-f, --file` (required): input file path
- `-t, --target`: target string to search
- `-c, --config`: YAML/JSON config containing rules
- `-i, --ignore-case`: case-insensitive search (single target mode only)
- `--chunk-size`: chunk size in bytes for line-boundary parallel search (`0` = auto by file size)

Constraints:

- Exactly one of `--target` or `--config` is required
- `--ignore-case` cannot be used with `--config`

Examples:

```bash
# Case-sensitive search
text-swap search -f ./input.txt -t "hello"

# Case-insensitive search
text-swap search -f ./input.txt -t "hello" -i

# Rule-based search from config
text-swap search -f ./input.txt -c ./configs/sample.yaml

# Force chunked processing with 64 KiB chunks
text-swap search -f ./input.txt -t "error" --chunk-size 65536
```

### `replace` command

Replace text in a file stream.

```bash
text-swap replace --file <path> --target <word> --replacement <new> [--ignore-case] [--out <path>] [--chunk-size <bytes>]
```

Or load multiple rules from config:

```bash
text-swap replace --file <path> --config <rules.yaml|rules.json> [--out <path>] [--chunk-size <bytes>]
```

Flags:

- `-f, --file` (required): input file path
- `-t, --target`: target string to replace
- `-r, --replacement`: replacement string (single target mode)
- `-c, --config`: YAML/JSON config containing replacement rules
- `-i, --ignore-case`: case-insensitive replacement (single target mode only)
- `-o, --out`: output file path (default: stdout)
- `--chunk-size`: chunk size in bytes for line-boundary parallel replace (`0` = auto by file size)

Constraints:

- Exactly one of `--target` or `--config` is required
- `--replacement` and `--ignore-case` cannot be used with `--config`
- `--out` must be different from `--file`

Examples:

```bash
# Replace to stdout
text-swap replace -f ./input.txt -t "foo" -r "bar"

# Save replaced output to a file
text-swap replace -f ./input.txt -t "foo" -r "bar" -o ./output.txt

# Case-insensitive replacement
text-swap replace -f ./input.txt -t "hello" -r "hi" -i

# Rule-based replacement from config
text-swap replace -f ./input.txt -c ./configs/sample.json -o ./output.txt

# Chunked replacement
text-swap replace -f ./large.txt -t "ERROR" -r "WARN" --chunk-size 65536 -o ./result.txt
```

## Config File Format

Both YAML and JSON are supported.

YAML example:

```yaml
rules:
	- target: "foo"
		replacement: "bar"
		ignore_case: true
	- target: "hello"
		replacement: "world"
```

JSON example:

```json
[
	{
		"target": "foo",
		"replacement": "bar",
		"ignore_case": true
	},
	{
		"target": "hello",
		"replacement": "world"
	}
]
```

## Development

Install task runner first if needed:

- https://taskfile.dev

Project setup:

```bash
task setup
```

Common tasks:

| Task | Description |
| --- | --- |
| `task setup` | Install development tools (`gopls`, `cobra-cli`, `goreleaser`, `golangci-lint`, `gofumpt`, `govulncheck`) |
| `task fmt` | Format code with `gofumpt` |
| `task lint` | Run `golangci-lint` |
| `task vulncheck` | Run `govulncheck ./...` |
| `task test` | Run all tests |
| `task build` | Build binary (`bin/app.exe`) |
| `task check` | Run `fmt`, `lint`, `vulncheck`, and `test` sequentially |

Additional test tasks:

- `task test:cover`: Generate coverage summary
- `task test:html`: Open HTML coverage report

## License

This project is licensed under the terms in [LICENSE](LICENSE).
