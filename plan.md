# 🛡️ Project Blueprint: Vibe-Shield (Garudie Proxy)
**Theme:** Secure Our Future (Hackathon Track 2: Digital Safety & Privacy)
**Concept:** A No-Code Visual Control Plane and Zero-Trust Reverse Proxy built to secure AI Agents and MCP endpoints for rapid ("Vibe") developers.

---

## 📋 Core Architectural Overview
[Vibe Coder's App]
│
▼ (Changes API Base URL to point here)
┌────────────────────────────────────────────────────────┐
│               CERBERUS PROXY (Go Backend)              │
│                                                        │
│  1. Guardrail Engine (Ingests Config JSON from UI)     │
│  2. Security Filters (Scan Inputs & Block Violations)  │
│  3. Ephemeral Transport (Passes keys securely in RAM)  |
|  4. etc.                                               |
└────────────────────────────────────────────────────────┘
│
▼ (Forwarded ONLY if clean)
[Target AI Service / MCP Endpoint] (OpenAI, Anthropic, etc.)

---

## 🎯 High-Priority Threat Matrix (OWASP LLM Top 10 Target)

1. **Direct Prompt Injection & Obfuscation**
   * *Problem:* Adversarial text payloads targeting model boundaries ("Ignore system prompts").
   * *Defense:* Go parsing engine employing semantic string-normalization, regex-matching arrays, and lightweight similarity lookups.
2. **Excessive Agency & Harmful Command Execution**
   * *Problem:* Injected prompts hijacking Model Context Protocol (MCP) or server tools to execute destructive terminal or DB commands.
   * *Defense:* Middleware firewall that intercepts JSON tool payloads before execution, dropping connection on non-whitelisted actions (`DROP TABLE`, `rm -rf`, etc.).
3. **Data Leaks & PII Exfiltration**
   * *Problem:* Models accidentally leaking internal context, system instructions, or sensitive database keys back to client strings.
   * *Defense:* Output scanning interceptor that redacts or masks sensitive data structural variables dynamically before reaching the user interface.

---

## Frontend & UX Specifications (SvelteKit / Vue 3 + Tailwind CSS)

### Page 1: Frictionless Authentication
* Minimalist layout leveraging **Supabase/Firebase Auth** for instantaneous **One-Click GitHub OAuth** login, targeting the developer user persona perfectly.

### Page 2: Three-Step Pipeline Builder
* **Layout:** A clean, structured vertical canvas representing the request flow lifecycle rather than an infinite node graph.
* **Draggable Elements:** Modular component cards (e.g., `Prompt Injection Filter`, `Destructive Command Firewall`, `PII Masker`) dropped into specific processing blocks using native HTML5 drag-and-drop mechanics or lightweight visual wrappers (`shadcn-svelte` / `PrimeVue`).
* **Deployment Output:** Compiles visual states to a clean configuration JSON, generating a one-line proxy endpoint hook (`https://api.garudie.tech/proxy/v1/shield_abc123`).

### Page 3: Real-Time Threat Intel Analytics
* **Metric Cards:** Real-time metrics highlighting total requests processed, intrusion vectors mitigated, and Go network execution latency.
* **Activity Log:** A scrolling table of security events. Detected attacks animate/flash red instantly to display the exact malicious input string intercepted and neutralised at the proxy layer.

---

## Zero-Trust Integration Vector (The Developer Workflow)
To avoid the security liability of storing root enterprise API credentials on a database, the system operates as an **ephemeral pass-through proxy**:

1. Developer configures security parameters visually on the frontend dashboard.
2. Developer replaces their local AI library target configuration endpoint with the Vibe-Shield Proxy link:
   ```javascript
   // Before: Direct API access
   const response = await fetch("https://api.openai.com/v1/chat/completions", { ... });

   // After: One-line proxy implementation
   const response = await fetch("https://api.garudie.tech/v1/proxy/shield_abc123", { ... });
   ```
