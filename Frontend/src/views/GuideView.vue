<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  User,
  LogOut,
  Sliders,
  Bug,
  BookOpen,
  Copy,
  Check,
} from '@lucide/vue'

const router = useRouter()
const activeRoute = computed(() => router.currentRoute.value.name)

const handleLogout = () => {
  localStorage.removeItem('cerberus_auth')
  router.push({ name: 'login' })
}

interface EnvVar {
  name: string
  purpose: string
  example: string
  fallback: string
  tier: 'required' | 'recommended' | 'optional'
  note?: string
}

const vars: EnvVar[] = [
  {
    name: 'CERBERUS_ADMIN_TOKEN',
    purpose: 'Password for the /admin API. Without it the admin API is disabled and these dashboard pages cannot load config, keys, or detectors.',
    example: 'admin-9f3b7c1e-change-me',
    fallback: 'empty = admin API OFF',
    tier: 'required',
    note: 'Operator secret. Never hand this to a customer.',
  },
  {
    name: 'CERBERUS_UPSTREAM',
    purpose: 'The real model API that Cerberus forwards to after scanning.',
    example: 'https://openrouter.ai/api/v1',
    fallback: 'https://openrouter.ai/api/v1',
    tier: 'required',
  },
  {
    name: 'CERBERUS_UPSTREAM_KEY',
    purpose: 'Auth key Cerberus injects toward the upstream. Cerberus swaps the client Authorization header for this before forwarding, so the provider key never leaves the server.',
    example: 'sk-or-v1-0a1b2c3d4e5f6071829',
    fallback: 'empty = client token forwarded as-is',
    tier: 'required',
    note: 'For OpenRouter this is your sk-or-v1-... key. Without it, forwarded requests get 401 from the upstream.',
  },
  {
    name: 'CERBERUS_STATE_PATH',
    purpose: 'File where settings + hashed client keys are persisted. Once set, this file is authoritative and you edit settings live from Manage Proxy (env only seeds the first boot).',
    example: './state.json',
    fallback: 'empty = in-memory, lost on restart',
    tier: 'recommended',
  },
  {
    name: 'CERBERUS_API_KEYS',
    purpose: 'Bootstrap client keys (comma-separated) that callers must present to Cerberus. Leave empty to run open, add keys here or generate them on Manage Proxy.',
    example: 'cbk_devtest1,cbk_devtest2',
    fallback: 'empty = auth OFF (all callers allowed)',
    tier: 'optional',
    note: 'Empty means anyone reaching the port can use the proxy - fine locally, not for production.',
  },
  {
    name: 'CERBERUS_OUTBOUND_MODE',
    purpose: 'How model responses are scanned: off (passthrough), buffer (hold + full scan), stream (scan a rolling window near-live).',
    example: 'buffer',
    fallback: 'buffer',
    tier: 'optional',
  },
  {
    name: 'CERBERUS_LISTEN',
    purpose: 'Address/port the gateway listens on.',
    example: ':8080',
    fallback: ':8080',
    tier: 'optional',
  },
  {
    name: 'CERBERUS_INCOMING_PREFIX',
    purpose: 'URL prefix clients call, stripped before forwarding. You hit /v1/chat/completions.',
    example: '/v1',
    fallback: '/v1',
    tier: 'optional',
  },
  {
    name: 'CERBERUS_MAX_BODY_BYTES',
    purpose: 'Max inbound request size before rejection, protecting the scanner.',
    example: '4194304',
    fallback: '4194304 (4 MiB)',
    tier: 'optional',
  },
]

const tierMeta = {
  required: { label: 'Required', cls: 'border-red-900/40 bg-red-950/10 text-red-400' },
  recommended: { label: 'Recommended', cls: 'border-amber-900/40 bg-amber-950/10 text-amber-400' },
  optional: { label: 'Optional', cls: 'border-zinc-800 bg-zinc-900/40 text-zinc-500' },
}

const groups = [
  { tier: 'required' as const, title: 'Start here', blurb: 'The three you must set to get a real completion back.' },
  { tier: 'recommended' as const, title: 'Recommended', blurb: 'Not required, but you will want it.' },
  { tier: 'optional' as const, title: 'Optional', blurb: 'Safe defaults - only touch these if you have a reason.' },
]

const varsFor = (tier: EnvVar['tier']) => vars.filter((v) => v.tier === tier)

