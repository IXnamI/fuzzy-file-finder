# fuzzy-file-finder

A tiny, fast **terminal fuzzy finder** for Windows focused on finding files on your filesystem. Built in Go, with a responsive TUI.

## Features
- Interactive fuzzy search over files/directories
- Real-time filtering as you type (snappy TUI via `tcell`)
- Open selected path in File Explorer

> Tech stack: **Go 1.22+**, `tcell`, `clipboard`, `x/sys` (see `go.mod`).


## Install

### From source
```bash
# Go 1.22+ required
git clone https://github.com/IXnamI/fuzzy-file-finder
cd fuzzy-file-finder
go build .
```
> License: **MIT**.

## Usage

```bash
# Search over the entire filesystem
./fuzzy-file-finder
```

### Basic flow
1. Launch application
2. Start typing to filter.
3. Use arrows/PageUp/PageDown to move.
4. Press **Enter** to open the highlighted item.

## Keybindings (suggested defaults)
- **Type** — fuzzy filter
- **↑/↓** — navigate results
- **Enter** — open file/folder
- **Esc** — quit

## Development

### Prerequesites
- Go 1.22+  
- Windows

### Run
```bash
go run .
```

## WIP
- [ ] Ignore patterns from `.gitignore`
- [ ] Preview pane (file head, syntax highlight)
- [ ] A way to change user configurations
- [ ] Cross-platform

## Alternatives / Inspiration
- [fzf](https://github.com/junegunn/fzf) — the classic general-purpose fuzzy finder (great reference)

## License
**MIT** — see [LICENSE](./LICENSE).