# Docs → Telegram AI Support Bot

**Business outcome:** L1 support answers from your Markdown docs, with citations and a clean `/human` handoff — same pipeline in Telegram and a small web Trace UI.

## Stack

- Go 1.22+
- Telegram Bot API (long polling locally; webhook notes below)
- OpenAI-compatible LLM (DeepSeek by default; OpenAI / Ollama via env)
- In-memory keyword retrieval over `./docs/*.md` (swap for embeddings + pgvector later)

## Run

```bash
cd demos/telegram-rag-bot
cp .env.example .env
# set LLM_API_KEY (DeepSeek) and optionally TELEGRAM_TOKEN
export $(grep -v '^#' .env | xargs)   # or use direnv / your preferred loader

go mod tidy
go run ./cmd/bot
# Web UI: http://localhost:8083
```

Put 2–3 Markdown files in `docs/` (samples included). Ask in the UI or in Telegram.

### Provider presets

| Provider | `LLM_BASE_URL` | `LLM_MODEL` | Notes |
|----------|----------------|-------------|--------|
| DeepSeek (default) | `https://api.deepseek.com` | `deepseek-chat` | Cheap cloud API |
| OpenAI | `https://api.openai.com/v1` | `gpt-4o-mini` | Set key accordingly |
| Ollama (local) | `http://localhost:11434/v1` | `llama3.2` | `LLM_API_KEY=ollama` |

### Env

See [`.env.example`](./.env.example). Required: `LLM_API_KEY`. Optional: `TELEGRAM_TOKEN`, `ADMIN_CHAT_ID` (forward `/human`).

Never commit `.env`.

## Commands (Telegram)

| Input | Behavior |
|-------|----------|
| `/start` | Welcome |
| question | RAG answer + sources |
| `/human` | Escalate (+ optional admin forward) |

## Prod path

1. Telegram webhook instead of polling (TLS endpoint + `setWebhook`)
2. Embeddings + pgvector/Qdrant instead of keyword retrieve
3. PDF/HTML ingest pipeline
4. Slack/CRM ticket on `/human`
5. Optional: split ingest service later — keep `store` / `llm` / `pipeline` boundaries

## Loom script (~2 min)

1. Open `http://localhost:8083` → ask “How long for refunds?”
2. Show Pipeline trace: retrieve → prompt → llm → answer + sources
3. Same question in Telegram → sources footer
4. `/human` → escalation message
5. Close: “Docs-grounded support in Telegram — live in days, not months.”

## One-liner

> Go RAG support bot: Markdown docs → grounded answers with citations → Telegram + Trace UI, cheap OpenAI-compatible models.

## Bot avatar

See [AVATAR-PROMPT.md](./AVATAR-PROMPT.md) for a ChatGPT image prompt (BotFather `/setuserpic`).
