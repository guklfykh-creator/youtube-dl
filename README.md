<div align="center">

  <h1>
    YouTube DL Bot 
  </h1>

  <p><strong>Telegram Bot for downloading YouTube videos & audio via GitHub Actions + MTProto</strong></p>

  <p>
    <img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go" alt="Go 1.22" />
    <img src="https://img.shields.io/badge/gotd%2Ftd-v0.99-00ADD8?style=for-the-badge" alt="gotd/td" />
    <img src="https://img.shields.io/badge/yt--dlp-latest-orange?style=for-the-badge" alt="yt-dlp" />
    <img src="https://img.shields.io/badge/MTProto-2.0-26A5E4?style=for-the-badge" alt="MTProto" />
  </p>

  <br/>

  <em>Send a link → Pick quality → Get the file in Telegram</em>

</div>

---

## Features

- **Video downloads** — Best, 1080p, 720p, 480p, 360p
- **Audio extraction** — MP3 & M4A formats
- **MTProto upload** — sends files up to 2 GB via gotd/td
- **Bot API fallback** — files ≤ 50 MB sent via Telegram Bot API
- **Secure** — secrets stay in `.env` and GitHub Secrets; no credentials in code
- **Session timeout** — 5 min auto-expiry for pending quality selections
- **Persian UI** — all bot messages in Farsi
- **Interactive session setup** — `--setup-session` command to authenticate and generate `TG_SESSION`

---

## How It Works

```
┌──────────┐      ┌──────────────┐      ┌──────────────────┐      ┌───────────┐
│  User    │─────▶│  Telegram Bot │─────▶│  GitHub Actions   │─────▶│  yt-dlp   │
│ (Telegram)│◀────│  (Go server)  │◀────│  (download.yml)   │◀────│  + ffmpeg │
└──────────┘      └──────────────┘      └──────────────────┘      └───────────┘
     │                   │                       │                       │
     │  1. Send URL      │                       │                       │
     │──────────────────▶│                       │                       │
     │                   │  2. Show quality menu  │                       │
     │◀──────────────────│                       │                       │
     │  3. Pick quality  │                       │                       │
     │──────────────────▶│                       │                       │
     │                   │  4. Dispatch workflow  │                       │
     │                   │──────────────────────▶│                       │
     │                   │                       │  5. Download & encode │
     │                   │                       │──────────────────────▶│
     │                   │                       │                       │
     │                   │                       │  6. Upload via MTProto│
     │                   │                       │◀──────────────────────│
     │  7. Receive file  │                       │                       │
     │◀──────────────────│                       │                       │
```

1. **User** sends a YouTube URL to the Telegram bot
2. **Bot** shows an inline keyboard with quality/format choices
3. **User** picks a quality (e.g. 720p video or MP3 audio)
4. **Bot** triggers a `workflow_dispatch` on GitHub Actions via the GitHub API
5. **GitHub Actions** runs `yt-dlp` + `ffmpeg` to download and encode the media
6. **Uploader mode** in the same binary sends the file to the user via MTProto (gotd/td), or falls back to Bot API for ≤ 50 MB
7. **User** receives the file directly in Telegram

---

## Quick Start

### 1. Clone

```sh
git clone https://github.com/Mezdia2/youtube-dl.git
cd youtube-dl
```

### 2. Configure environment

```sh
cp .env.example .env
```

