# statusline

A status line for Claude Code: a context-usage bar with a smooth thermometer
gradient, green when the context is nearly empty and warming to red as it fills,
followed by the session cost.

```
500k ██████████████░░░░░░░░░░░░░░ 950k · $20.57
```

## What it shows

- Left: input tokens used in the current context window, read from Claude
  Code's `context_window.total_input_tokens` field with a transcript fallback
  when live token fields are absent.
- Bar: fills against the context limit, colored green to red by how full it is,
  ending on a fractional sub-cell so it grows smoothly.
- Right: the context limit and the session cost from `cost.total_cost_usd`.

The context-limit logic lives in `internal/tokenbudget`, which should be treated
as the source of truth for context-window defaults. When usage exceeds the limit,
the right number rises to match usage so the bar reads full and the left never
exceeds the right.

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
