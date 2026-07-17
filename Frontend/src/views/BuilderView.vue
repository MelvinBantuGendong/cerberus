<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Play, 
  Copy, 
  Check,
  Cpu,
  User,
  LogOut,
  Sliders,
  Trash2
} from '@lucide/vue'

const router = useRouter()

// Authentication Logout
const handleLogout = () => {
  localStorage.removeItem('cerberus_auth')
  router.push({ name: 'login' })
}

// Config States
const shieldId = ref('shield_a8b92c')
const isDeploying = ref(false)
const justDeployed = ref(false)
const copied = ref(false)
const activeTab = ref<'js' | 'python' | 'curl'>('js')

// Multi-Project Matrix Support
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
const copiedUrl = ref('')

const copyProjectUrl = (url: string) => {
  navigator.clipboard.writeText(url)
  copiedUrl.value = url
  setTimeout(() => {
    copiedUrl.value = ''
  }, 2000)
}



const selectProject = (project: Project) => {
  activeProjectId.value = project.id
  shieldId.value = project.id
}

// Pipeline Guardrails Configurations
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

// Real API Integration with Go Backend Proxy
const adminToken = ref(localStorage.getItem('cerberus_admin_token') || 'test_token')
const syncError = ref('')
const apiKeys = ref<any[]>([])
const newKeyLabel = ref('')

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
      alert(`Key generated successfully:\n\n${data.key}\n\nMake sure to copy it now. It will not be shown again!`)
      newKeyLabel.value = ''
      await fetchApiKeys()
    } else {
      const txt = await res.text()
      alert(`Failed to generate key: ${txt}`)
    }
  } catch (err: any) {
    alert(`Failed to generate key: ${err.message}`)
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
    
    // Toggle active project status to shielding
    const current = projects.value.find(p => p.id === activeProjectId.value)
    if (current) {
      current.status = 'active'
    }
    
    setTimeout(() => {
      justDeployed.value = false
    }, 3000)
    await fetchBackendConfig() // Refresh settings from server
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
})

watch(adminToken, (newVal) => {
  localStorage.setItem('cerberus_admin_token', newVal)
  fetchBackendConfig()
  fetchApiKeys()
})

// Proxy Endpoint Links
const proxyEndpoint = computed(() => window.location.origin)

