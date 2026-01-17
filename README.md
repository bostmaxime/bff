# bff

Track file changes using SHA-256 hashes stored in a JSON file.

## Build the executable

```bash
go build -o bff
```

## Usage

In the directory where the executable is located:

### Index files
```bash
./bff index [directory]
```

Creates or updates `bff.json` with file data (hash, size, modification time). You can specify a directory. The current directory is used by default if no directory is specified. Note that the path is relative to where the command is run.

### Compare the current state with the last indexed state
```bash
./bff compare [directory]
```

Shows which files were added, modified, moved/renamed, or deleted since the last indexing. Here too you can specify a directory with the same logic as above.

### Include hidden files

Use `--hidden` or `-h` to include hidden files and directories:

```bash
./bff index --hidden
./bff compare --hidden
```

By default, hidden files are ignored. The compare command uses the same hidden setting as the one of the last indexing to reindex the directory for the comparison. You can override it with `--hidden` if it was not set before.
