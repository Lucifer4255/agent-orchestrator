#!/usr/bin/env bash
# Build a self-contained tmux binary for bundling in the AO desktop app.
#
# Pinned upstream versions (update checksums when bumping):
#   tmux     3.4   https://github.com/tmux/tmux/releases/download/3.4/tmux-3.4.tar.gz
#   libevent 2.1.12-stable
#   ncurses  6.4
#
# Output: frontend/tmux/tmux (+ NOTICE.md with license text)
#
# macOS: links libevent and ncurses statically; system libc remains dynamic.
# Linux: links libevent and ncurses statically against glibc (same pattern as CI).
#
# Prerequisites: curl, tar, make, a C compiler (clang/gcc), perl.
set -euo pipefail

# Host terminfo paths (e.g. Ghostty) must not leak into the ncurses build.
unset TERMINFO TERMINFO_DIRS

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT_DIR="${1:-$ROOT/frontend/tmux}"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

TMUX_VERSION=3.4
LIBEVENT_VERSION=2.1.12-stable
NCURSES_VERSION=6.4

TMUX_URL="https://github.com/tmux/tmux/releases/download/${TMUX_VERSION}/tmux-${TMUX_VERSION}.tar.gz"
TMUX_SHA256="551ab8dea0bf505c0ad6b7bb35ef567cdde0ccb84357df142c254f35a23e19aa"
LIBEVENT_URL="https://github.com/libevent/libevent/releases/download/release-${LIBEVENT_VERSION}/libevent-${LIBEVENT_VERSION}.tar.gz"
LIBEVENT_SHA256="92e6de1be9ec176428fd2367677e61ceffc2ee1cb119035037a27d346b0403bb"
NCURSES_URL="https://invisible-mirror.net/archives/ncurses/ncurses-${NCURSES_VERSION}.tar.gz"
NCURSES_SHA256="6931283d9ac87c5073f30b6290c4c75f21632bb4fc3603ac8100812bed248159"

verify() {
  local file="$1" expected="$2"
  local actual
  actual="$(openssl dgst -sha256 "$file" | awk '{print $2}')"
  if [[ "$actual" != "$expected" ]]; then
    echo "checksum mismatch for $(basename "$file"): got $actual want $expected" >&2
    exit 1
  fi
}

fetch() {
  local url="$1" dest="$2" expected="$3"
  curl -fsSL "$url" -o "$dest"
  verify "$dest" "$expected"
}

DEPS="$WORK/deps"
mkdir -p "$DEPS" "$OUT_DIR"

fetch "$NCURSES_URL" "$WORK/ncurses.tar.gz" "$NCURSES_SHA256"
tar -xzf "$WORK/ncurses.tar.gz" -C "$WORK"
(
  cd "$WORK/ncurses-${NCURSES_VERSION}"
  ./configure \
    --prefix="$DEPS" \
    --with-shared=no \
    --with-normal \
    --without-debug \
    --without-termlib \
    --enable-overwrite \
    --without-man \
    --without-tests \
    --without-progs \
    --disable-db-install
  make -j"$(sysctl -n hw.ncpu 2>/dev/null || nproc)"
  make install
)

fetch "$LIBEVENT_URL" "$WORK/libevent.tar.gz" "$LIBEVENT_SHA256"
tar -xzf "$WORK/libevent.tar.gz" -C "$WORK"
(
  cd "$WORK/libevent-${LIBEVENT_VERSION}"
  ./configure --prefix="$DEPS" --disable-shared --enable-static --disable-openssl
  make -j"$(sysctl -n hw.ncpu 2>/dev/null || nproc)"
  make install
)

fetch "$TMUX_URL" "$WORK/tmux.tar.gz" "$TMUX_SHA256"
tar -xzf "$WORK/tmux.tar.gz" -C "$WORK"
(
  cd "$WORK/tmux-${TMUX_VERSION}"
  tmux_ldflags="-L$DEPS/lib"
  tmux_libs="-levent_core -lncurses -lresolv"
  if [[ "$(uname -s)" == "Linux" ]]; then
    tmux_ldflags="-static -L$DEPS/lib"
    tmux_libs="-levent_core -lncurses -lresolv -ldl -lpthread"
  fi
  PKG_CONFIG_PATH="$DEPS/lib/pkgconfig${PKG_CONFIG_PATH:+:$PKG_CONFIG_PATH}" \
    ./configure \
      --prefix="$DEPS" \
      --with-libevent="$DEPS" \
      --disable-utf8proc \
      CPPFLAGS="-I$DEPS/include" \
      LDFLAGS="$tmux_ldflags"
  make -j"$(sysctl -n hw.ncpu 2>/dev/null || nproc)" LIBS="$tmux_libs"
  make install LIBS="$tmux_libs"
)

install -m 755 "$DEPS/bin/tmux" "$OUT_DIR/tmux"
"$OUT_DIR/tmux" -V >/dev/null

cat >"$OUT_DIR/NOTICE.md" <<EOF
Bundled tmux ${TMUX_VERSION}

Built from source by scripts/build-tmux.sh for the Agent Orchestrator desktop app.
Upstream components:

- tmux ${TMUX_VERSION} (ISC License)
  https://github.com/tmux/tmux
- libevent ${LIBEVENT_VERSION} (3-clause BSD License)
  https://github.com/libevent/libevent
- ncurses ${NCURSES_VERSION} (MIT-style License)
  https://invisible-mirror.net/archives/ncurses/

Redistribution follows each upstream license. tmux is Copyright (c) Nicholas Marriott
and contributors.
EOF

echo "built $OUT_DIR/tmux ($(file -b "$OUT_DIR/tmux" 2>/dev/null || true))"
