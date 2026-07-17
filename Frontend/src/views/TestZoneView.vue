<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  User,
  LogOut,
  Sliders,
  Bug,
  Play,
  ShieldAlert,
  Shield,
  ArrowRight,
} from '@lucide/vue'

const router = useRouter()
const activeRoute = computed(() => router.currentRoute.value.name)

const handleLogout = () => {
  localStorage.removeItem('cerberus_auth')
  router.push({ name: 'login' })
}

const adminToken = ref(localStorage.getItem('cerberus_admin_token') || '')
const clientKey = ref(localStorage.getItem('cerberus_last_client_key') || '')
const model = ref(localStorage.getItem('cerberus_test_model') || 'openai/gpt-4o')

const persistCreds = () => {
  localStorage.setItem('cerberus_admin_token', adminToken.value)
  localStorage.setItem('cerberus_last_client_key', clientKey.value)
  localStorage.setItem('cerberus_test_model', model.value)
}

interface Detector {
  id: string
  name: string
  description: string
  direction: 'inbound' | 'outbound'
}

const catalog = ref<Detector[]>([])
const disabled = ref<Set<string>>(new Set())
const configError = ref('')
const savingRule = ref<string | null>(null)

const upstream = ref('')
const outboundMode = ref('buffer')
const maxBodyBytes = ref(4194304)
const isDeploying = ref(false)
const justDeployed = ref(false)

const inboundDetectors = computed(() => catalog.value.filter((d) => d.direction === 'inbound'))
const outboundDetectors = computed(() => catalog.value.filter((d) => d.direction === 'outbound'))

const authHeaders = () => ({
  'Content-Type': 'application/json',
  Authorization: `Bearer ${adminToken.value}`,
})

const loadCatalog = async () => {
  configError.value = ''
  if (!adminToken.value) {
    configError.value = 'Enter the admin token to load and toggle detectors.'
    return
  }
  try {
    const [detRes, cfgRes] = await Promise.all([
      fetch('/admin/detectors', { headers: authHeaders() }),
      fetch('/admin/config', { headers: authHeaders() }),
    ])
    if (!detRes.ok || !cfgRes.ok) {
      throw new Error(`HTTP ${detRes.status}/${cfgRes.status}`)
    }
    catalog.value = await detRes.json()
    const cfg = await cfgRes.json()
    disabled.value = new Set<string>(cfg.disabled_rules || [])

    // Set configuration variables
    upstream.value = cfg.upstream || ''
    outboundMode.value = cfg.outbound_mode || 'buffer'
    maxBodyBytes.value = cfg.max_body_bytes || 4194304
  } catch (err: any) {
    configError.value = `Could not reach the admin API (${err.message}). Verify the token and that the Go gateway is running with CERBERUS_ADMIN_TOKEN set.`
  }
}

const deployGateway = async () => {
  isDeploying.value = true
  configError.value = ''
  justDeployed.value = false
  try {
    const res = await fetch('/admin/config', {
      method: 'PUT',
      headers: authHeaders(),
      body: JSON.stringify({
        upstream: upstream.value,
        max_body_bytes: Number(maxBodyBytes.value),
        outbound_mode: outboundMode.value,
        disabled_rules: Array.from(disabled.value)
      })
    })
    if (!res.ok) {
      throw new Error(await res.text())
    }
    justDeployed.value = true
    setTimeout(() => {
      justDeployed.value = false
    }, 3000)
  } catch (err: any) {
    configError.value = `Sync failed: ${err.message}`
  } finally {
    isDeploying.value = false
  }
}

const toggleRule = async (id: string) => {
  if (!adminToken.value) {
    configError.value = 'Enter the admin token to toggle detectors.'
    return
  }
  const next = new Set(disabled.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)

  savingRule.value = id
  configError.value = ''
  try {
    const res = await fetch('/admin/config', {
      method: 'PUT',
      headers: authHeaders(),
      body: JSON.stringify({ disabled_rules: Array.from(next) }),
    })
    if (!res.ok) throw new Error(await res.text())
    const cfg = await res.json()
    disabled.value = new Set<string>(cfg.disabled_rules || [])
  } catch (err: any) {
    configError.value = `Failed to update detector: ${err.message}`
  } finally {
    savingRule.value = null
  }
}

const isEnabled = (id: string) => !disabled.value.has(id)

const promptText = ref('')
const sending = ref(false)

interface Result {
  status: number
  action: 'allow' | 'block' | 'flag' | 'error' | 'forwarded'
  verdict: any | null
  message: string
}
const result = ref<Result | null>(null)

