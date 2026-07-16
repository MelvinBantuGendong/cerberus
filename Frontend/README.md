# 🛡️ Cerberus Gateway | Frontend Control Plane

This is the frontend dashboard for **Cerberus**, a Visual Control Plane and Zero-Trust Reverse Proxy built to secure AI Agents and MCP endpoints.

Built for Track 2 (Digital Safety & Privacy) using **Vue 3**, **Vite**, **TypeScript**, **PrimeVue 4**, and **Tailwind CSS v4**.

---

## 🎨 Theme: Cerberus Crimson
Following the thematic direction of the project (Cerberus, the multi-headed guardian of the underworld), the frontend uses the **Cerberus Crimson** theme preset:
- **Primary Accents**: Vibrant `rose` color scheme, highlighting active warnings and mitigations.
- **Surfaces**: Dark `zinc` background palette to create a premium, high-tech security operations center.
- **Typography**: Sleek geometric typography utilizing **Outfit** and **Inter** from Google Fonts.

---

## 📋 Features Implemented

### 1. Frictionless Authentication (`LoginView.vue`)
- Developer-centric access gate featuring a **live booting system terminal log stream** (simulating core engine handshakes, key registration, and OAuth verification).
- One-click simulated GitHub OAuth authorization that stores session details in local RAM state.

### 2. Three-Step Pipeline Builder (`BuilderView.vue`)
- **Vertical Canvas**: Displays the request pipeline flow (Incoming Agent Request → Vetting Middleware Guardrails → Outgoing API Forwarding).
- **Interactive Component Cards**:
  - **Prompt Injection Guard**: Similarity sensitivity threshold slider (30-99%) and custom string blacklists.
  - **Destructive Command Firewall**: Command argument blacklists (e.g. `rm -rf`, `DROP TABLE`).
  - **PII & Leak Masker**: Checkbox selectors for PII types (API keys, IP addresses, credit cards) and custom mask string token.
  - **Budget Governor**: TPM limits and daily cost cap parameters.
- **Live Compiled JSON Schema**: Real-time reactive code viewport showing the compiled JSON configuration payload.
- **One-Line SDK Integration**: Copyable code snippet panel with tabs for **Javascript Fetch**, **Python SDK**, and **cURL**.

### 3. Threat Intelligence Center (`AnalyticsView.vue`)
- **Metric Cards**: Real-time telemetry monitoring total requests routed, average latencies (Go overhead), transactions per second, and threats blocked.
- **Live Traffic Simulation**: Spawns continuous traffic streams with normal completions, PII redactions, and prompt injection attacks.
- **Event Feed Table**: Scrolling real-time table of transactions. Attack events **flash red dynamically** and reveal the exact blocked injection/tool payload inputs. Includes controllers to **Pause/Resume** or clear the stream.

---

## 🚀 Running the Project

### Development Server
To launch the interactive hot-reload dev server locally:
```bash
npm install
npm run dev
```

### Production Build
To verify type-checking and compile for static deployment:
```bash
npm run build
```
The compiled assets will be bundled in the `dist/` directory.
