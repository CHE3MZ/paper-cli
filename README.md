<img src="docs/docs/images/logo.png" alt="Paper logo" width="320" />

# Paper CLI

A tiny offline-first PaperMC server deployer. One `paper` binary, three
commands, no downloads at runtime — the server files ship inside it.

```sh
paper new ./my-server     # create a server
paper run ./my-server     # run it
paper delete ./my-server  # wipe everything but paper.jar
```

## Build

```sh
# Windows (PowerShell)
powershell -ExecutionPolicy Bypass -File scripts/build.ps1
# Linux/macOS
sh scripts/build.sh
```

Needs the [Go toolchain](https://go.dev/dl/) plus Java 21+ to actually run
servers. Full guides per platform, the command reference, and developer
docs are in **`docs/`** (mkdocs project, published as the project site).

# License

**See [the paper MC project license here](https://github.com/PaperMC/Paper/blob/main/LICENSE.md)**

Paper CLI itself is licensed under the [MIT License](LICENSE).
