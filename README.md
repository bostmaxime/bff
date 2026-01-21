# bff

Track file changes and find duplicate files using SHA-256 content hashing.

## Build the executable

```bash
go build -o bff
```

## Commands

### Index files
```bash
./bff index [--hidden] [directory]
```
Creates or updates `bff.json` with file information. Use `--hidden` to include hidden files.

### Compare changes
```bash
./bff compare [directory]
```
Shows added, modified, moved/renamed, or deleted files since last indexing. Uses the saved `--hidden` setting.

### Find all duplicates
```bash
./bff duplicates [directory]
```
Shows all groups of files with identical content.

### Find duplicates of a specific file
```bash
./bff find <file-path> [directory]
```
Shows all files with the same content as the specified file.

## Notes

- `compare`, `duplicates`, and `find` commands require running `./bff index` first in the specified directory
- Specifying a directory is optional, it defaults to current directory if not specified.
