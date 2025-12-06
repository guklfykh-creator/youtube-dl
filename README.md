<div align="center">

  <h1>
    <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Television.png" alt="📺" width="40" height="40" />
    YouTube DL Bot
    <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Smileys/Cat%20Face.png" alt="🐱" width="40" height="40" />
  </h1>

  <p><strong>Telegram Bot for downloading YouTube videos & audio via GitHub Actions + MTProto</strong></p>

  <p>
    <img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go" alt="Go 1.22" />
    <img src="https://img.shields.io/badge/gotd%2Ftd-v0.99-00ADD8?style=for-the-badge" alt="gotd/td" />
    <img src="https://img.shields.io/badge/yt--dlp-latest-orange?style=for-the-badge" alt="yt-dlp" />
    <img src="https://img.shields.io/badge/MTProto-2.0-26A5E4?style=for-the-badge" alt="MTProto" />
  </p>

  <p>
    <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Hand%20gestures/Rightward%20Hand.png" alt="👉" width="20" height="20" />
    <a href="#how-it-works">How It Works</a>
    <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Hand%20gestures/Rightward%20Hand.png" alt="👉" width="20" height="20" />
    <a href="#quick-start">Quick Start</a>
    <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Hand%20gestures/Rightward%20Hand.png" alt="👉" width="20" height="20" />
    <a href="#configuration">Configuration</a>
    <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Hand%20gestures/Rightward%20Hand.png" alt="👉" width="20" height="20" />
    <a href="#architecture">Architecture</a>
  </p>

  <br/>

  <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Satellite%20Antenna.png" alt="📡" width="30" height="30" />
  <em>Send a link → Pick quality → Get the file in Telegram</em>

</div>

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Sparkles.png" alt="✨" width="22" height="22" /> Features

- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Video%20Camera.png" alt="🎥" width="18" height="18" /> **Video downloads** — Best, 1080p, 720p, 480p, 360p
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Musical%20Notes.png" alt="🎵" width="18" height="18" /> **Audio extraction** — MP3 & M4A formats
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Rocket.png" alt="🚀" width="18" height="18" /> **MTProto upload** — sends files up to 2 GB via gotd/td
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Fallback.png" alt="🔄" width="18" height="18" /> **Bot API fallback** — files ≤ 50 MB sent via Telegram Bot API
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Shield.png" alt="🛡" width="18" height="18" /> **Secure** — secrets stay in `.env` and GitHub Secrets; no credentials in code
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Clock.png" alt="⏱" width="18" height="18" /> **Session timeout** — 5 min auto-expiry for pending quality selections
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Globe%20Showing.png" alt="🌐" width="18" height="18" /> **Persian UI** — all bot messages in Farsi

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Gear.png" alt="⚙️" width="22" height="22" /> How It Works

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
6. **Uploader** (`cmd/uploader`) sends the file to the user via MTProto (gotd/td), or falls back to Bot API for ≤ 50 MB
7. **User** receives the file directly in Telegram

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Rocket.png" alt="🚀" width="22" height="22" /> Quick Start

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

### 3. Run the bot

```sh
go run .
```

### 4. Set up GitHub Secrets

In your GitHub repository → **Settings → Secrets and variables → Actions**, add:

