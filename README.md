# vaultwatch

A CLI tool that monitors HashiCorp Vault secret expiration and sends configurable alerts before rotation deadlines.

---

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or download a pre-built binary from the [releases page](https://github.com/yourusername/vaultwatch/releases).

---

## Usage

Set your Vault address and token, then run `vaultwatch` with a config file:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.your-token-here"

vaultwatch --config config.yaml
```

**Example `config.yaml`:**

```yaml
secrets:
  - path: secret/data/my-app/db-password
    warn_before: 72h
  - path: secret/data/my-app/api-key
    warn_before: 48h

alerts:
  slack:
    webhook_url: "https://hooks.slack.com/services/..."
  email:
    to: "ops-team@example.com"
```

Run a one-time check instead of continuous monitoring:

```bash
vaultwatch check --config config.yaml
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to config file | `./config.yaml` |
| `--interval` | Polling interval | `1h` |
| `--dry-run` | Check without sending alerts | `false` |

---

## License

MIT © [yourusername](https://github.com/yourusername)