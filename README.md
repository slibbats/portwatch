# portwatch

A lightweight CLI daemon that monitors port usage and alerts on unexpected listeners.

---

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

---

## Usage

Start the daemon with a list of allowed ports:

```bash
portwatch --allow 22,80,443 --interval 10s
```

Run a one-time scan and print current listeners:

```bash
portwatch scan
```

Example alert output when an unexpected listener is detected:

```
[ALERT] Unexpected listener detected: 0.0.0.0:8080 (PID 4821, process: python3)
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--allow` | `""` | Comma-separated list of allowed ports |
| `--interval` | `30s` | How often to poll port usage |
| `--log` | `stdout` | Log output destination |
| `--quiet` | `false` | Suppress non-alert output |

---

## Configuration

Optionally place a config file at `~/.portwatch.yaml`:

```yaml
allow:
  - 22
  - 80
  - 443
interval: 15s
log: /var/log/portwatch.log
```

---

## License

MIT © 2024 [yourusername](https://github.com/yourusername)