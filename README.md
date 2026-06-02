# statusline

A status line for Claude Code: a context-usage bar with a smooth thermometer
gradient, green when the context is nearly empty and warming to red as it fills,
followed by the session cost.

```
500k ██████████████░░░░░░░░░░░░░░ 872k · $20.57
```

## What it shows

- Left: tokens used in the current session, read from the most recent `usage`
  block in the transcript Claude Code passes on stdin.
- Bar: fills against the usable input ceiling, colored green to red by how full
  it is, ending on a fractional sub-cell so it grows smoothly.
- Right: the usable input ceiling and the session cost.

The usable input ceiling is the model's context window minus the tokens reserved
for the reply:

```
ceiling = context_window - CLAUDE_CODE_MAX_OUTPUT_TOKENS
```

The window is 1M when the model id carries `[1m]`, otherwise 200k. The reply
reserve is read from `CLAUDE_CODE_MAX_OUTPUT_TOKENS` at runtime, falling back to
32k when unset. When usage exceeds the ceiling, the right number rises to match
usage so the bar reads full and the left never exceeds the right.

## Build

This repo consumes [go-makefile](https://github.com/agoodkind/go-makefile) for
its build and lint pipeline.

```
make build
```

The binary is written to `dist/statusline`.

## Install

Point Claude Code's status line at the built binary in `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "/absolute/path/to/statusline/dist/statusline"
  }
}
```
