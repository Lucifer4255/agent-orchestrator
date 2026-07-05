# Bundled tmux

The desktop app ships a self-contained `tmux` binary for macOS and Linux so users
do not need a system install to run agent sessions.

## Build

From the repo root (or via `npm run build:tmux` in `frontend/`):

```bash
scripts/build-tmux.sh
```

Output lands in `frontend/tmux/`:

- `tmux` — the executable copied into the app bundle via `extraResource`
- `NOTICE.md` — upstream license attribution (tmux ISC, libevent BSD, ncurses MIT-style)

The script downloads pinned upstream tarballs with SHA-256 verification, then
builds static libevent and ncurses before linking tmux:

| Component | Version |
|-----------|---------|
| tmux      | 3.4     |
| libevent  | 2.1.12-stable |
| ncurses   | 6.4     |

macOS builds link libevent/ncurses statically (system `libSystem` + `libresolv`
remain dynamic). Linux CI builds use the same script on `ubuntu-latest` with
`-static` linkage against glibc.

Windows does not bundle tmux; the daemon uses ConPTY there instead.

## Runtime resolution

The daemon picks a tmux binary in this order (see `tmux.ResolveBinary`):

1. `AO_TMUX_BIN` — explicit operator override
2. `tmux` on `PATH` — preserves a user's own install
3. `AO_BUNDLED_TMUX_BIN` — set by the Electron app to the packaged binary

The desktop supervisor sets `AO_BUNDLED_TMUX_BIN` when spawning the daemon.
`ao doctor` reports which source was chosen.

## Packaging

`frontend/forge.config.ts` lists `tmux` in `extraResource` alongside `daemon`.
`npm run prepackage` / `premake` / `publish` run `build:tmux` before
`osxSign`, so the binary is sealed with the macOS app bundle.

Per-OS release runners (`frontend-release.yml`) build host-native tmux the same
way as the bundled `ao` daemon (arm64 + x64 macOS legs, linux x64 on Ubuntu).
