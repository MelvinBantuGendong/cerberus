# Cerberus

**A zero-trust security gateway for LLM traffic.** Cerberus sits between your
application and any OpenAI-compatible model endpoint and inspects every request
and response in both directions, blocking prompt injection, jailbreaks, and data
leakage before they reach the model or your users.

## What it solves

Shipping an LLM feature means exposing a model to untrusted input and returning
model output to untrusted users. That opens the door to prompt injection,
jailbreaks, and accidental leakage of secrets, PII, or your system prompt.
Cerberus closes it without changing your app:

- **One-line integration.** Point your OpenAI client's base URL at Cerberus and
  keep your existing code. It speaks the OpenAI Chat Completions schema and
  forwards to any backend you configure (OpenRouter, vLLM, Ollama, a custom
  model), so one integration covers every provider.
- **Inbound defense.** Requests are scanned for direct injection, jailbreaks,
  indirect injection hidden in tool output or documents, and obfuscated payloads.
  Malicious requests are blocked before they eveVr reach the model.
- **Outbound defense.** Model responses are scanned for leaked secrets, PII, and
  system-prompt echoes, and cut off before they reach the client. Choose full
  buffered scanning or near-live streaming.
- **Your provider key stays server-side.** Clients authenticate to Cerberus with
  their own keys, Cerberus swaps in the real upstream key before forwarding, so
  the provider credential never leaves your infrastructure.
- **A control plane, not just a proxy.** A web dashboard lets you configure the
  upstream, issue and revoke client keys, and toggle detectors live, with a Test
  Zone for firing attacks and watching the verdicts in real time.

Every request exits with a structured verdict, so you always know what was
allowed, flagged, or blocked and why.

## Run it locally

### Prerequisites

- Go 1.26+
- Node.js 20+ (for the dashboard)
- An [OpenRouter](https://openrouter.ai) API key (or any OpenAI-compatible
  upstream + key)

### 1. Configure

Cerberus ships with sensible defaults, so the only value you must provide to get
real completions back is your upstream API key. Copy the example and fill it in:

```bash
cp .env.example .env
```

Minimal `.env`:

```bash
# Required: the provider key Cerberus uses to reach the model.
OPENROUTER_API_KEY=sk-or-v1-your-key-here

# Recommended if you want to use the web dashboard (Manage Proxy / Test Zone or Playground).
# Without it, the /admin API is disabled.
CERBERUS_ADMIN_TOKEN=pick-any-secret

# Everything else has a default:
#   CERBERUS_UPSTREAM=https://openrouter.ai/api/v1
#   CERBERUS_LISTEN=:8080
#   CERBERUS_OUTBOUND_MODE=buffer
```

That is the whole configuration for local use: the OpenRouter key is enough to
proxy requests, and the admin token unlocks the dashboard. Pointing at a
different provider is just a matter of changing `CERBERUS_UPSTREAM` and its key.

### 2. Run the gateway (backend)

Go reads real environment variables, so load your `.env` into the shell first,
then start the gateway:

Bash is the recommended way.
```bash
set -a; source .env; set +a
go run ./cmd/gateway
```



The gateway listens on `http://localhost:8080`. Health check: `GET /healthz`.

On Windows (PowerShell), load `.env` like this instead, then run the same
`go run`:

```powershell
Get-Content .env | Where-Object { $_ -match '=' -and $_ -notmatch '^\s*#' } | ForEach-Object {
  $k, $v = $_ -split '=', 2
  Set-Item "env:$($k.Trim())" $v.Trim()
}
go run ./cmd/gateway
```



### 3. Run the dashboard (frontend)

In a second terminal:

```bash
cd Frontend
npm install
npm run dev
```

Open **http://localhost:5173**. The dev server proxies `/admin` and `/v1` to the
gateway on port 8080, so both halves work together with no extra config. Enter
your admin token on the Manage Proxy page to configure the gateway, issue client
keys, and use the Test Zone.

### 4. Use it as a proxy

To route your own app through Cerberus, generate a client key on the Manage Proxy
page, then point your OpenAI client at the gateway:

```
from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="cbk_your_generated_key",   # a Cerberus key, not your provider key
)
resp = client.chat.completions.create(
    model="openai/gpt-4o",
    messages=[{"role": "user", "content": "Hello"}],
)
```