// Snippets code
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
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold bg-zinc-900 text-white border border-zinc-800"
          >
            <Sliders class="w-3.5 h-3.5" />
            Manage Proxy
          </router-link>
          
          <router-link 
            :to="{ name: 'analytics' }" 
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold text-zinc-400 hover:bg-zinc-900/60 hover:text-white transition-colors"
          >
            <Cpu class="w-3.5 h-3.5" />
            Threat Intel
            <span class="ml-auto text-[9px] bg-zinc-800 text-zinc-300 px-1.5 py-0.5 rounded">Live</span>
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
      
      <!-- Top Header Bar -->
      <header class="h-16 border-b border-zinc-900 bg-zinc-950/20 backdrop-blur-md flex items-center justify-between px-8 shrink-0">
        <div>
          <h2 class="text-sm font-bold text-white font-push">
            Control Plane
          </h2>
          <p class="text-[10px] text-zinc-500">Secure agent endpoints via multi-project proxy matrices</p>
        </div>

        <div class="flex items-center gap-4 font-push"></div>
      </header>

      <div class="px-8 pt-6 pb-2 text-left shrink-0">
        <div class="flex items-center justify-between mb-3.5">
          <h3 class="text-[10px] font-bold uppercase tracking-wider text-zinc-400 font-push">Active Proxy</h3>

        </div>
        
        <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
          <div 
            v-for="project in projects" 
            :key="project.id"
            @click="selectProject(project)"
            class="cyber-card rounded p-3 border text-left cursor-pointer transition-all hover:bg-zinc-900/10 flex flex-col justify-between space-y-2.5"
            :class="[
              activeProjectId === project.id ? 'border-zinc-600 bg-zinc-900/10' : 'border-zinc-900 bg-zinc-950/20'
            ]"
          >
            <div class="flex items-center justify-between">
              <h4 class="text-[11px] font-bold text-white font-push truncate pr-2">{{ project.name }}</h4>
              <span class="text-[8px] font-bold px-1.5 py-0.5 rounded border font-sans uppercase shrink-0" :class="[
                project.status === 'active' ? 'bg-emerald-950/20 text-emerald-400 border-emerald-900/30' : 'bg-red-950/20 text-red-400 border-red-900/30'
              ]">
                {{ project.status === 'active' ? '🟢 Active' : '🔴 Inactive' }}
              </span>
            </div>
            
            <div class="space-y-1">
              <span class="text-[8px] text-zinc-650 uppercase font-mono tracking-wider font-semibold">Custom Proxy URL</span>
              <div class="flex items-center justify-between bg-zinc-950 border border-zinc-900 rounded p-1.5 text-[9px] font-mono text-zinc-400">
                <span class="truncate pr-1.5">{{ project.proxyUrl }}</span>
                <button 
                  @click.stop="copyProjectUrl(project.proxyUrl)"
                  class="text-zinc-600 hover:text-white p-0.5 rounded transition-all shrink-0 cursor-pointer"
                  title="Copy URL"
                >
                  <Check v-if="copiedUrl === project.proxyUrl" class="w-3 h-3 text-zinc-300" />
                  <Copy v-else class="w-3 h-3" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Layout Panel Grid -->
      <div class="grid lg:grid-cols-12">
        
        <!-- Left Column: Instance Info & Client Keys -->
        <div class="lg:col-span-3 p-6 border-r border-zinc-900 flex flex-col space-y-6 text-left">
          <!-- Active Gateway Instance Status Info -->
          <div>
            <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-450 mb-2">Instance Profile</h3>
            <div class="space-y-2 text-[10px] font-mono">
              <div class="p-2.5 rounded bg-zinc-950/40 border border-zinc-900 space-y-1">
                <div class="flex items-center justify-between">
                  <span class="text-zinc-600 uppercase text-[8px] font-sans font-bold">Status</span>
                  <span class="flex items-center gap-1.5 text-emerald-450 font-bold font-sans text-[8px] uppercase">
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse shrink-0"></span>
                    Online
                  </span>
                </div>
                <div class="flex items-center justify-between pt-1">
                  <span class="text-zinc-600 uppercase text-[8px] font-sans font-bold">Listen IP</span>
                  <span class="text-zinc-300">127.0.0.1:8080</span>
                </div>
                <div class="flex items-center justify-between pt-1">
                  <span class="text-zinc-600 uppercase text-[8px] font-sans font-bold">Prefix</span>
                  <span class="text-zinc-300">/v1</span>
                </div>
                <div class="flex items-center justify-between pt-1">
                  <span class="text-zinc-600 uppercase text-[8px] font-sans font-bold">Config Store</span>
                  <span class="text-zinc-300">state.json</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Go Gateway Client API Keys Manager -->
          <div class="border-t border-zinc-900 pt-4 flex flex-col space-y-3 shrink-0">
            <div>
              <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-450 mb-1">Gateway Client Keys</h3>
              <p class="text-[9px] text-zinc-600 leading-relaxed">Generated tokens authorized to proxy prompts through Cerberus.</p>
            </div>

            <!-- Create new key form -->
            <div class="flex gap-2">
              <input 
                type="text" 
                v-model="newKeyLabel" 
                placeholder="Key label..."
                @keyup.enter="generateNewKey"
                class="flex-1 text-[9px] bg-zinc-950 border border-zinc-900 text-zinc-300 p-1.5 rounded focus:outline-none font-mono"
              />
              <button 
                @click="generateNewKey"
                class="text-[9px] font-bold text-zinc-950 bg-zinc-100 hover:bg-white px-2.5 py-1.5 rounded cursor-pointer transition-all"
              >
                Create
              </button>
            </div>

            <!-- Key List scrollbox -->
            <div class="space-y-1.5 max-h-48 overflow-y-auto pr-1">
              <div 
                v-for="key in apiKeys" 
                :key="key.id"
                class="flex items-center justify-between p-2 rounded bg-zinc-950/40 border border-zinc-900 text-[9px] font-mono"
              >
                <div class="text-left truncate pr-2">
                  <span class="text-zinc-300 block font-sans font-bold leading-normal truncate">{{ key.label }}</span>
                  <span class="text-zinc-600 block text-[8px] mt-0.5 font-mono">ID: {{ key.id }}</span>
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
              <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-350 font-push">Gateway Core Configuration</h3>
              <p class="text-[9.5px] text-zinc-500 leading-normal mt-0.5">Parameters synced in real-time with the running Go proxy instance</p>
            </div>

            <div class="grid md:grid-cols-2 gap-5">
              <div class="space-y-1.5 md:col-span-2">
                <label class="text-[10px] font-semibold text-zinc-350 font-push">Proxy Instance Name</label>
                <input 
                  type="text" 
                  v-model="activeProject.name" 
                  placeholder="Enter proxy instance name..."
                  class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono" 
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-[10px] font-semibold text-zinc-350 font-push">Admin Secret Token</label>
                <input 
                  type="password" 
                  v-model="adminToken" 
                  placeholder="Enter CERBERUS_ADMIN_TOKEN..."
                  class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono" 
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-[10px] font-semibold text-zinc-350 font-push">Upstream Destination URL</label>
                <input 
                  type="text" 
                  v-model="guardsConfig.gatewaySettings.upstream" 
                  placeholder="e.g. https://openrouter.ai/api/v1"
                  class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono" 
                />
              </div>

              <div class="space-y-1.5">
                <label class="text-[10px] font-semibold text-zinc-350 font-push">Upstream Authorization Key</label>
                <input 
                  type="password" 
                  v-model="guardsConfig.gatewaySettings.upstreamKey" 
                  placeholder="Keep empty to forward client token unchanged"
                  class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono" 
                />
              </div>

              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-1.5">
                  <label class="text-[10px] font-semibold text-zinc-350 font-push">Max Payload (Bytes)</label>
                  <input 
                    type="number" 
                    v-model.number="guardsConfig.gatewaySettings.maxBodyBytes" 
                    class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono" 
                  />
                </div>
                <div class="space-y-1.5">
                  <label class="text-[10px] font-semibold text-zinc-350 font-push">Outbound Mode</label>
                  <select 
                    v-model="guardsConfig.gatewaySettings.outboundMode"
                    class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-150 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-push"
                  >
                    <option value="off">Off (Passthrough)</option>
                    <option value="buffer">Buffer (Scan Full)</option>
                    <option value="stream">Stream (Scan SSE)</option>
                  </select>
                </div>
              </div>
            </div>

            <div v-if="syncError" class="text-[9.5px] text-red-400 font-mono leading-relaxed bg-red-950/20 border border-red-900/30 p-2.5 rounded">
              {{ syncError }}
            </div>
          </div>

          <!-- Config JSON & Integration snippet Split Grid -->
          <div class="grid md:grid-cols-12 gap-6 items-start">
            
            <!-- Compiled JSON details -->
            <div class="md:col-span-5 flex flex-col justify-start space-y-3 text-left">
              <div class="text-xs font-semibold uppercase tracking-wider text-zinc-350 font-push">
                Compiled Configuration
              </div>
              <div class="cyber-card rounded p-3.5 font-mono text-[9px] text-zinc-400 border border-zinc-900 bg-zinc-950 h-56 overflow-y-auto">
                <pre class="leading-relaxed">{{ compiledJson }}</pre>
              </div>

              <div class="space-y-2">
                <button 
                  @click="deployGateway" 
                  :disabled="isDeploying"
                  class="w-full flex items-center justify-center gap-2 bg-zinc-100 hover:bg-white text-zinc-950 font-bold text-[10px] py-2.5 px-4 rounded transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer font-push"
                >
                  <Play class="w-3.5 h-3.5 fill-current" />
                  <span v-if="!isDeploying">Sync Config to Proxy</span>
                  <span v-else>Deploying cluster...</span>
                </button>
                <div v-if="justDeployed" class="bg-zinc-900 border border-zinc-850 rounded p-2 text-[9px] text-zinc-450 text-center">
                  Gateway endpoints successfully updated.
                </div>
              </div>
            </div>

            <!-- Integration snippets -->
            <div class="md:col-span-7 flex flex-col justify-start space-y-3 text-left">
              <div class="text-xs font-semibold uppercase tracking-wider text-zinc-350 font-push">
                Integration Hook
              </div>
              
              <!-- Tabs -->
              <div class="flex border-b border-zinc-900 text-[10px] font-mono">
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
                
                <div class="cyber-card rounded p-3.5 font-mono text-[9px] text-zinc-450 bg-zinc-950 border border-zinc-800 text-left overflow-x-auto h-full overflow-y-auto">
                  <pre class="leading-relaxed">{{ codeSnippets[activeTab] }}</pre>
                </div>
              </div>
            </div>
            
          </div>

        </div>
      </div>

    </main>
  </div>
</template>