const sendAttack = async () => {
  if (!promptText.value.trim()) return
  sending.value = true
  result.value = null
  persistCreds()

  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (clientKey.value) headers['Authorization'] = `Bearer ${clientKey.value}`

  try {
    const res = await fetch('/v1/chat/completions', {
      method: 'POST',
      headers,
      body: JSON.stringify({
        model: model.value,
        messages: [{ role: 'user', content: promptText.value }],
      }),
    })

    if (res.status === 403) {
      const verdict = await res.json()
      result.value = {
        status: 403,
        action: 'block',
        verdict,
        message: 'Blocked by Cerberus inbound scan - request never reached the upstream model.',
      }
    } else if (res.status === 401) {

      const text = await res.text()
      const fromCerberus = text.trim().toLowerCase() === 'unauthorized'
      result.value = {
        status: 401,
        action: 'error',
        verdict: null,
        message: fromCerberus
          ? 'Cerberus rejected the request: auth is enabled and the client key is missing or invalid. Supply a valid client key above (generate one on Manage Proxy).'
          : 'Passed the inbound scan and was forwarded, but the upstream returned 401 - set the upstream key (CERBERUS_UPSTREAM_KEY / upstream authorization) on the Manage Proxy page.',
      }
    } else if (res.ok) {
      const outbound = res.headers.get('X-Cerberus-Outbound')
      let body: any = null
      try {
        body = await res.json()
      } catch {

      }
      const output = body?.choices?.[0]?.message?.content ?? '(forwarded to upstream - no chat content returned)'
      result.value = {
        status: res.status,
        action: outbound === 'block' ? 'block' : outbound === 'flag' ? 'flag' : 'allow',
        verdict: null,
        message:
          outbound === 'block'
            ? 'Response blocked by the outbound (egress) scan.'
            : outbound === 'flag'
              ? `Response flagged by the outbound scan. Model said: ${output}`
              : `Passed the inbound scan and was forwarded. Model said: ${output}`,
      }
    } else {

      const text = await res.text()
      let detail = text
      try {
        const j = JSON.parse(text)
        detail = j?.error?.message || j?.error || j?.message || text
      } catch {

      }
      result.value = {
        status: res.status,
        action: 'forwarded',
        verdict: null,
        message: `Passed the inbound scan and was forwarded, but the upstream returned HTTP ${res.status}: ${detail || '(empty body)'}`,
      }
    }
  } catch (err: any) {
    result.value = {
      status: 0,
      action: 'error',
      verdict: null,
      message: `Request failed: ${err.message}. Is the Go gateway running on port 8080?`,
    }
  } finally {
    sending.value = false
  }
}

onMounted(loadCatalog)
</script>

