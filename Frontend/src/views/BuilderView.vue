<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  Play,
  Copy,
  Check,
  User,
  LogOut,
  Sliders,
  Trash2,
  Bug
} from '@lucide/vue'

const router = useRouter()
const activeRoute = computed(() => router.currentRoute.value.name)

const handleLogout = () => {
  localStorage.removeItem('cerberus_auth')
  router.push({ name: 'login' })
}

const isDeploying = ref(false)
const justDeployed = ref(false)
const copied = ref(false)
const activeTab = ref<'js' | 'python' | 'curl'>('js')

const showKeyModal = ref(false)
const generatedKey = ref('')
const generatedKeyLabel = ref('')
const copiedGeneratedKey = ref(false)
const keyGenError = ref('')

const copyGeneratedKey = () => {
  navigator.clipboard.writeText(generatedKey.value)
  copiedGeneratedKey.value = true
  setTimeout(() => {
    copiedGeneratedKey.value = false
  }, 2000)
}

interface Project {
  id: string
  name: string
  proxyUrl: string
  status: 'active' | 'inactive'
  activeGuardsCount: number
}

const projects = ref<Project[]>([
  {
    id: 'gateway_cluster',
    name: 'Go Gateway Core',
    proxyUrl: 'http://localhost:8080/v1/chat/completions',
    status: 'active',
    activeGuardsCount: 4
  }
])

const activeProjectId = ref('gateway_cluster')
const activeProject = computed(() => {
  return projects.value.find(p => p.id === activeProjectId.value) || projects.value[0]
})


const guardsConfig = ref({
  promptInjection: {
    name: 'Ingress Sentry (Left Head)',
    description: 'Blocks adversarial text payloads',
    similarityThreshold: 75,
    customPatterns: '',
    threatAction: 'BLOCK_CONNECTION'
  },
  toolFirewall: {
    name: 'Runtime Firewall (Middle Head)',
    description: 'Blocks hazardous system calls',
    commandBlacklist: '',
    alertSeverity: 'CRITICAL'
  },
  piiMasker: {
    name: 'Egress Censor (Right Head) - PII',
    description: 'Redacts outgoing sensitive variables',
    maskString: '[REDACTED_SECURE]',
    maskTypes: [] as string[],
    threatAction: 'REDACT'
  },
  rateLimiter: {
    name: 'Egress Censor (Right Head) - Governor',
    description: 'Enforces hard ceilings and token limits',
    maxTokensPerMinute: 0,
    costCeilingPerDay: 0.0
  },
  gatewaySettings: {
    name: 'Gateway Core Config',
    description: 'Binds parameters directly to the active Go backend proxy',
    upstream: '',
    upstreamKey: '',
    maxBodyBytes: 0,
    outboundMode: 'off'
  }
})

const compiledJson = computed(() => {
  return JSON.stringify({
    upstream: guardsConfig.value.gatewaySettings.upstream || null,
    max_body_bytes: Number(guardsConfig.value.gatewaySettings.maxBodyBytes) || 0,
    outbound_mode: guardsConfig.value.gatewaySettings.outboundMode || 'off',
    admin_token_configured: !!adminToken.value,
    client_auth_keys_count: apiKeys.value.length
  }, null, 2)
})

const adminToken = ref(localStorage.getItem('cerberus_admin_token') || 'test_token')
const syncError = ref('')
const apiKeys = ref<any[]>([])
const newKeyLabel = ref('')

interface Detector {
  id: string
  name: string
  direction: 'inbound' | 'outbound'
}
const detectors = ref<Detector[]>([])
const disabledRules = ref<string[]>([])
const enabledCount = computed(() => detectors.value.filter(d => !disabledRules.value.includes(d.id)).length)

const fetchDetectors = async () => {
  if (!adminToken.value) return
  try {
    const res = await fetch('/admin/detectors', {
      headers: { 'Authorization': `Bearer ${adminToken.value}` }
    })
    if (res.ok) detectors.value = await res.json()
  } catch (err) {
    console.error('Failed to fetch detectors:', err)
  }
}

