# Notices

paper-cli's own source code is MIT licensed — see [LICENSE](LICENSE).

The `paper` binary also ships third-party software inside it. If you
distribute the binary, these notices travel with it.

## PaperMC Paper (GPL-3.0)

The binary embeds a PaperMC server template (`paper.jar`, `libraries/`,
`cache/`, configs) so `paper new` works offline. Paper is licensed under
the **GNU General Public License v3.0** (inherited via Spigot/Bukkit). A
copy of the license is included at [LICENSES/GPL-3.0.txt](LICENSES/GPL-3.0.txt),
and Paper's source code is at <https://github.com/PaperMC/Paper>.

paper-cli does not link against or modify Paper's code — it only unpacks
the template files and launches them, which the GPL treats as an
*aggregate*. That means paper-cli's own MIT license is unaffected, but
anyone distributing the `paper` binary must keep this notice, the GPL copy,
and the source link intact.

The exact build bundled in a given binary is recorded in
`current-build-version.json` (`current-build-path`) at build time.

## Minecraft / Mojang (proprietary, EULA)

The embedded template contains the vanilla Minecraft server code, which
belongs to Mojang AB / Microsoft and is **not** open source. Two things
follow from that:

1. **Running a server means agreeing to the Minecraft EULA**
   (<https://aka.ms/MinecraftEULA>). Note that `paper new` writes
   `eula.txt` with `eula=true` so the server starts immediately — by using
   a created server you are accepting the EULA on your own behalf. If
   you'd rather accept it consciously, flip it to `false`, read the EULA,
   then set it back to `true`.
2. **Redistributing Mojang's game files is restricted by the EULA**
   ("do not distribute anything we've made" — this explicitly covers
   server software). Paper itself avoids this by downloading the vanilla
   files from Mojang's servers at first launch instead of shipping them.
   paper-cli deliberately ships them for offline use, which is more
   convenient but sits in a grayer area. This is standard practice for
   small non-commercial tooling and enforcement against such tools is
   unheard of, but you should know the tradeoff: full offline support in
   exchange for redistributing files Mojang's EULA says not to pass
   around. If that worries you, the compliant alternative is excluding
   `cache/` from the bundle so the files download from Mojang on first
   boot (at the cost of requiring internet for that first run).

## Trademarks

`Minecraft` is a trademark of Mojang Synergies AB. paper-cli is not
affiliated with, endorsed by, or connected to Mojang AB, Microsoft, or
PaperMC. The project logo is original artwork and uses no Mojang assets.
