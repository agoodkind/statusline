#!/usr/bin/env bash
# Build the statusline binary from this worktree and atomically install it to a
# stable, worktree-independent location ($XDG_BIN_HOME or ~/.local/bin). Intended
# to run from the repo post-commit hook, but safe to run by hand at any time.
#
# Concurrency model (many Claude instances exec the installed binary constantly):
#   - stamp file makes a re-run for already-built source a no-op (idempotent)
#   - flock dedups concurrent builds; the lock auto-releases on process exit,
#     so a killed build never leaves a stale lock behind
#   - build to a unique temp on the same filesystem, then rename(2) into place,
#     so readers/execs never see a torn file and never hit ETXTBSY
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

DEST_DIR="${XDG_BIN_HOME:-$HOME/.local/bin}"
DEST="$DEST_DIR/statusline"
STAMP="$DEST_DIR/.statusline.stamp"
LOCK_FILE="$DEST_DIR/.statusline.lock"
LOCK_FD=9

TMP_BIN=""
FLOCK_BIN=""

cleanup() {
    if [[ -n "$TMP_BIN" && -e "$TMP_BIN" ]]; then
        rm -f "$TMP_BIN" || true
    fi
}
trap cleanup EXIT
trap 'exit 130' INT TERM

ensure_go() {
    if command -v go >/dev/null 2>&1; then
        return 0
    fi
    local candidate
    for candidate in /opt/homebrew/bin /usr/local/go/bin "$HOME/go/bin" /usr/local/bin; do
        if [[ -x "$candidate/go" ]]; then
            PATH="$candidate:$PATH"
            return 0
        fi
    done
    return 1
}

# Locate flock. Correctness does not depend on it (the atomic rename below is
# what prevents corruption); flock only dedups concurrent builds. If it is not
# found we proceed without the dedup lock rather than fail the commit.
resolve_flock() {
    if FLOCK_BIN="$(command -v flock 2>/dev/null)"; then
        return 0
    fi
    local candidate
    for candidate in /opt/homebrew/bin/flock /usr/bin/flock /usr/local/bin/flock; do
        if [[ -x "$candidate" ]]; then
            FLOCK_BIN="$candidate"
            return 0
        fi
    done
    FLOCK_BIN=""
    return 1
}

# Identity of the build inputs. Changes only when source that affects the binary
# changes, so doc-only commits skip the rebuild. Falls back to the commit id if
# any expected path is missing.
compute_source_id() {
    local ids
    if ids="$(git -C "$REPO_ROOT" rev-parse 'HEAD:cmd' 'HEAD:internal' 'HEAD:go.mod' 'HEAD:go.sum' 2>/dev/null)"; then
        printf '%s' "$ids" | tr '\n' '-'
    else
        git -C "$REPO_ROOT" rev-parse HEAD
    fi
}

is_current() {
    local want="$1"
    [[ -x "$DEST" && -f "$STAMP" && "$(cat "$STAMP" 2>/dev/null)" == "$want" ]]
}

build_and_install() {
    local source_id="$1"
    TMP_BIN="$(mktemp "$DEST_DIR/.statusline.XXXXXX")"
    if go build -C "$REPO_ROOT" -o "$TMP_BIN" ./cmd/statusline; then
        mv -f "$TMP_BIN" "$DEST"
        TMP_BIN=""
        printf '%s\n' "$source_id" >"$STAMP"
        echo "deploy-statusline: installed $DEST ($source_id)" >&2
    else
        echo "deploy-statusline: build failed; kept existing $DEST" >&2
        return 1
    fi
}

main() {
    if ! ensure_go; then
        echo "deploy-statusline: go toolchain not found on PATH; skipping rebuild" >&2
        exit 0
    fi

    local source_id
    source_id="$(compute_source_id)"

    if is_current "$source_id"; then
        exit 0
    fi

    mkdir -p "$DEST_DIR"

    if resolve_flock; then
        eval "exec ${LOCK_FD}>\"\$LOCK_FILE\""
        if ! "$FLOCK_BIN" -n "$LOCK_FD"; then
            # Another build holds the lock and will produce a current binary.
            exit 0
        fi
        # Re-check under the lock: a build that finished between our first check
        # and acquiring the lock may already have installed this exact source.
        if is_current "$source_id"; then
            exit 0
        fi
    fi

    build_and_install "$source_id"
}

main "$@"