| Secret | Description |
|--------|-------------|
| `BOT_TOKEN` | Telegram bot token from @BotFather |
| `TG_APP_ID` | Telegram API ID from my.telegram.org |
| `TG_APP_HASH` | Telegram API Hash from my.telegram.org |
| `TG_SESSION` | Base64-encoded MTProto session (see [Auth Setup](#mtproto-auth-setup)) |

### 5. Authenticate the MTProto uploader

```sh
# First-time: run auth to create a session
cd cmd/uploader
go run . --auth --session session.json

# Then base64-encode the session for GitHub Secrets
# On Linux/macOS:
base64 -w0 session.json

# On Windows (PowerShell):
[Convert]::ToBase64String([IO.File]::ReadAllBytes("session.json"))
```

Copy the base64 string into the `TG_SESSION` GitHub Secret.

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Clipboard.png" alt="📋" width="22" height="22" /> Configuration

All settings are loaded from environment variables (or `.env` file):

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BOT_TOKEN` | **Yes** | — | Telegram bot token from @BotFather |
| `GH_TOKEN` | **Yes** | — | GitHub PAT with `actions:write` on the target repo |
| `GH_REPO_OWNER` | **Yes** | — | GitHub repo owner (e.g. `Mezdia2`) |
| `GH_REPO_NAME` | **Yes** | — | GitHub repo name (e.g. `youtube-dl`) |
| `GH_WORKFLOW_FILE` | No | `download.yml` | Workflow filename in `.github/workflows/` |
| `GH_DEFAULT_BRANCH` | No | `main` | Branch to dispatch the workflow on |
| `TG_APP_ID` | **Yes*** | — | Telegram API ID (used by uploader) |
| `TG_APP_HASH` | **Yes*** | — | Telegram API Hash (used by uploader) |
| `TG_SESSION` | **Yes*** | — | Base64-encoded MTProto session |
| `TG_PHONE` | No | — | Phone number for interactive auth |

\* *Required in GitHub Secrets for the workflow; not needed on the bot server itself.*

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Floppy%20Disk.png" alt="💾" width="22" height="22" /> Architecture

### Bot Server (`main.go`, `handlers.go`, `config.go`, `state.go`, `ghactions.go`)

| File | Responsibility |
|------|----------------|
| `main.go` | Bot API types, polling loop, Telegram HTTP client |
| `handlers.go` | `/start`, `/help`, `/cancel`, URL detection, quality selection |
| `config.go` | Environment loading, `.env` parser, validation |
| `state.go` | In-memory session store with 5-min TTL |
| `ghactions.go` | GitHub Actions `workflow_dispatch` API call |

### MTProto Uploader (`cmd/uploader/main.go`)

| Feature | Detail |
|---------|--------|
| Auth | Interactive phone+code via `--auth` flag |
| Peer resolution | Username-based or dialog-based `InputPeer` lookup |
| Upload | `gotd/td` uploader with `FromPath` for large files |
| Send | `MessagesSendMedia` with proper MIME type & attributes |
| Fallback | Bot API `sendVideo`/`sendAudio` for ≤ 50 MB files |

### GitHub Actions Workflow (`.github/workflows/download.yml`)

| Step | Action |
|------|--------|
| Checkout & Build | Checks repo, builds uploader binary |
| Validate Secrets | Ensures `TG_APP_ID`, `TG_APP_HASH`, `TG_SESSION` exist |
| Install tools | Installs `yt-dlp` + `ffmpeg` on Ubuntu |
| Download | Runs `yt-dlp` with selected format/quality |
| Upload via MTProto | Sends file with 2 GB limit; falls back to Bot API ≤ 50 MB |
| Error notify | Sends error message to user on failure |
| Cleanup | Removes temp files & session |

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Lock.png" alt="🔐" width="22" height="22" /> Security

- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Shield.png" alt="🛡" width="16" height="16" /> `.env` is excluded from git via `.gitignore`
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Shield.png" alt="🛡" width="16" height="16" /> `GH_TOKEN` only used server-side to dispatch workflows; never sent to Actions
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Shield.png" alt="🛡" width="16" height="16" /> `TG_SESSION` stored as GitHub Secret, written to disk only during upload
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Shield.png" alt="🛡" width="16" height="16" /> Workflow inputs contain only `url`, `format_type`, `quality`, `chat_id`, `username` — no secrets
- <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Shield.png" alt="🛡" width="16" height="16" /> Session files cleaned up in `always()` step

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Gear.png" alt="⚙️" width="22" height="22" /> MTProto Auth Setup

The MTProto uploader needs an authenticated Telegram session. Create one locally:

```sh
# Option A: Phone authentication (recommended)
cd cmd/uploader
go run . --auth --session session.json

# Option B: Set TG_PHONE env var for non-interactive auth
export TG_PHONE="+989123456789"
go run . --auth --session session.json
```

After authentication, encode the session and add it to GitHub Secrets:

```sh
# Linux/macOS
base64 -w0 session.json | pbcopy   # or copy output manually

# Windows PowerShell
$b64 = [Convert]::ToBase64String([IO.File]::ReadAllBytes("session.json"))
Set-Content -Path "session_b64.txt" -Value $b64
```

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Hammer.png" alt="🔨" width="22" height="22" /> Building

```sh
# Build bot server
go build -o youtube-dl-bot .

# Build MTProto uploader
cd cmd/uploader
go build -o ../../uploader .
```

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Wrench.png" alt="🔧" width="22" height="22" /> Supported YouTube URLs

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

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Objects/Scale.png" alt="⚖️" width="22" height="22" /> File Size Limits

| Method | Max Size | Condition |
|--------|----------|-----------|
| Bot API | 50 MB | Automatic fallback when MTProto fails |
| MTProto (gotd/td) | ~2 GB | Primary upload method |
| Rejected | >2 GB | User notified to pick lower quality |

---

## <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Hand%20gestures/Handshake.png" alt="🤝" width="22" height="22" /> License

This project is open source. See the [LICENSE](LICENSE) file for details.

---

<div align="center">

  <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Smileys/Star-struck.png" alt="🤩" width="20" height="20" />
  Made with Go + gotd/td + yt-dlp
  <img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Smileys/Star-struck.png" alt="🤩" width="20" height="20" />

</div>