const fetchBackendConfig = async () => {
  if (!adminToken.value) return
  syncError.value = ''
  try {
    const res = await fetch('/admin/config', {
      headers: {
        'Authorization': `Bearer ${adminToken.value}`
      }
    })
    if (!res.ok) {
      throw new Error(`HTTP error ${res.status}`)
    }
    const data = await res.json()
    guardsConfig.value.gatewaySettings.upstream = data.upstream
    guardsConfig.value.gatewaySettings.maxBodyBytes = data.max_body_bytes
    guardsConfig.value.gatewaySettings.outboundMode = data.outbound_mode
    disabledRules.value = data.disabled_rules || []
  } catch (err: any) {
    console.error('Failed to sync with backend:', err)
    syncError.value = 'Could not sync with Go backend. Verify token and port.'
  }
}

const fetchApiKeys = async () => {
  if (!adminToken.value) return
  try {
    const res = await fetch('/admin/keys', {
      headers: {
        'Authorization': `Bearer ${adminToken.value}`
      }
    })
    if (res.ok) {
      apiKeys.value = await res.json()
    }
  } catch (err) {
    console.error('Failed to fetch keys:', err)
  }
}

const generateNewKey = async () => {
  if (!adminToken.value) return
  keyGenError.value = ''
  try {
    const res = await fetch('/admin/keys', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${adminToken.value}`
      },
      body: JSON.stringify({ label: newKeyLabel.value || 'Dashboard Key' })
    })
    if (res.ok) {
      const data = await res.json()

      localStorage.setItem('cerberus_last_client_key', data.key)
      
      generatedKey.value = data.key
      generatedKeyLabel.value = data.label || 'Dashboard Key'
      showKeyModal.value = true

      newKeyLabel.value = ''
      await fetchApiKeys()
    } else {
      const txt = await res.text()
      keyGenError.value = `Failed: ${txt}`
    }
  } catch (err: any) {
    keyGenError.value = `Error: ${err.message}`
  }
}

const deleteApiKey = async (id: string) => {
  if (!adminToken.value) return
  if (!confirm('Are you sure you want to revoke this client API key?')) return
  try {
    const res = await fetch(`/admin/keys/${id}`, {
      method: 'DELETE',
      headers: {
        'Authorization': `Bearer ${adminToken.value}`
      }
    })
    if (res.ok) {
      await fetchApiKeys()
    }
  } catch (err) {
    console.error('Failed to delete key:', err)
  }
}

const deployGateway = async () => {
  isDeploying.value = true
  syncError.value = ''
  try {
    const res = await fetch('/admin/config', {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${adminToken.value}`
      },
      body: JSON.stringify({
        upstream: guardsConfig.value.gatewaySettings.upstream,
        upstream_key: guardsConfig.value.gatewaySettings.upstreamKey || undefined,
        max_body_bytes: Number(guardsConfig.value.gatewaySettings.maxBodyBytes),
        outbound_mode: guardsConfig.value.gatewaySettings.outboundMode
      })
    })
    if (!res.ok) {
      const errMsg = await res.text()
      throw new Error(errMsg || `HTTP error ${res.status}`)
    }
    justDeployed.value = true

    const current = projects.value.find(p => p.id === activeProjectId.value)
    if (current) {
      current.status = 'active'
    }

    setTimeout(() => {
      justDeployed.value = false
    }, 3000)
    await fetchBackendConfig()
  } catch (err: any) {
    console.error('Deployment failed:', err)
    syncError.value = `Deployment failed: ${err.message}`
  } finally {
    isDeploying.value = false
  }
}

onMounted(() => {
  fetchBackendConfig()
  fetchApiKeys()
  fetchDetectors()
})

watch(adminToken, (newVal) => {
  localStorage.setItem('cerberus_admin_token', newVal)
  fetchBackendConfig()
  fetchApiKeys()
  fetchDetectors()
})

