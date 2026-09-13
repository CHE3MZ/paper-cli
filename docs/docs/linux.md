# Linux

## What you need

- **Java 21 or newer.** Check with `java -version`. Install it with your
  package manager, for example:

  ```
  sudo apt install openjdk-21-jre-headless    # Debian/Ubuntu
  sudo dnf install java-21-openjdk-headless   # Fedora
  ```

- **The `paper` binary.** Either download a prebuilt `paper` binary, or
  build it yourself (needs the [Go toolchain](https://go.dev/dl/)):

  ```
  sh scripts/build.sh
  ```

  This drops `build/paper` (~177MB — the server files live inside it).
  Move it somewhere on your PATH, e.g. `~/.local/bin` or `/usr/local/bin`,
  or call it by its full path.

## Your first server

```
paper new ./my-server
paper run ./my-server --nogui
```

The first launch takes a bit while the world generates; when you see
`Done`, the server is up. Type `stop` in the console to shut it down
cleanly.

To run a server from its own folder, skip the path entirely:

```
cd my-server
paper run --nogui -m=2gb
```

A common way to keep a server running after you log out is a terminal
multiplexer (`tmux` / `screen`) or a systemd service — `paper run` itself
is a normal foreground process, so wrap it however you like.

## Notes

- `--nogui` matters on headless machines: without it the server tries to
  open a GUI and crashes.
- The full command list lives on the [Home](index.md) page.
