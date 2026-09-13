# Windows

## What you need

- **Java 21 or newer.** Paper 1.21 needs at least Java 21 to run. Check
  with `java -version`. If you don't have it:

  ```
  winget install Microsoft.OpenJDK.21
  ```

  or, with Scoop:

  ```
  scoop install openjdk21
  ```

- **The `paper` binary.** Either download a prebuilt `paper.exe`, or build
  it yourself (needs the [Go toolchain](https://go.dev/dl/)):

  ```
  powershell -ExecutionPolicy Bypass -File scripts/build.ps1
  ```

  This drops `build/paper.exe` (~177MB — the server files live inside it).
  Put it somewhere on your PATH, or call it by its full path.

## Your first server

```
paper new .\my-server
paper run .\my-server --nogui
```

That's it. The first launch takes a bit while the world generates; when you
see `Done`, the server is up. Type `stop` in the console to shut it down
cleanly.

To run a server from its own folder, skip the path entirely:

```
cd my-server
paper run --nogui -m=2gb
```

## Notes

- `--nogui` is worth using on Windows so the server doesn't open a
  separate GUI window.
- If PowerShell blocks the build script, that's the execution policy —
  the `-ExecutionPolicy Bypass` in the command above handles it for that
  one run, nothing on your system changes.
- The full command list lives on the [Home](index.md) page.