<template>
  <div class="h-screen flex bg-zinc-950 text-zinc-100 font-sans relative overflow-hidden">
    <div class="absolute inset-0 minimal-dashed z-0 opacity-15 pointer-events-none"></div>

    <!-- Navigation Sidebar -->
    <aside class="w-64 border-r border-zinc-900 bg-zinc-900/20 backdrop-blur-md flex flex-col justify-between z-10 shrink-0">
      <div>
        <div class="h-16 flex items-center gap-3 px-6 border-b border-zinc-900">
          <span class="font-bold tracking-wider text-xs text-white font-push">Cerberus</span>
        </div>

        <nav class="p-4 space-y-1">
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
            Test Zone
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

    <!-- Main Content -->
    <main class="flex-1 flex flex-col z-10 min-w-0 overflow-y-auto">


      <div class="grid lg:grid-cols-12 flex-1 min-h-0">
        <!-- Left: Detector toggles -->
        <div class="lg:col-span-5 p-6 border-r border-zinc-900 space-y-6 text-left overflow-y-auto">
          <!-- Admin token -->
          <div class="space-y-1.5">
            <label class="text-[10px] font-semibold text-zinc-350 font-push uppercase tracking-wider">Admin Token</label>
            <div class="flex gap-2">
              <input
                type="password"
                v-model="adminToken"
                @change="persistCreds"
                placeholder="CERBERUS_ADMIN_TOKEN"
                class="flex-1 text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono"
              />
              <button
                @click="loadCatalog"
                class="text-xs font-bold text-rose-455 border border-rose-900/40 bg-rose-950/10 hover:bg-rose-900/25 hover:text-rose-200 px-3 py-1.5 rounded cursor-pointer transition-all font-push"
              >
                Load
              </button>
            </div>
          </div>

          <!-- Gateway Settings Sync -->
          <div class="space-y-3.5 border-t border-zinc-900 pt-5">
            <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-350 font-push">Gateway Settings</h3>
            <div class="grid grid-cols-2 gap-4">
              <div class="col-span-2 space-y-1.5">
                <label class="text-[10px] font-semibold text-zinc-400 font-push">Upstream Destination</label>
                <input 
                  type="text" 
                  v-model="upstream"
                  placeholder="e.g. https://openrouter.ai/api/v1"
                  class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:outline-none font-mono"
                />
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-semibold text-zinc-400 font-push">Outbound Mode</label>
                <select 
                  v-model="outboundMode"
                  class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:outline-none font-push"
                >
                  <option value="off">Off (Passthrough)</option>
                  <option value="buffer">Buffer (Scan Full)</option>
                  <option value="stream">Stream (Scan SSE)</option>
                </select>
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-semibold text-zinc-400 font-push">Max Payload (Bytes)</label>
                <input 
                  type="number" 
                  v-model.number="maxBodyBytes"
                  class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:outline-none font-mono"
                />
              </div>
            </div>
            <button
              @click="deployGateway"
              :disabled="isDeploying"
              class="w-full flex items-center justify-center gap-2 border border-rose-900/60 bg-rose-950/15 text-rose-455 hover:bg-rose-900/30 hover:text-rose-200 font-bold text-xs py-2 px-4 rounded transition-all duration-150 disabled:opacity-50 cursor-pointer font-push"
            >
              <Play class="w-3 h-3 fill-current" />
              <span v-if="!isDeploying">Sync Config to Proxy</span>
              <span v-else>Syncing...</span>
            </button>
            <div v-if="justDeployed" class="text-[10px] text-emerald-450 font-mono text-center mt-1.5">
              Gateway config successfully synced.
            </div>
          </div>

          <div v-if="configError" class="text-xs text-red-400 font-mono leading-relaxed bg-red-950/20 border border-red-900/30 p-2.5 rounded">
            {{ configError }}
          </div>

          <!-- Inbound detectors -->
          <div class="space-y-3">
            <div class="flex items-center gap-2">
              <ShieldAlert class="w-3.5 h-3.5 text-red-400" />
              <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-350 font-push">Inbound Detectors</h3>
            </div>
            <div class="space-y-2">
              <div
                v-for="d in inboundDetectors"
                :key="d.id"
                class="cyber-card rounded p-3 border border-zinc-900 bg-zinc-950/20 flex items-start justify-between gap-3"
              >
                <div class="space-y-0.5 min-w-0">
                  <p class="text-xs font-bold text-zinc-200 font-push">{{ d.name }}</p>
                  <p class="text-xs text-zinc-550 leading-relaxed">{{ d.description }}</p>
                  <code class="text-[10px] text-zinc-600 font-mono">{{ d.id }}</code>
                </div>
                <button
                  @click="toggleRule(d.id)"
                  :disabled="savingRule === d.id"
                  class="shrink-0 mt-0.5 w-9 h-5 rounded-full relative transition-colors duration-200 cursor-pointer disabled:opacity-50"
                  :class="isEnabled(d.id) ? 'bg-emerald-600/80' : 'bg-zinc-800'"
                  :title="isEnabled(d.id) ? 'Enabled - click to disable' : 'Disabled - click to enable'"
                >
                  <span
                    class="absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white transition-transform duration-200"
                    :class="isEnabled(d.id) ? 'translate-x-4' : 'translate-x-0'"
                  ></span>
                </button>
              </div>
              <p v-if="inboundDetectors.length === 0" class="text-xs text-zinc-600 font-mono">No detectors loaded.</p>
            </div>
          </div>

          <!-- Outbound detectors -->
          <div class="space-y-3">
            <div class="flex items-center gap-2">
              <Shield class="w-3.5 h-3.5 text-amber-400" />
              <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-350 font-push">Outbound Detectors</h3>
            </div>
            <p class="text-xs text-zinc-500 leading-relaxed">
              Outbound detectors scan the model's <span class="text-zinc-300">response</span>, not your prompt. Trigger them with a live upstream that echoes secrets/PII, using outbound mode <code class="text-zinc-400">buffer</code> or <code class="text-zinc-400">stream</code>.
            </p>
            <div class="space-y-2">
              <div
                v-for="d in outboundDetectors"
                :key="d.id"
                class="cyber-card rounded p-3 border border-zinc-900 bg-zinc-950/20 flex items-start justify-between gap-3"
              >
                <div class="space-y-0.5 min-w-0">
                  <p class="text-xs font-bold text-zinc-200 font-push">{{ d.name }}</p>
                  <p class="text-xs text-zinc-550 leading-relaxed">{{ d.description }}</p>
                  <code class="text-[10px] text-zinc-600 font-mono">{{ d.id }}</code>
                </div>
                <button
                  @click="toggleRule(d.id)"
                  :disabled="savingRule === d.id"
                  class="shrink-0 mt-0.5 w-9 h-5 rounded-full relative transition-colors duration-200 cursor-pointer disabled:opacity-50"
                  :class="isEnabled(d.id) ? 'bg-emerald-600/80' : 'bg-zinc-800'"
                  :title="isEnabled(d.id) ? 'Enabled - click to disable' : 'Disabled - click to enable'"
                >
                  <span
                    class="absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white transition-transform duration-200"
                    :class="isEnabled(d.id) ? 'translate-x-4' : 'translate-x-0'"
                  ></span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Right: Attack tester -->
        <div class="lg:col-span-7 p-6 space-y-5 text-left overflow-y-auto">
          <div class="grid grid-cols-2 gap-4">
            <div class="space-y-1.5">
              <label class="text-xs font-semibold text-zinc-350 font-push uppercase tracking-wider">Model</label>
              <input
                type="text"
                v-model="model"
                @change="persistCreds"
                placeholder="e.g. openai/gpt-4o"
                class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono"
              />
              <p class="text-[10px] text-zinc-550 leading-relaxed">Forwarded verbatim to the upstream. OpenRouter uses namespaced slugs (<code class="text-zinc-400">openai/gpt-4o</code>, <code class="text-zinc-400">anthropic/claude-3.5-sonnet</code>).</p>
            </div>
            <div class="space-y-1.5">
              <label class="text-xs font-semibold text-zinc-350 font-push uppercase tracking-wider">Client Key</label>
              <input
                type="password"
                v-model="clientKey"
                @change="persistCreds"
                placeholder="cbk_..."
                class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono"
              />
            </div>
          </div>

          <!-- Prompt input -->
          <div class="space-y-1.5">
            <label class="text-xs font-semibold text-zinc-350 font-push uppercase tracking-wider">Prompt</label>
            <textarea
              v-model="promptText"
              rows="4"
              placeholder="Type your prompt, then send it through the proxy..."
              class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-300 p-3 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono placeholder:text-zinc-700 resize-none"
            ></textarea>
            <button
              @click="sendAttack"
              :disabled="sending || !promptText.trim()"
              class="w-full flex items-center justify-center gap-2 border border-rose-900/60 bg-rose-950/15 text-rose-455 hover:bg-rose-900/30 hover:text-rose-200 font-bold text-xs py-2.5 px-4 rounded transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer font-push"
            >
              <Play class="w-3.5 h-3.5 fill-current" />
              <span v-if="!sending">Send Through Proxy</span>
              <span v-else>Scanning...</span>
            </button>
          </div>

          <!-- Result -->
          <div v-if="result" class="space-y-3">
            <div
              class="cyber-card rounded p-4 border space-y-3"
              :class="[
                result.action === 'block' ? 'border-red-900/50 bg-red-950/10' :
                result.action === 'flag' ? 'border-amber-900/50 bg-amber-950/10' :
                result.action === 'allow' ? 'border-emerald-900/40 bg-emerald-950/10' :
                result.action === 'error' ? 'border-zinc-800 bg-zinc-900/20' :
                'border-zinc-800 bg-zinc-900/20'
              ]"
            >
              <div class="flex items-center gap-2">
                <span
                  class="px-2 py-0.5 rounded border text-[10px] font-mono font-bold uppercase inline-flex items-center gap-1"
                  :class="[
                    result.action === 'block' ? 'border-red-900 bg-red-950/30 text-red-400' :
                    result.action === 'flag' ? 'border-amber-900 bg-amber-950/30 text-amber-400' :
                    result.action === 'allow' ? 'border-emerald-900 bg-emerald-950/30 text-emerald-400' :
                    'border-zinc-800 bg-zinc-900 text-zinc-400'
                  ]"
                >
                  <ShieldAlert v-if="result.action === 'block'" class="w-2.5 h-2.5" />
                  <Shield v-else class="w-2.5 h-2.5" />
                  {{ result.action === 'block' ? 'Blocked' : result.action === 'flag' ? 'Flagged' : result.action === 'allow' ? 'Allowed' : result.action === 'forwarded' ? 'Forwarded' : 'Error' }}
                </span>
                <span class="text-[10px] text-zinc-500 font-mono flex items-center gap-1">
                  HTTP {{ result.status || '-' }}
                  <ArrowRight class="w-2.5 h-2.5" />
                  {{ result.action === 'block' ? 'stopped at gateway' : 'passed inbound scan' }}
                </span>
              </div>
              <p class="text-xs text-zinc-300 leading-relaxed font-mono">{{ result.message }}</p>
            </div>

            <!-- Verdict JSON -->
            <div v-if="result.verdict" class="space-y-1.5">
              <span class="text-xs text-zinc-400 font-bold uppercase tracking-wider font-push">Verdict Payload</span>
              <div class="cyber-card rounded p-3 font-mono text-xs text-zinc-450 bg-zinc-950 border border-zinc-900 overflow-y-auto max-h-56">
                <pre class="leading-relaxed">{{ JSON.stringify(result.verdict, null, 2) }}</pre>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>