Edit `.env` and fill in your values (see [Configuration](#configuration)).

### 3. Set up MTProto session

Before running the bot, you need to generate a `TG_SESSION` for the MTProto uploader. The program provides an interactive setup command:

```sh
go run . --setup-session
```

This will:
1. Ask for your phone number (or use `TG_PHONE` from `.env` if set)
2. Request a verification code from Telegram and ask you to enter it
3. If your account has 2FA (two-factor authentication), ask for your password; otherwise it skips this step
4. Authenticate and save the session
5. Output the `TG_SESSION` base64 value
6. Offer to automatically write `TG_SESSION` into your `.env` file

After the setup completes, the `TG_SESSION` value will be in your `.env`. You also need to add it as a **GitHub repository secret** named `TG_SESSION`.

### 4. Set up GitHub Secrets

In your GitHub repository → **Settings → Secrets and variables → Actions**, add:

| Secret | Description |
|--------|-------------|
| `BOT_TOKEN` | Telegram bot token from @BotFather |
| `TG_APP_ID` | Telegram API ID from my.telegram.org |
| `TG_APP_HASH` | Telegram API Hash from my.telegram.org |
| `TG_SESSION` | Base64-encoded MTProto session (generated by `--setup-session`) |

### 5. Run the bot

```sh
go run .
```

---

## Configuration

All settings are loaded from environment variables (or `.env` file). The `.env.example` file contains all required variables:

```env
# Telegram bot token from @BotFather.
BOT_TOKEN=

# GitHub token used by the bot server to dispatch download.yml.
# Required permission: Actions write on the target repository.
GH_TOKEN=

# Repository where .github/workflows/download.yml lives.
GH_REPO_OWNER=
GH_REPO_NAME=
GH_WORKFLOW_FILE=download.yml
GH_DEFAULT_BRANCH=main

# MTProto uploader secrets. Put the same values in GitHub repository secrets.
TG_APP_ID=
TG_APP_HASH=
TG_SESSION=
```

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOT_TOKEN` | **Yes** | — | Telegram bot token from @BotFather |
| `GH_TOKEN` | **Yes** | — | GitHub PAT with `actions:write` on the target repo |
| `GH_REPO_OWNER` | **Yes** | — | GitHub repo owner (e.g. `Mezdia2`) |
| `GH_REPO_NAME` | **Yes** | — | GitHub repo name (e.g. `youtube-dl`) |
| `GH_WORKFLOW_FILE` | No | `download.yml` | Workflow filename in `.github/workflows/` |
| `GH_DEFAULT_BRANCH` | No | `main` | Branch to dispatch the workflow on |
| `TG_APP_ID` | **Yes** | — | Telegram API ID from my.telegram.org |
| `TG_APP_HASH` | **Yes** | — | Telegram API Hash from my.telegram.org |
| `TG_SESSION` | **Yes** | — | Base64-encoded MTProto session (generate with `--setup-session`) |
| `TG_PHONE` | No | — | Phone number for `--setup-session` (optional; asked interactively if not set) |

---

## MTProto Session Setup (`--setup-session`)

The MTProto uploader needs an authenticated Telegram session. Use the built-in `--setup-session` command to generate the session interactively:

```sh
go run . --setup-session
```

**Steps:**

1. **Phone number** — Enter your phone number in international format (e.g. `+989123456789`). You can also set `TG_PHONE` in `.env` to skip this prompt.
2. **Verification code** — Telegram sends a code to your account. Enter it when prompted.
3. **2FA password** — If your account has two-factor authentication enabled, you'll be asked for your password. If you don't have 2FA, this step is skipped automatically.
4. **Session output** — After successful authentication, the program outputs the `TG_SESSION` base64 string.
5. **Auto-update `.env`** — The program asks whether to automatically write `TG_SESSION` to your `.env` file. Choose `y` to auto-update, or `n` to set it manually.

**Required before running `--setup-session`:**
- `TG_APP_ID` and `TG_APP_HASH` must be set in `.env` (get them from https://my.telegram.org)

**Manual session encoding (alternative):**

If you prefer to generate the session manually:

```sh
# Authenticate and save session to file
go run . --auth --session session.json

# Base64-encode the session
# Linux/macOS:
base64 -w0 session.json

# Windows PowerShell:
[Convert]::ToBase64String([IO.File]::ReadAllBytes("session.json"))
```

Copy the base64 string and set it as `TG_SESSION` in `.env` and as a GitHub secret.

---

## Architecture

### Bot Server (`main.go`, `handlers.go`, `config.go`, `state.go`, `ghactions.go`)

| File | Responsibility |
|------|----------------|
| `main.go` | Bot API types, polling loop, CLI flags (`--setup-session`, `--file`, `--chat-id`) |
| `handlers.go` | `/start`, `/help`, `/cancel`, URL detection, quality selection |
| `config.go` | Environment loading, `.env` parser, validation (all env vars: BOT_TOKEN, GH_*, TG_*) |
| `state.go` | In-memory session store with 5-min TTL |
| `ghactions.go` | GitHub Actions `workflow_dispatch` API call |

### MTProto Uploader Mode (`uploader.go`)

| Feature | Detail |
|---------|--------|
| Session setup | `runSessionSetup()` — interactive phone+code auth, 2FA support, auto-write `.env` |
| Auth | `terminalAuth` struct implements gotd/td auth flow |
| Peer resolution | Username-based or dialog-based `InputPeer` lookup |
| Upload | `gotd/td` uploader with `FromPath` for large files |
| Send | `MessagesSendMedia` with proper MIME type & attributes |
| Fallback | Bot API `sendVideo`/`sendAudio` for ≤ 50 MB files (in download.yml) |

### GitHub Actions Workflow (`.github/workflows/download.yml`)

| Step | Action |
|------|--------|
| Checkout & Build | Checks repo, builds the main binary |
| Validate Secrets | Ensures `TG_APP_ID`, `TG_APP_HASH`, `TG_SESSION` exist |
| Install tools | Installs `yt-dlp` + `ffmpeg` on Ubuntu |
| Download | Runs `yt-dlp` with selected format/quality |
| Upload via MTProto | Sends file with 2 GB limit; falls back to Bot API ≤ 50 MB |
| Error notify | Sends error message to user on failure |
| Cleanup | Removes temp files & session |

---

## Security

- `.env` is excluded from git via `.gitignore`
- `GH_TOKEN` only used server-side to dispatch workflows; never sent to Actions
- `TG_SESSION` stored as GitHub Secret, written to disk only during upload
- Workflow inputs contain only `url`, `format_type`, `quality`, `chat_id`, `username` — no secrets
- Session files cleaned up in `always()` step
- `--setup-session` asks before writing to `.env` and uses file permission `0600`

---

## Building

```sh
go build -o youtube-dl-bot .
```

---

## Supported YouTube URLs

| Pattern | Example |
|---------|---------|
| Standard watch | `youtube.com/watch?v=...` |
| Short link | `youtu.be/...` |
| Shorts | `youtube.com/shorts/...` |
| Embed | `youtube.com/embed/...` |
| Legacy | `youtube.com/v/...` |
| Mobile | `m.youtube.com/watch?v=...` |
| Live | `youtube.com/live/...` |
| Music | `music.youtube.com/...` |

---

## File Size Limits

| Method | Max Size | Condition |
|--------|----------|-----------|
| Bot API | 50 MB | Automatic fallback when MTProto fails |
| MTProto (gotd/td) | ~2 GB | Primary upload method |
| Rejected | >2 GB | User notified to pick lower quality |

---

## License

This project is open source. See the [LICENSE](LICENSE) file for details.