const proxyEndpoint = computed(() => window.location.origin)

const codeSnippets = computed(() => {
  return {
    js: `// Initialize Cerberus Secure Pass-through Proxy
const response = await fetch("${proxyEndpoint.value}/v1/chat/completions", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "Authorization": "Bearer YOUR_CLIENT_KEY",
  },
  body: JSON.stringify({
    model: "gpt-4o",
    messages: [{ role: "user", content: "Analyze file system path..." }]
  })
});

const data = await response.json();
console.log(data);`,
    python: `# Connect your OpenAI / Anthropic agent securely
import openai

client = openai.OpenAI(
    base_url="${proxyEndpoint.value}/v1",
    api_key="YOUR_CLIENT_KEY"
)

completion = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Execute agent tools..."}]
)
print(completion.choices[0].message.content)`,
    curl: `curl ${proxyEndpoint.value}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_CLIENT_KEY" \\
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Test prompt injection vector..."}]
  }'`
  }
})

const copySnippet = () => {
  const currentSnippet = codeSnippets.value[activeTab.value]
  navigator.clipboard.writeText(currentSnippet)
  copied.value = true
  setTimeout(() => {
    copied.value = false
  }, 2000)
}
</script>

<template>
  <div class="h-screen flex bg-zinc-950 text-zinc-100 font-sans relative overflow-hidden">
    <div class="absolute inset-0 minimal-dashed z-0 opacity-15 pointer-events-none"></div>

    <!-- Navigation Sidebar -->
    <aside class="w-64 border-r border-zinc-900 bg-zinc-900/20 backdrop-blur-md flex flex-col justify-between z-10">
      <div>
        <!-- Brand Header -->
        <div class="h-16 flex items-center gap-3 px-6 border-b border-zinc-900">
          <span class="font-bold tracking-wider text-xs text-white font-push">Cerberus</span>
        </div>

        <!-- Navigation Links -->
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

      <!-- User Profile Footer -->
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

    <!-- Main Content Area -->
    <main class="flex-1 flex flex-col z-10 min-w-0 overflow-y-auto">



      <!-- Layout Panel Grid -->
      <div class="grid lg:grid-cols-12">

        <!-- Left Column: Instance Info & Client Keys -->
        <div class="lg:col-span-3 p-6 border-r border-zinc-900 flex flex-col space-y-6 text-left">
          <!-- Active Gateway Instance Status Info -->
          <div>
            <h3 class="text-sm font-bold uppercase tracking-wider text-zinc-350 mb-2">Instance Profile</h3>
            <div class="space-y-2 text-xs font-mono">
              <div class="p-2.5 rounded bg-zinc-950/40 border border-zinc-900 space-y-1">
                <div class="flex items-center justify-between">
                  <span class="text-zinc-650 uppercase text-[10px] font-sans font-bold">Status</span>
                  <span class="flex items-center gap-1.5 text-emerald-400 font-bold font-sans text-[10px] uppercase">
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shrink-0"></span>
                    Online
                  </span>
                </div>
                <div class="flex items-center justify-between pt-1">
                  <span class="text-zinc-650 uppercase text-[10px] font-sans font-bold">Listen IP</span>
                  <span class="text-zinc-300">127.0.0.1:8080</span>
                </div>
                <div class="flex items-center justify-between pt-1">
                  <span class="text-zinc-650 uppercase text-[10px] font-sans font-bold">Prefix</span>
                  <span class="text-zinc-300">/v1</span>
                </div>
                <div class="flex items-center justify-between pt-1">
                  <span class="text-zinc-650 uppercase text-[10px] font-sans font-bold">Config Store</span>
                  <span class="text-zinc-300">state.json</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Go Gateway Client API Keys Manager -->
          <div class="border-t border-zinc-900 pt-4 flex flex-col space-y-3 shrink-0">
            <div>
              <h3 class="text-sm font-bold uppercase tracking-wider text-zinc-350 mb-1">Gateway Client Keys</h3>
              <p class="text-xs text-zinc-500 leading-relaxed">Generated tokens authorized to proxy prompts through Cerberus.</p>
            </div>

            <!-- Create new key form -->
            <div class="flex gap-2">
              <input
                type="text"
                v-model="newKeyLabel"
                placeholder="Key label..."
                @keyup.enter="generateNewKey"
                class="flex-1 text-xs bg-zinc-950 border border-zinc-900 text-zinc-300 p-1.5 rounded focus:outline-none font-mono"
              />
              <button
                @click="generateNewKey"
                class="text-xs font-bold text-rose-455 border border-rose-900/40 bg-rose-950/10 hover:bg-rose-900/25 hover:text-rose-200 px-2.5 py-1.5 rounded cursor-pointer transition-all font-push"
              >
                Create
              </button>
            </div>

            <div v-if="keyGenError" class="text-[10px] text-red-400 font-mono leading-relaxed bg-red-950/20 border border-red-900/30 p-2.5 rounded">
              {{ keyGenError }}
            </div>

            <!-- Key List scrollbox -->
            <div class="space-y-1.5 max-h-48 overflow-y-auto pr-1">
              <div
                v-for="key in apiKeys"
                :key="key.id"
                class="flex items-center justify-between p-2 rounded bg-zinc-950/40 border border-zinc-900 text-xs font-mono"
              >
                <div class="text-left truncate pr-2">
                  <span class="text-zinc-300 block font-sans font-bold text-xs leading-normal truncate">{{ key.label }}</span>
                  <span class="text-zinc-600 block text-[10px] mt-0.5 font-mono">ID: {{ key.id }}</span>
                </div>
                <button
                  @click="deleteApiKey(key.id)"
                  class="text-zinc-600 hover:text-red-400 p-1 rounded hover:bg-zinc-900/60 transition-colors cursor-pointer shrink-0"
                  title="Revoke Key"
                >
                  <Trash2 class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Side: Config Forms and Telemetry Details -->
        <div class="lg:col-span-9 flex flex-col p-6 space-y-6">

          <!-- Proxy Settings Form Card -->
          <div class="cyber-card rounded p-5 border border-zinc-900 bg-zinc-900/10 text-left space-y-5">
            <div>
              <h3 class="text-sm font-bold uppercase tracking-wider text-zinc-300 font-push">Gateway Core Configuration</h3>
              <p class="text-xs text-zinc-500 leading-normal mt-0.5">Parameters synced in real-time with the running Go proxy instance</p>
            </div>

            <div class="grid md:grid-cols-2 gap-5">
              <div class="space-y-1.5 md:col-span-2">
                <label class="text-xs font-semibold text-zinc-350 font-push">Proxy Instance Name</label>
                <input
                  type="text"
                  v-model="activeProject.name"
                  placeholder="Enter proxy instance name..."
                  class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono"
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-xs font-semibold text-zinc-350 font-push">Admin Secret Token</label>
                <input
                  type="password"
                  v-model="adminToken"
                  placeholder="Enter CERBERUS_ADMIN_TOKEN..."
                  class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono"
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-xs font-semibold text-zinc-350 font-push">Upstream Destination URL</label>
                <input
                  type="text"
                  v-model="guardsConfig.gatewaySettings.upstream"
                  placeholder="e.g. https://openrouter.ai/api/v1"
                  class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono"
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-xs font-semibold text-zinc-350 font-push">Upstream Authorization Key</label>
                <input
                  type="password"
                  v-model="guardsConfig.gatewaySettings.upstreamKey"
                  placeholder="Keep empty to forward client token unchanged"
                  class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono"
                />
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-1.5">
                  <label class="text-xs font-semibold text-zinc-350 font-push">Max Payload (Bytes)</label>
                  <input
                    type="number"
                    v-model.number="guardsConfig.gatewaySettings.maxBodyBytes"
                    class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono"
                  />
                </div>
                <div class="space-y-1.5">
                  <label class="text-xs font-semibold text-zinc-350 font-push">Outbound Mode</label>
                  <select
                    v-model="guardsConfig.gatewaySettings.outboundMode"
                    class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-push"
                  >
                    <option value="off">Off (Passthrough)</option>
                    <option value="buffer">Buffer (Scan Full)</option>
                    <option value="stream">Stream (Scan SSE)</option>
                  </select>
                </div>
              </div>
            </div>

            <div v-if="syncError" class="text-xs text-red-400 font-mono leading-relaxed bg-red-950/20 border border-red-900/30 p-2.5 rounded">
              {{ syncError }}
            </div>
          </div>

          <!-- Detection Rules Summary Card -->
          <div class="cyber-card rounded p-5 border border-zinc-900 bg-zinc-900/10 text-left space-y-4">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-sm font-bold uppercase tracking-wider text-zinc-300 font-push">Detection Rules</h3>
                <p class="text-xs text-zinc-500 leading-normal mt-0.5">
                  {{ enabledCount }} of {{ detectors.length }} detectors active · toggle them in the Test Zone
                </p>
              </div>
              <router-link
                :to="{ name: 'testzone' }"
                class="text-xs font-bold text-rose-455 border border-rose-900/40 bg-rose-950/10 hover:bg-rose-900/25 hover:text-rose-200 px-3 py-1.5 rounded cursor-pointer transition-all font-push"
              >
                Open Test Zone
              </router-link>
            </div>

            <div class="flex flex-wrap gap-2">
              <span
                v-for="d in detectors"
                :key="d.id"
                class="text-[11px] font-mono px-2.5 py-1 rounded border inline-flex items-center gap-1.5"
                :class="disabledRules.includes(d.id)
                  ? 'border-zinc-850 bg-zinc-950/40 text-zinc-650'
                  : 'border-emerald-900/30 bg-emerald-950/10 text-emerald-400/90'"
              >
                <span class="w-1.5 h-1.5 rounded-full" :class="disabledRules.includes(d.id) ? 'bg-zinc-700' : 'bg-emerald-500'"></span>
                {{ d.name }}
              </span>
              <span v-if="detectors.length === 0" class="text-xs text-zinc-600 font-mono">
                No detectors loaded - check the admin token.
              </span>
            </div>
          </div>

          <!-- Config JSON & Integration snippet Split Grid -->
          <div class="grid md:grid-cols-12 gap-6 items-start">

            <!-- Compiled JSON details -->
            <div class="md:col-span-5 flex flex-col justify-start space-y-3 text-left">
              <div class="text-sm font-semibold uppercase tracking-wider text-zinc-300 font-push">
                Compiled Configuration
              </div>
              <div class="cyber-card rounded p-3.5 font-mono text-xs text-zinc-450 border border-zinc-900 bg-zinc-950 h-56 overflow-y-auto">
                <pre class="leading-relaxed">{{ compiledJson }}</pre>
              </div>

              <div class="space-y-2">
                <button
                  @click="deployGateway"
                  :disabled="isDeploying"
                  class="w-full flex items-center justify-center gap-2 border border-rose-900/60 bg-rose-950/15 text-rose-455 hover:bg-rose-900/30 hover:text-rose-200 font-bold text-xs py-2.5 px-4 rounded transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer font-push"
                >
                  <Play class="w-3.5 h-3.5 fill-current" />
                  <span v-if="!isDeploying">Sync Config to Proxy</span>
                  <span v-else>Deploying cluster...</span>
                </button>
                <div v-if="justDeployed" class="bg-zinc-900 border border-zinc-850 rounded p-2 text-xs text-zinc-500 text-center">
                  Gateway endpoints successfully updated.
                </div>
              </div>
            </div>

            <!-- Integration snippets -->
            <div class="md:col-span-7 flex flex-col justify-start space-y-3 text-left">
              <div class="text-sm font-semibold uppercase tracking-wider text-zinc-300 font-push">
                Integration Hook
              </div>

              <!-- Tabs -->
              <div class="flex border-b border-zinc-900 text-xs font-mono">
                <button
                  v-for="tab in (['js', 'python', 'curl'] as const)"
                  :key="tab"
                  @click="activeTab = tab"
                  class="px-3 py-1.5 border-b font-semibold capitalize cursor-pointer transition-all"
                  :class="activeTab === tab ? 'border-zinc-300 text-white bg-zinc-900/40' : 'border-transparent text-zinc-500 hover:text-white'"
                >
                  {{ tab === 'js' ? 'Javascript' : tab === 'python' ? 'Python' : 'cURL' }}
                </button>
              </div>

              <!-- Snippet Box -->
              <div class="relative h-64 flex flex-col justify-between">
                <button
                  @click="copySnippet"
                  class="absolute top-2.5 right-2.5 text-zinc-500 hover:text-white p-1 rounded hover:bg-zinc-800 transition-colors cursor-pointer"
                  title="Copy snippet"
                >
                  <Check v-if="copied" class="w-3 h-3 text-zinc-300" />
                  <Copy v-else class="w-3 h-3" />
                </button>

                <div class="cyber-card rounded p-3.5 font-mono text-xs text-zinc-450 bg-zinc-950 border border-zinc-800 text-left overflow-x-auto h-full overflow-y-auto">
                  <pre class="leading-relaxed">{{ codeSnippets[activeTab] }}</pre>
                </div>
              </div>
            </div>

          </div>

        </div>
      </div>

    </main>

    <!-- API Key Generated Modal Overlay -->
    <div 
      v-if="showKeyModal" 
      class="fixed inset-0 bg-zinc-950/80 backdrop-blur-sm z-50 flex items-center justify-center p-4"
    >
      <div 
        class="cyber-card rounded p-6 border border-zinc-800 bg-zinc-900 max-w-md w-full space-y-5 text-left shadow-2xl relative"
      >
        <div class="space-y-1">
          <h3 class="text-sm font-bold text-white font-push uppercase tracking-wider">API Key Generated</h3>
          <p class="text-xs text-zinc-500">Copy this secret key. For security, you won't be able to view it again.</p>
        </div>

        <div class="space-y-1.5">
          <label class="text-[10px] font-semibold text-zinc-400 font-push uppercase tracking-wider">Key Label</label>
          <div class="text-xs text-zinc-350 font-mono bg-zinc-950 border border-zinc-950 p-2.5 rounded">
            {{ generatedKeyLabel }}
          </div>
        </div>

        <div class="space-y-1.5">
          <label class="text-[10px] font-semibold text-zinc-400 font-push uppercase tracking-wider">Secret API Key</label>
          <div class="flex items-center justify-between bg-zinc-950 border border-zinc-950 rounded p-2.5 text-xs font-mono text-zinc-300">
            <span class="truncate pr-4 select-all">{{ generatedKey }}</span>
            <button 
              @click="copyGeneratedKey"
              class="text-zinc-500 hover:text-white p-1 rounded hover:bg-zinc-900 transition-all shrink-0 cursor-pointer"
              title="Copy Key"
            >
              <Check v-if="copiedGeneratedKey" class="w-3.5 h-3.5 text-emerald-400" />
              <Copy v-else class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <div class="bg-amber-950/10 border border-amber-900/30 rounded p-3 text-[10px] text-amber-500 font-mono leading-relaxed">
          WARNING: Store it safely! Anyone who has this key can access your reverse proxy backend admin endpoints.
        </div>

        <button 
          @click="showKeyModal = false"
          class="w-full bg-zinc-100 hover:bg-white text-zinc-950 font-bold text-xs py-2.5 rounded transition-all duration-150 cursor-pointer font-push text-center"
        >
          I have saved this key
        </button>
      </div>
    </div>
  </div>
</template>
