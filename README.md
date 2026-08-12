# statusline

A status line for Claude Code: the active model, a context-usage bar with a
smooth thermometer gradient, green when the context is nearly empty and warming
to red as it fills, followed by the session cost and remaining usage limits.

```text
Opus · 500k ██████████████░░░░░░░░░░░░░░ · $20.57 · 5h 76% · 7d 59%
```

## What it shows

- Left: the current model from `model.display_name` (falling back to `model.id`
  when the display name is absent), followed by input tokens used in the current
  context window, read from Claude Code's `context_window.total_input_tokens`
  field. Display
  names shorten as the terminal narrows, for example `Opus (1M Context)` becomes
  `Opus (1M)` and then `Opus` before the model is omitted. Raw `model.id` values
  are shown as-is or omitted.
- Bar: fills against the context limit, colored green to red by how full it is,
  ending on a fractional sub-cell so it grows smoothly.
- Right: the session cost from `cost.total_cost_usd`, followed by the remaining
  five-hour and seven-day usage limits when Claude Code provides them.

## Build

This repo consumes [go-makefile](https://github.com/agoodkind/go-makefile) for
its build and lint pipeline.

```makefile
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
