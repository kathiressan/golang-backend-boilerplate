# Configuration Reference

This document lists all environment variables used to configure the application. These are defined and parsed in `configs/config.go`.

## 🌐 Core Settings

| Variable      | Default       | Description                                      |
| :---          | :---          | :---                                             |
| `ENVIRONMENT` | `development` | `development`, `staging`, or `production`.       |
| `PORT`        | `8000`        | The port the server listens on.                  |
| `GIN_MODE`    | `debug`       | `debug`, `release`, or `test`.                   |
| `APP_NAME`    | (Required)    | Name of the application (used in logs/metadata). |

---

## 🗄 Database Settings

| Variable       | Default     | Description                                         |
| :---           | :---        | :---                                                |
| `DATABASE_URL` | (Optional)  | Full Postgres connection string (takes precedence). |
| `DB_HOST`      | `localhost` | Database host.                                      |
| `DB_PORT`      | `5432`      | Database port.                                      |
| `DB_USER`      | (Required)  | Database user.                                      |
| `DB_PASSWORD`  | (Required)  | Database password.                                  |
| `DB_NAME`      | (Required)  | Database name.                                      |
| `DB_SSLMODE`   | `disable`   | Postgres SSL mode (`disable`, `require`, etc).      |

---

## 🔐 Security & JWT

| Variable                      | Default | Description                                              |
| :---                          | :---    | :---                                                     |
| `JWT_SIGNING_METHOD`          | `HS256` | `HS256` or `RS256`.                                      |
| `JWT_SECRET`                  | (Required for HS256) | Shared secret for token signing. |
| `JWT_PRIVATE_KEY`             | (Required for RS256) | RSA private key in PEM format.   |
| `JWT_PUBLIC_KEY`              | (Required for RS256) | RSA public key in PEM format.    |
| `ACCESS_TOKEN_EXPIRY_MINUTES` | `60`    | Access token lifetime.                                   |
| `REFRESH_TOKEN_EXPIRY_DAYS`   | `30`    | Refresh token lifetime.                                  |
| `RATE_LIMIT_ENABLED`          | `false` | Enables/disables IP-based rate limiting.                 |
| `TRUSTED_PROXIES`             | (Empty) | Comma-separated list of trusted proxy IPs.               |

---

## ⏱ Timeouts

| Variable                  | Default | Description                               |
| :---                      | :---    | :---                                      |
| `READ_TIMEOUT_SECONDS`    | `10`    | Max duration for reading a request.       |
| `WRITE_TIMEOUT_SECONDS`   | `10`    | Max duration for writing a response.      |
| `IDLE_TIMEOUT_SECONDS`    | `120`   | Max duration to keep idle connections.    |
