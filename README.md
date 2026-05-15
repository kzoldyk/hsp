# HSP — HTTP Superpowers

Postman-like API testing in your terminal. Interactive, scriptable, 10MB binary.

## Installation

```bash
go install github.com/hitesh103/hsp@latest
```

Or build from source:

```bash
git clone https://github.com/hitesh103/hsp.git && cd hsp
go build -o hsp && sudo mv hsp /usr/local/bin/
```

## Problem

cURL is powerful but you have to memorize flags. Postman is visual but heavy (500MB+). You just want to test an API without context-switching or flag reference cards.

## Solution

HSP gives you the best of both: an interactive prompt (no flags to remember) and a lightweight CLI (single ~10MB binary). One command, answer the questions, get beautiful output.

## Commands

| Command | Description |
|---------|-------------|
| `hsp request` / `r` | Interactive request builder — step-by-step prompts |
| `hsp get` / `g` `<url>` | Quick GET |
| `hsp post` / `p` `<url>` | Quick POST |
| `hsp put` / `pu` `<url>` | Quick PUT |
| `hsp patch` / `pa` `<url>` | Quick PATCH |
| `hsp delete` / `d` `<url>` | Quick DELETE |
| `hsp env` | Manage environments (dev/staging/prod) |
| `hsp var` | Manage variables (set/list/delete/export) |
| `hsp profile` | Save/run/list/delete/edit request profiles |
| `hsp test` | Run JSON test suites with assertions |

## Features

**Interactive mode** — Run `hsp request`, answer prompts for URL, method, headers, params, body. Preview before sending. Auto-saves to history.

**Quick commands** — `hsp get <url>`, `hsp post <url> --json '{"key":"val"}'`. Short aliases: `g`, `p`, `pu`, `pa`, `d`.

**Variables & environments** — `hsp var set BASE_URL https://api.example.com`, then use `{{BASE_URL}}` anywhere in requests. Switch between dev/staging/prod with `hsp env prod`.

**Session memory** — `hsp request --last` re-sends your last request. `--resume` loads it for modification.

**Profiles** — Save named request templates with `hsp profile save <name>`. Run with `hsp profile run <name>`.

**Test suites** — Define tests as JSON with assertions (status codes, body content, headers, response time). Run with `hsp test run suite.json`.

```json
{
  "tests": [{
    "name": "Create user",
    "request": { "method": "POST", "url": "{{BASE_URL}}/users", "body": {"name":"test"} },
    "assertions": [
      { "type": "status", "expected": "201" },
      { "type": "body_contains", "path": "$.name", "value": "test" }
    ]
  }]
}
```

**Color-coded TUI** — Status codes colored by range (2xx green, 4xx red, 5xx magenta), JSON pretty-printing with syntax highlighting, boxed layouts.

**Auto history** — Every request saved to `~/.hsp/history/` automatically.

## Configuration

Config file at `~/.hsp/config.yaml`:

```yaml
environments:
  default:
    BASE_URL: ""
    API_KEY: ""
  prod:
    BASE_URL: "https://api.example.com"
    API_KEY: "prod-token"
activeEnv: prod
```

File locations:

| Path | Purpose |
|------|---------|
| `~/.hsp/config.yaml` | Variables & environments |
| `~/.hsp/history/` | Request history |
| `~/.hsp/profiles/` | Saved profiles |
| `~/.hsp/suites/` | Test suites |
| `~/.hsp/.last_request.json` | Session memory |

## License

MIT
