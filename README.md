<img src="docs/docs/images/logo.png" alt="Paper logo" width="320" />

# Paper CLI

A tiny offline-first PaperMC server deployer. One `paper` binary, three
commands, no downloads at runtime — the server files ship inside it.

```sh
paper new ./my-server     # create a server
paper run ./my-server     # run it
paper delete ./my-server  # wipe everything but paper.jar
```

## Install

```bat
install\windows.bat
# or
irm https://raw.githubusercontent.com/CHE3MZ/paper-cli/master/install/windows.bat | cmd
```
```sh
sh install/macos.sh
# or
curl -fsSL https://raw.githubusercontent.com/USER/paper-cli/master/install/macos.sh | bash

sh install/linux.sh
or
curl -fsSL https://raw.githubusercontent.com/USER/paper-cli/master/install/linux.sh | bash
```

### Uninstall

```bat
install\uninstall-windows.bat
```
```sh
sh install/uninstall-macos.sh
sh install/uninstall-linux.sh
```

## Build

```sh
# Windows (PowerShell)
powershell -ExecutionPolicy Bypass -File scripts/build.ps1
# Linux/macOS
sh scripts/build.sh
```

Needs the [Go toolchain](https://go.dev/dl/) plus Java 21+ to actually run
servers. The build scripts run the embed generator first; plain
`go build`/`go test` on a fresh clone need `go run ./src/genembed` first.
Full guides per platform, the command reference, and developer
docs are in **`docs/`** (mkdocs project, published as the project site).

Want to contribute? See [CONTRIBUTING.md](CONTRIBUTING.md).

## Experimental

The main experimental branch is the "dev" branch, any features that are deemed 
unpolished or unready/not production-ready will be worked on inside of that 
branch specifically in order to keep the master branch clean.

# License

Paper CLI's own code is licensed under the [MIT License](LICENSE).

The binary bundles third-party software (PaperMC under GPL-3.0, Minecraft
server files under Mojang's EULA) — see [NOTICE.md](NOTICE.md), including
the GPL copy at [LICENSES/GPL-3.0.txt](LICENSES/GPL-3.0.txt). Running a
server means you accept the [Minecraft EULA](https://aka.ms/MinecraftEULA).

<img src="docs/docs/images/logo-small.png" alt="Paper logo small" width="96" />