const runTabs = {
  bash: `set -a; source .env; set +a
go run ./cmd/gateway`,
  powershell: `Get-Content .env | Where-Object { $_ -match '=' -and $_ -notmatch '^\\s*#' } | ForEach-Object {
  $k, $v = $_ -split '=', 2
  Set-Item "env:$($k.Trim())" $v.Trim()
}
go run ./cmd/gateway`,
  cmd: `for /f "usebackq tokens=1,* delims==" %A in ("\`.env\`") do set %A=%B
go run ./cmd/gateway`,
}
const activeTab = ref<'bash' | 'powershell' | 'cmd'>('bash')

const copiedKey = ref('')
const copy = (text: string, key: string) => {
  navigator.clipboard.writeText(text)
  copiedKey.value = key
  setTimeout(() => {
    copiedKey.value = ''
  }, 2000)
}
</script>

<template>
  <div class="h-screen flex bg-zinc-950 text-zinc-100 font-sans relative overflow-hidden">
    <div class="absolute inset-0 minimal-dashed z-0 opacity-15 pointer-events-none"></div>

    <aside class="w-64 border-r border-zinc-900 bg-zinc-900/20 backdrop-blur-md flex flex-col justify-between z-10 shrink-0">
      <div>
        <div class="h-16 flex items-center gap-2 px-6 border-b border-zinc-900">
          <span class="font-black tracking-widest text-sm text-white uppercase font-push">Cerberus</span>
        </div>

        <nav class="p-4 space-y-1">
          <router-link
            :to="{ name: 'guide' }"
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold border transition-all"
            :class="activeRoute === 'guide' ? 'border-rose-900/60 bg-rose-950/10 text-rose-455' : 'border-transparent text-zinc-400 hover:bg-zinc-900/60 hover:text-white'"
          >
            <BookOpen class="w-3.5 h-3.5" />
            Quick Start
          </router-link>

          <router-link
            :to="{ name: 'builder' }"
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold border transition-all"
            :class="activeRoute === 'builder' ? 'border-rose-900/60 bg-rose-950/10 text-rose-455' : 'border-transparent text-zinc-400 hover:bg-zinc-900/60 hover:text-white'"
          >
            <Sliders class="w-3.5 h-3.5" />
            Manage Proxy
          </router-link>

          <router-link
            :to="{ name: 'testzone' }"
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold border transition-all"
            :class="activeRoute === 'testzone' ? 'border-rose-900/60 bg-rose-950/10 text-rose-455' : 'border-transparent text-zinc-400 hover:bg-zinc-900/60 hover:text-white'"
          >
            <Bug class="w-3.5 h-3.5" />
            Rule Set & Playground
          </router-link>
        </nav>
      </div>

      <div class="p-4 border-t border-zinc-900">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 rounded-full bg-zinc-850 border border-zinc-800 flex items-center justify-center">
            <User class="w-4 h-4 text-zinc-400" />
          </div>
          <div class="text-left">
            <p class="text-xs font-semibold text-zinc-300">dev_mode</p>
          </div>
          <button
            @click="handleLogout"
            class="ml-auto text-zinc-500 hover:text-white transition-colors p-1.5 rounded hover:bg-zinc-900 cursor-pointer"
            title="Log Out"
          >
            <LogOut class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </aside>

    <main class="flex-1 flex flex-col z-10 min-w-0 overflow-y-auto">
      <div class="p-8 space-y-8 max-w-4xl text-left">
        <!-- Mental model -->
        <div class="cyber-card rounded p-5 border border-zinc-900 bg-zinc-900/10 space-y-2">
          <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-350 font-push">How config works</h3>
          <p class="text-xs text-zinc-400 leading-relaxed">
            The gateway reads these from the <span class="text-zinc-200">environment</span>. Env is the boot seed,
            once <code class="text-zinc-300">CERBERUS_STATE_PATH</code> is set, the saved store becomes the source of
            truth and you change settings live from <span class="text-zinc-200">Manage Proxy</span> - no restart. Only
            three variables really matter to get started, the rest have safe defaults.
          </p>
        </div>

        <!-- Variable groups -->
        <div v-for="g in groups" :key="g.tier" class="space-y-3">
          <div class="flex items-baseline gap-3">
            <h3 class="text-sm font-bold text-white font-push">{{ g.title }}</h3>
            <span
              class="text-[10px] font-mono font-bold uppercase px-1.5 py-0.5 rounded border"
              :class="tierMeta[g.tier].cls"
            >{{ tierMeta[g.tier].label }}</span>
            <span class="text-xs text-zinc-500">{{ g.blurb }}</span>
          </div>

          <div class="space-y-2.5">
            <div
              v-for="v in varsFor(g.tier)"
              :key="v.name"
              class="cyber-card rounded p-4 border border-zinc-900 bg-zinc-950/20 space-y-2"
            >
              <div class="flex items-center justify-between gap-3 flex-wrap">
                <code class="text-xs font-mono font-bold text-zinc-100">{{ v.name }}</code>
                <div class="flex items-center gap-2">
                  <span class="text-[10px] text-zinc-650 font-mono">default: {{ v.fallback }}</span>
                </div>
              </div>
              <p class="text-xs text-zinc-450 leading-relaxed">{{ v.purpose }}</p>
              <div class="flex items-center gap-2 bg-zinc-950 border border-zinc-900 rounded px-2.5 py-1.5">
                <span class="text-[10px] text-zinc-600 uppercase font-push tracking-wider shrink-0">Example</span>
                <code class="text-xs font-mono text-emerald-400/90 truncate">{{ v.example }}</code>
                <button
                  @click="copy(v.name + '=' + v.example, v.name)"
                  class="ml-auto text-zinc-600 hover:text-white p-0.5 rounded transition-colors shrink-0 cursor-pointer"
                  title="Copy line"
                >
                  <Check v-if="copiedKey === v.name" class="w-3 h-3 text-emerald-400" />
                  <Copy v-else class="w-3 h-3" />
                </button>
              </div>
              <p v-if="v.note" class="text-xs text-amber-500/80 leading-relaxed">{{ v.note }}</p>
            </div>
          </div>
        </div>

        <!-- Run it -->
        <div class="space-y-3">
          <h3 class="text-sm font-bold text-white font-push">Run the gateway</h3>
          <p class="text-xs text-zinc-500 leading-relaxed">
            The gateway reads real environment variables, so load your <code class="text-zinc-400">.env</code> into the
            shell first. Pick your shell:
          </p>

          <div class="flex border-b border-zinc-900 text-xs font-mono">
            <button
              v-for="tab in (['bash', 'powershell', 'cmd'] as const)"
              :key="tab"
              @click="activeTab = tab"
              class="px-3 py-1.5 border-b font-semibold capitalize cursor-pointer transition-all"
              :class="activeTab === tab ? 'border-zinc-300 text-white bg-zinc-900/40' : 'border-transparent text-zinc-500 hover:text-white'"
            >
              {{ tab === 'bash' ? 'bash / Git Bash' : tab === 'powershell' ? 'PowerShell' : 'cmd.exe' }}
            </button>
          </div>

          <div class="relative">
            <button
              @click="copy(runTabs[activeTab], 'run')"
              class="absolute top-2.5 right-2.5 text-zinc-500 hover:text-white p-1 rounded hover:bg-zinc-800 transition-colors cursor-pointer"
              title="Copy"
            >
              <Check v-if="copiedKey === 'run'" class="w-3 h-3 text-emerald-400" />
              <Copy v-else class="w-3 h-3" />
            </button>
            <div class="cyber-card rounded p-4 font-mono text-xs text-zinc-400 bg-zinc-950 border border-zinc-900 overflow-x-auto">
              <pre class="leading-relaxed">{{ runTabs[activeTab] }}</pre>
            </div>
          </div>
        </div>

        <!-- Next steps -->
        <div class="cyber-card rounded p-5 border border-zinc-900 bg-zinc-900/10 space-y-3">
          <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-350 font-push">Then what</h3>
          <ol class="space-y-2 text-xs text-zinc-400 leading-relaxed list-decimal list-inside">
            <li>Open <router-link :to="{ name: 'builder' }" class="text-zinc-200 underline hover:text-white">Manage Proxy</router-link>, paste your admin token, and confirm settings sync (upstream_key shows as set).</li>
            <li>Use the <router-link :to="{ name: 'testzone' }" class="text-zinc-200 underline hover:text-white">Rule Set & Playground</router-link> to fire attack presets and watch verdicts - toggle detectors on and off live.</li>
            <li>To use it as a real proxy, generate a client key on Manage Proxy, then point your app's base URL at <code class="text-zinc-300">http://localhost:8080/v1</code> with that key.</li>
          </ol>
        </div>
      </div>
    </main>
  </div>
</template>
