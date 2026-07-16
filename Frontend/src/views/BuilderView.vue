<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { 
  Shield, 
  Settings, 
  Play, 
  Copy, 
  Check,
  Cpu,
  User,
  LogOut,
  Sliders,
  GripVertical,
  Plus,
  Trash2
} from '@lucide/vue'
import Slider from 'primevue/slider'
import InputText from 'primevue/inputtext'

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
    id: 'shield_a8b92c',
    name: 'Website Chatbot',
    proxyUrl: 'https://api.cerberus.sh/v1/proxy/shield_a8b92c',
    status: 'active',
    activeGuardsCount: 2
  },
  {
    id: 'shield_df47x1',
    name: 'Slack Support Agent',
    proxyUrl: 'https://api.cerberus.sh/v1/proxy/shield_df47x1',
    status: 'active',
    activeGuardsCount: 3
  },
  {
    id: 'shield_9281ab',
    name: 'Internal Data SQL Bot',
    proxyUrl: 'https://api.cerberus.sh/v1/proxy/shield_9281ab',
    status: 'inactive',
    activeGuardsCount: 1
  }
])

const activeProjectId = ref('shield_a8b92c')
const copiedUrl = ref('')

const copyProjectUrl = (url: string) => {
  navigator.clipboard.writeText(url)
  copiedUrl.value = url
  setTimeout(() => {
    copiedUrl.value = ''
  }, 2000)
}

const createNewProject = () => {
  const names = ['E-Commerce Conversationalist', 'HR Automation Sentry', 'Logistics Routing Node', 'Marketing Copilot Core']
  const randomName = names[Math.floor(Math.random() * names.length)] + ' #' + Math.floor(Math.random() * 900 + 100)
  const id = 'shield_' + Math.random().toString(36).substring(2, 8)
  const newProj = {
    id,
    name: randomName,
    proxyUrl: `https://api.cerberus.sh/v1/proxy/${id}`,
    status: 'active' as const,
    activeGuardsCount: 1
  }
  projects.value.push(newProj)
  selectProject(newProj)
}

const selectProject = (project: Project) => {
  activeProjectId.value = project.id
  shieldId.value = project.id
  
  // Dynamically configure nodes to represent loading project configs
  if (project.name.includes('Chatbot')) {
    nodes.value.promptInjection.active = true
    nodes.value.toolFirewall.active = false
    nodes.value.piiMasker.active = true
    nodes.value.rateLimiter.active = false
  } else if (project.name.includes('Slack') || project.name.includes('Automation')) {
    nodes.value.promptInjection.active = true
    nodes.value.toolFirewall.active = true
    nodes.value.piiMasker.active = false
    nodes.value.rateLimiter.active = true
  } else if (project.name.includes('SQL') || project.name.includes('Routing')) {
    nodes.value.promptInjection.active = true
    nodes.value.toolFirewall.active = true
    nodes.value.piiMasker.active = true
    nodes.value.rateLimiter.active = false
  } else {
    nodes.value.promptInjection.active = true
    nodes.value.toolFirewall.active = false
    nodes.value.piiMasker.active = false
    nodes.value.rateLimiter.active = false
  }
}

// Pipeline Guardrails Configurations
const guardsConfig = ref({
  promptInjection: {
    name: 'Ingress Sentry (Left Head)',
    description: 'Blocks adversarial text payloads',
    similarityThreshold: 75,
    customPatterns: 'ignore previous, act as developer, system override',
    threatAction: 'BLOCK_CONNECTION'
  },
  toolFirewall: {
    name: 'Runtime Firewall (Middle Head)',
    description: 'Blocks hazardous system calls',
    commandBlacklist: 'rm -rf, DROP TABLE, rm -d, mkfs, format',
    alertSeverity: 'CRITICAL'
  },
  piiMasker: {
    name: 'Egress Censor (Right Head) - PII',
    description: 'Redacts outgoing sensitive variables',
    maskString: '[REDACTED_SECURE]',
    maskTypes: ['API Keys', 'Emails', 'Credit Cards'],
    threatAction: 'REDACT'
  },
  rateLimiter: {
    name: 'Egress Censor (Right Head) - Governor',
    description: 'Enforces hard ceilings and token limits',
    maxTokensPerMinute: 15000,
    costCeilingPerDay: 5.0
  }
})

// Dynamic Tag Rules builders
const newToolTag = ref('')
const toolFirewallTags = ref(['rm -rf', 'DROP TABLE', 'rm -d', 'mkfs', 'format'])

const addToolFirewallTag = () => {
  const cleanVal = newToolTag.value.trim().replace(/,/g, '')
  if (cleanVal && !toolFirewallTags.value.includes(cleanVal)) {
    toolFirewallTags.value.push(cleanVal)
  }
  newToolTag.value = ''
  guardsConfig.value.toolFirewall.commandBlacklist = toolFirewallTags.value.join(', ')
}

const removeToolFirewallTag = (idx: number) => {
  toolFirewallTags.value.splice(idx, 1)
  guardsConfig.value.toolFirewall.commandBlacklist = toolFirewallTags.value.join(', ')
}

const handleToolTagBackspace = () => {
  if (newToolTag.value === '' && toolFirewallTags.value.length > 0) {
    toolFirewallTags.value.pop()
    guardsConfig.value.toolFirewall.commandBlacklist = toolFirewallTags.value.join(', ')
  }
}

// Ingress Sentry Blacklist Tag Builders
const newPatternTag = ref('')
const promptInjectionTags = ref(['ignore previous', 'act as developer', 'system override'])

const addPatternTag = () => {
  const cleanVal = newPatternTag.value.trim().replace(/,/g, '')
  if (cleanVal && !promptInjectionTags.value.includes(cleanVal)) {
    promptInjectionTags.value.push(cleanVal)
  }
  newPatternTag.value = ''
  guardsConfig.value.promptInjection.customPatterns = promptInjectionTags.value.join(', ')
}

const removePatternTag = (idx: number) => {
  promptInjectionTags.value.splice(idx, 1)
  guardsConfig.value.promptInjection.customPatterns = promptInjectionTags.value.join(', ')
}

const handlePatternTagBackspace = () => {
  if (newPatternTag.value === '' && promptInjectionTags.value.length > 0) {
    promptInjectionTags.value.pop()
    guardsConfig.value.promptInjection.customPatterns = promptInjectionTags.value.join(', ')
  }
}

// Security Strictness levels linked to math threshold parameters
const strictnessLevel = computed({
  get() {
    const score = guardsConfig.value.promptInjection.similarityThreshold
    if (score <= 60) return 'low'
    if (score <= 85) return 'medium'
    return 'high'
  },
  set(val: 'low' | 'medium' | 'high') {
    if (val === 'low') {
      guardsConfig.value.promptInjection.similarityThreshold = 55
      guardsConfig.value.promptInjection.threatAction = 'ALERT_ONLY'
    } else if (val === 'medium') {
      guardsConfig.value.promptInjection.similarityThreshold = 75
      guardsConfig.value.promptInjection.threatAction = 'BLOCK_CONNECTION'
    } else {
      guardsConfig.value.promptInjection.similarityThreshold = 90
      guardsConfig.value.promptInjection.threatAction = 'BLOCK_CONNECTION'
    }
  }
})

type GuardKey = 'promptInjection' | 'toolFirewall' | 'piiMasker' | 'rateLimiter';
const selectedGuardKey = ref<GuardKey>('promptInjection')

// 2D Node Sandbox Definitions
interface SandboxNode {
  key: string
  name: string
  type: 'input' | 'output' | 'guard'
  description: string
  x: number
  y: number
  active: boolean
}

const nodes = ref<Record<string, SandboxNode>>({
  input: {
    key: 'input',
    name: 'Agent Request Init',
    type: 'input',
    description: 'Entry endpoint for agent prompts',
    x: 30,
    y: 180,
    active: true
  },
  promptInjection: {
    key: 'promptInjection',
    name: 'Ingress Sentry (Left Head)',
    type: 'guard',
    description: 'Blocks adversarial text payloads',
    x: 270,
    y: 50,
    active: true
  },
  toolFirewall: {
    key: 'toolFirewall',
    name: 'Runtime Firewall (Middle Head)',
    type: 'guard',
    description: 'Blocks hazardous system calls',
    x: 270,
    y: 280,
    active: true
  },
  piiMasker: {
    key: 'piiMasker',
    name: 'Egress Censor (Right Head) - PII',
    type: 'guard',
    description: 'Redacts outgoing sensitive variables',
    x: 520,
    y: 50,
    active: false
  },
  rateLimiter: {
    key: 'rateLimiter',
    name: 'Egress Censor (Right Head) - Governor',
    type: 'guard',
    description: 'Enforces token speed thresholds',
    x: 520,
    y: 280,
    active: false
  },
  output: {
    key: 'output',
    name: 'Target AI Service',
    type: 'output',
    description: 'Forwarding endpoint for verified logs',
    x: 770,
    y: 180,
    active: true
  }
})

// Node Drag & Drop Sandbox Engine
const draggingNodeKey = ref<string | null>(null)
const mouseOffset = ref({ x: 0, y: 0 })
const draggedSidebarItem = ref<GuardKey | null>(null)

const startNodeDrag = (e: MouseEvent, key: string) => {
  e.preventDefault()
  draggingNodeKey.value = key
  const node = nodes.value[key]
  mouseOffset.value = {
    x: e.clientX - node.x,
    y: e.clientY - node.y
  }
}

const onCanvasMouseMove = (e: MouseEvent) => {
  if (!draggingNodeKey.value) return
  const key = draggingNodeKey.value
  const node = nodes.value[key]
  
  let newX = e.clientX - mouseOffset.value.x
  let newY = e.clientY - mouseOffset.value.y
  
  // Keep nodes within the sandbox boundaries
  newX = Math.max(10, Math.min(800, newX))
  newY = Math.max(10, Math.min(370, newY))
  
  node.x = newX
  node.y = newY
}

const stopNodeDrag = () => {
  draggingNodeKey.value = null
}

// Sidebar Drag and Drop handlers
const handleSidebarDragStart = (key: GuardKey) => {
  draggedSidebarItem.value = key
}

const handleSidebarDragEnd = () => {
  draggedSidebarItem.value = null
}

const handleCanvasDrop = (e: DragEvent) => {
  if (!draggedSidebarItem.value) return
  
  const key = draggedSidebarItem.value
  const canvas = e.currentTarget as HTMLElement
  const rect = canvas.getBoundingClientRect()
  
  let dropX = e.clientX - rect.left - 100 // Center offset (half of card width 200px)
  let dropY = e.clientY - rect.top - 40  // Center offset (half of card height 80px)
  
  dropX = Math.max(10, Math.min(800, dropX))
  dropY = Math.max(10, Math.min(370, dropY))
  
  nodes.value[key].active = true
  nodes.value[key].x = dropX
  nodes.value[key].y = dropY
  
  selectedGuardKey.value = key
  draggedSidebarItem.value = null
  
  // Sync active project counters
  syncProjectGuardsCount()
}

// Click triggers for accessibility
const activateNode = (key: GuardKey) => {
  nodes.value[key].active = true
  nodes.value[key].x = 400
  nodes.value[key].y = 150
  selectedGuardKey.value = key
  syncProjectGuardsCount()
}

const deactivateNode = (key: GuardKey) => {
  nodes.value[key].active = false
  syncProjectGuardsCount()
}

const syncProjectGuardsCount = () => {
  const current = projects.value.find(p => p.id === activeProjectId.value)
  if (current) {
    current.activeGuardsCount = Object.values(nodes.value).filter(n => n.active && n.type === 'guard').length
  }
}

// Compute connection paths dynamically based on horizontal sorted coordinates
const connectionPaths = computed(() => {
  const activeNodes = Object.values(nodes.value).filter(n => n.active)
  activeNodes.sort((a, b) => a.x - b.x)
  
  const paths: string[] = []
  for (let i = 0; i < activeNodes.length - 1; i++) {
    const nodeA = activeNodes[i]
    const nodeB = activeNodes[i + 1]
    
    // Output center point of Node A (right edge)
    const x1 = nodeA.x + 200
    const y1 = nodeA.y + 40
    
    // Input center point of Node B (left edge)
    const x2 = nodeB.x
    const y2 = nodeB.y + 40
    
    // Bezier control points
    const cp1x = x1 + (x2 - x1) / 2
    const cp1y = y1
    const cp2x = x1 + (x2 - x1) / 2
    const cp2y = y2
    
    paths.push(`M ${x1} ${y1} C ${cp1x} ${cp1y}, ${cp2x} ${cp2y}, ${x2} ${y2}`)
  }
  return paths
})

// Computed Configuration JSON (ordered dynamically by X coordinate)
const compiledJson = computed(() => {
  const activeGuards: Record<string, any> = {}
  
  const sortedActiveGuards = Object.values(nodes.value)
    .filter(n => n.active && n.type === 'guard')
    .sort((a, b) => a.x - b.x)
    
  sortedActiveGuards.forEach(node => {
    const key = node.key
    if (key === 'promptInjection') {
      activeGuards.prompt_injection_detector = {
        action: guardsConfig.value.promptInjection.threatAction,
        strictness_profile: strictnessLevel.value.toUpperCase(),
        sensitivity_threshold: guardsConfig.value.promptInjection.similarityThreshold / 100,
        blacklist_phrases: promptInjectionTags.value
      }
    }
    if (key === 'toolFirewall') {
      activeGuards.tool_execution_firewall = {
        severity: guardsConfig.value.toolFirewall.alertSeverity,
        blocked_signatures: toolFirewallTags.value
      }
    }
    if (key === 'piiMasker') {
      activeGuards.pii_redactor = {
        action: guardsConfig.value.piiMasker.threatAction,
        redaction_token: guardsConfig.value.piiMasker.maskString,
        entities: guardsConfig.value.piiMasker.maskTypes
      }
    }
    if (key === 'rateLimiter') {
      activeGuards.budget_governor = {
        max_tokens_tpm: guardsConfig.value.rateLimiter.maxTokensPerMinute,
        max_cost_usd_daily: guardsConfig.value.rateLimiter.costCeilingPerDay
      }
    }
  })

  return JSON.stringify({
    shield_id: shieldId.value,
    version: '1.0.0',
    meta: {
      client: 'cerberus-dashboard',
      updated_at: new Date().toISOString()
    },
    pipeline: activeGuards
  }, null, 2)
})

// Deploy simulation
const deployGateway = () => {
  isDeploying.value = true
  setTimeout(() => {
    isDeploying.value = false
    justDeployed.value = true
    
    // Toggle active project status to shielding
    const current = projects.value.find(p => p.id === activeProjectId.value)
    if (current) {
      current.status = 'active'
    }
    
    setTimeout(() => {
      justDeployed.value = false
    }, 3000)
  }, 1500)
}

// Proxy Endpoint Links
const proxyEndpoint = computed(() => `https://api.cerberus.sh/v1/proxy/${shieldId.value}`)

// Snippets code
const codeSnippets = computed(() => {
  return {
    js: `// Initialize Cerberus Secure Pass-through Proxy
const response = await fetch("${proxyEndpoint.value}/v1/chat/completions", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "Authorization": "Bearer YOUR_API_KEY",
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
    api_key="YOUR_API_KEY"
)

completion = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Execute agent tools..."}]
)
print(completion.choices[0].message.content)`,
    curl: `curl ${proxyEndpoint.value}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer $OPENAI_API_KEY" \\
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
          <div class="w-8 h-8 rounded bg-zinc-900 border border-zinc-800 flex items-center justify-center">
            <Shield class="w-4 h-4 text-zinc-300" />
          </div>
          <span class="font-bold tracking-wider text-sm text-white font-push">CERBERUS</span>
        </div>

        <!-- Navigation Links -->
        <nav class="p-4 space-y-1">
          <router-link 
            :to="{ name: 'builder' }" 
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold bg-zinc-900 text-white border border-zinc-800"
          >
            <Sliders class="w-3.5 h-3.5" />
            Pipeline Builder
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
            <p class="text-[9px] text-zinc-550">Verified Session</p>
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
          <h2 class="text-sm font-bold text-white flex items-center gap-2 font-push">
            Control Plane
            <span class="text-[9px] bg-zinc-900 border border-zinc-800 text-zinc-400 px-1.5 py-0.5 rounded font-mono">v1.0.0</span>
          </h2>
          <p class="text-[10px] text-zinc-500">Secure agent endpoints via multi-project proxy matrices</p>
        </div>

        <div class="flex items-center gap-4">
          <div class="flex items-center gap-2 text-[10px]">
            <span class="w-2 h-2 rounded-full bg-zinc-400"></span>
            <span class="text-zinc-550 font-medium">Gateway Cluster: </span>
            <span class="font-bold text-zinc-300">Active</span>
          </div>
        </div>
      </header>

      <!-- 2. Project Matrix Panel Grid (Multi-Project Support) -->
      <div class="px-8 pt-6 pb-2 text-left shrink-0">
        <div class="flex items-center justify-between mb-3.5">
          <h3 class="text-[10px] font-bold uppercase tracking-wider text-zinc-400 font-push">Active Shields Matrix</h3>
          <button 
            @click="createNewProject"
            class="text-[9px] font-bold text-zinc-350 hover:text-white border border-zinc-850 bg-zinc-900/40 px-2.5 py-1 rounded cursor-pointer transition-all hover:border-zinc-700"
          >
            + New Project Shield
          </button>
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
                {{ project.status === 'active' ? '🟢 Shielding' : '🔴 Blocked' }}
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
      <div class="flex-1 grid lg:grid-cols-12 min-h-0">
        
        <!-- Left Column: Available Shelf (Node Library Shelf) -->
        <div class="lg:col-span-3 p-6 border-r border-zinc-900 flex flex-col space-y-4 text-left">
          <div>
            <h3 class="text-xs font-bold uppercase tracking-wider text-zinc-450 mb-1">Node Shelf</h3>
            <p class="text-[9px] text-zinc-550 leading-relaxed">Drag nodes to the canvas zone below or click add buttons to hook them up.</p>
          </div>

          <div class="space-y-2 flex-1">
            <div 
              v-for="(node, key) in nodes" 
              v-show="node.type === 'guard'"
              :key="key"
              :draggable="!node.active"
              @dragstart="handleSidebarDragStart(key as GuardKey)"
              @dragend="handleSidebarDragEnd"
              @click="selectedGuardKey = key as GuardKey"
              class="cyber-card rounded p-3 border text-left transition-all hover:bg-zinc-900/30"
              :class="[
                node.active ? 'border-zinc-900 bg-zinc-900/10 opacity-50 cursor-default' : 'border-zinc-900 bg-zinc-950/40 cursor-grab active:cursor-grabbing',
                selectedGuardKey === key ? 'ring-1 ring-zinc-700' : ''
              ]"
            >
              <div class="flex items-start justify-between">
                <div class="flex items-start gap-2.5">
                  <div class="mt-0.5 w-5.5 h-5.5 rounded flex items-center justify-center bg-zinc-900 border border-zinc-855 text-zinc-600 shrink-0">
                    <GripVertical class="w-3 h-3 cursor-grab" v-if="!node.active" />
                    <Check class="w-3 h-3 text-zinc-450" v-else />
                  </div>
                  <div>
                    <h4 class="text-[10px] font-bold text-zinc-300 font-push leading-tight">{{ node.name }}</h4>
                    <p class="text-[8px] text-zinc-500 leading-normal mt-0.5">{{ node.description }}</p>
                  </div>
                </div>
                
                <button 
                  v-if="!node.active"
                  @click.stop="activateNode(key as GuardKey)"
                  class="p-0.5 rounded border border-zinc-800 bg-zinc-900 text-zinc-400 hover:text-white hover:bg-zinc-800 cursor-pointer shrink-0"
                  title="Add to Workspace"
                >
                  <Plus class="w-3 h-3" />
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Right Side: Infinite Sandbox Canvas and Parameter Panel -->
        <div class="lg:col-span-9 flex flex-col min-h-0">
          
          <!-- Infinite Sandbox Canvas (400px Height) -->
          <div class="p-6 border-b border-zinc-900 shrink-0">
            <div 
              @dragover.prevent
              @drop="handleCanvasDrop"
              @mousemove="onCanvasMouseMove"
              @mouseup="stopNodeDrag"
              @mouseleave="stopNodeDrag"
              class="relative w-full h-[380px] border border-zinc-900 bg-zinc-950/40 rounded-lg overflow-hidden select-none"
            >
              <!-- Grid background lines -->
              <div class="absolute inset-0 minimal-dashed opacity-25"></div>

              <!-- Connections SVG Overlay Layer -->
              <svg class="absolute inset-0 pointer-events-none w-full h-full">
                <path 
                  v-for="(path, idx) in connectionPaths" 
                  :key="idx" 
                  :d="path" 
                  stroke="#3f3f46" 
                  stroke-width="1.5" 
                  fill="none" 
                  stroke-dasharray="3,3"
                />
              </svg>

              <!-- Node Cards list -->
              <div 
                v-for="node in Object.values(nodes).filter(n => n.active)" 
                :key="node.key"
                @mousedown="selectedGuardKey = node.key as GuardKey"
                class="absolute w-52 h-[76px] rounded bg-zinc-900 border border-zinc-800 flex flex-col justify-between text-left p-2.5 shadow-lg select-none"
                :class="[
                  selectedGuardKey === node.key ? 'border-zinc-500 bg-zinc-900/90' : 'border-zinc-850 bg-zinc-900/60',
                  node.type === 'guard' ? 'hover:border-zinc-600' : 'bg-zinc-950/80 border-dashed border-zinc-800'
                ]"
                :style="{ left: node.x + 'px', top: node.y + 'px' }"
              >
                <!-- Left Connection Input Port -->
                <div v-if="node.type !== 'input'" class="absolute -left-1.5 top-1/2 -translate-y-1/2 w-3 h-3 rounded-full bg-zinc-800 border border-zinc-955 flex items-center justify-center">
                  <div class="w-1.5 h-1.5 rounded-full bg-zinc-500"></div>
                </div>

                <!-- Right Connection Output Port -->
                <div v-if="node.type !== 'output'" class="absolute -right-1.5 top-1/2 -translate-y-1/2 w-3 h-3 rounded-full bg-zinc-800 border border-zinc-955 flex items-center justify-center">
                  <div class="w-1.5 h-1.5 rounded-full bg-zinc-500"></div>
                </div>

                <!-- Header drag handle -->
                <div 
                  @mousedown="startNodeDrag($event, node.key)"
                  class="flex items-center gap-1.5 pb-1 border-b border-zinc-850/60 cursor-grab active:cursor-grabbing text-zinc-450 hover:text-white"
                >
                  <GripVertical class="w-3 h-3 text-zinc-600 shrink-0 animate-pulse" />
                  <span class="text-[10px] font-bold truncate leading-none font-push">{{ node.name }}</span>
                  
                  <!-- Trash button on guard nodes -->
                  <button 
                    v-if="node.type === 'guard'"
                    @click.stop="deactivateNode(node.key as GuardKey)"
                    class="ml-auto p-0.5 text-zinc-650 hover:text-white cursor-pointer shrink-0"
                    title="Remove Node"
                  >
                    <Trash2 class="w-3 h-3" />
                  </button>
                </div>

                <!-- Node Metadata Description -->
                <div class="text-[8px] text-zinc-500 leading-normal line-clamp-2 select-none">
                  {{ node.description }}
                </div>
              </div>
            </div>
          </div>

          <!-- Bottom Panel: Config details and compiled JSON -->
          <div class="flex-1 grid md:grid-cols-12 p-6 gap-6 min-h-0 bg-zinc-900/10 border-t border-zinc-900 overflow-y-auto">
            
            <!-- Parameters configuration detail -->
            <div class="md:col-span-5 flex flex-col justify-start space-y-4">
              <div class="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-zinc-350 font-push text-left">
                <Settings class="w-3.5 h-3.5 text-zinc-450" />
                Parameters: {{ guardsConfig[selectedGuardKey].name }}
              </div>

              <!-- 1. The Policy Visualizer Interface -->
              <div class="cyber-card rounded p-4 border border-zinc-850 bg-zinc-900/10 text-left space-y-4">
                
                <!-- Active/Enable switch control -->
                <div class="flex items-center justify-between pb-3 border-b border-zinc-850/50">
                  <span class="text-[10px] font-bold text-zinc-300 font-push">Active Pipeline State</span>
                  <button 
                    @click="nodes[selectedGuardKey].active = !nodes[selectedGuardKey].active; syncProjectGuardsCount()"
                    class="w-8 h-4.5 rounded-full p-0.5 transition-colors duration-200 cursor-pointer focus:outline-none shrink-0"
                    :class="nodes[selectedGuardKey].active ? 'bg-zinc-200' : 'bg-zinc-800'"
                  >
                    <div 
                      class="w-3.5 h-3.5 rounded-full bg-zinc-950 transition-transform duration-200"
                      :class="nodes[selectedGuardKey].active ? 'translate-x-3.5' : 'translate-x-0'"
                    ></div>
                  </button>
                </div>

                <!-- A. Ingress Sentry (Prompt Injection) Visual settings -->
                <div v-if="selectedGuardKey === 'promptInjection'" class="space-y-4">
                  <!-- Strictness Level buttons toggle -->
                  <div class="space-y-1.5">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">Security Strictness Level</label>
                    <div class="flex bg-zinc-950 border border-zinc-850 rounded p-0.5 text-[9px] font-mono">
                      <button 
                        @click="strictnessLevel = 'low'"
                        class="flex-1 py-1 text-center font-bold rounded cursor-pointer transition-all uppercase"
                        :class="strictnessLevel === 'low' ? 'bg-zinc-900 text-white' : 'text-zinc-500 hover:text-zinc-300'"
                      >
                        Low
                      </button>
                      <button 
                        @click="strictnessLevel = 'medium'"
                        class="flex-1 py-1 text-center font-bold rounded cursor-pointer transition-all uppercase"
                        :class="strictnessLevel === 'medium' ? 'bg-zinc-900 text-white' : 'text-zinc-500 hover:text-zinc-300'"
                      >
                        Medium
                      </button>
                      <button 
                        @click="strictnessLevel = 'high'"
                        class="flex-1 py-1 text-center font-bold rounded cursor-pointer transition-all uppercase"
                        :class="strictnessLevel === 'high' ? 'bg-zinc-900 text-white' : 'text-zinc-500 hover:text-zinc-300'"
                      >
                        High
                      </button>
                    </div>
                  </div>

                  <!-- Sensitivity Slider -->
                  <div class="space-y-1.5">
                    <div class="flex justify-between text-[10px] font-semibold text-zinc-350">
                      <span>Vector Match Sensitivity</span>
                      <span class="text-zinc-450 font-mono">{{ guardsConfig.promptInjection.similarityThreshold }}%</span>
                    </div>
                    <div class="pt-1">
                      <Slider v-model="guardsConfig.promptInjection.similarityThreshold" :min="30" :max="99" class="w-full" />
                    </div>
                  </div>

                  <!-- Blacklist custom term tag builder -->
                  <div class="space-y-1.5">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">Custom Blacklist Terms</label>
                    <div class="flex flex-wrap gap-1.5 p-2 bg-zinc-950 border border-zinc-900 rounded min-h-[38px] items-center">
                      <span 
                        v-for="(tag, idx) in promptInjectionTags" 
                        :key="idx" 
                        class="inline-flex items-center gap-1 bg-zinc-900 text-zinc-300 px-2 py-0.5 rounded text-[9px] border border-zinc-800 font-mono"
                      >
                        {{ tag }}
                        <button 
                          @click="removePatternTag(idx)" 
                          type="button" 
                          class="text-zinc-500 hover:text-zinc-300 font-bold cursor-pointer text-[8px]"
                        >
                          ×
                        </button>
                      </span>
                      <input 
                        v-model="newPatternTag" 
                        @keydown.enter.prevent="addPatternTag"
                        @keydown.backspace="handlePatternTagBackspace"
                        placeholder="Add term + Enter..."
                        class="flex-1 bg-transparent border-none text-[10px] text-zinc-200 focus:outline-none focus:ring-0 min-w-[100px] placeholder:text-zinc-700"
                      />
                    </div>
                  </div>

                  <div class="space-y-1">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">Threat Action Mode</label>
                    <select v-model="guardsConfig.promptInjection.threatAction" class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-200 p-2 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none">
                      <option value="BLOCK_CONNECTION">BLOCK_CONNECTION</option>
                      <option value="SIMULATE_SAFE">SIMULATE_SAFE</option>
                      <option value="ALERT_ONLY">ALERT_ONLY</option>
                    </select>
                  </div>
                </div>

                <!-- B. Runtime Firewall (Tool call blacklist tags) -->
                <div v-else-if="selectedGuardKey === 'toolFirewall'" class="space-y-4">
                  <div class="space-y-1.5">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">System Call Blacklist</label>
                    <div class="flex flex-wrap gap-1.5 p-2 bg-zinc-950 border border-zinc-900 rounded min-h-[38px] items-center">
                      <span 
                        v-for="(tag, idx) in toolFirewallTags" 
                        :key="idx" 
                        class="inline-flex items-center gap-1 bg-zinc-900 text-zinc-300 px-2 py-0.5 rounded text-[9px] border border-zinc-800 font-mono"
                      >
                        {{ tag }}
                        <button 
                          @click="removeToolFirewallTag(idx)" 
                          type="button" 
                          class="text-zinc-500 hover:text-zinc-300 font-bold cursor-pointer text-[8px]"
                        >
                          ×
                        </button>
                      </span>
                      <input 
                        v-model="newToolTag" 
                        @keydown.enter.prevent="addToolFirewallTag"
                        @keydown.backspace="handleToolTagBackspace"
                        placeholder="Add command + Enter..."
                        class="flex-1 bg-transparent border-none text-[10px] text-zinc-200 focus:outline-none focus:ring-0 min-w-[100px] placeholder:text-zinc-700"
                      />
                    </div>
                  </div>

                  <div class="space-y-1">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">Incident Severity</label>
                    <select v-model="guardsConfig.toolFirewall.alertSeverity" class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-200 p-2 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none">
                      <option value="CRITICAL">CRITICAL</option>
                      <option value="WARNING">WARNING</option>
                    </select>
                  </div>
                </div>

                <!-- C. Egress Censor PII -->
                <div v-else-if="selectedGuardKey === 'piiMasker'" class="space-y-4">
                  <div class="space-y-1">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">Redaction Replacement String</label>
                    <InputText v-model="guardsConfig.piiMasker.maskString" class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-150 p-2 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono" />
                  </div>

                  <div class="space-y-1.5">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">Target Masking Entities</label>
                    <div class="grid grid-cols-2 gap-2 mt-1">
                      <div v-for="type in ['API Keys', 'Emails', 'Credit Cards', 'Phone Numbers', 'IP Addresses', 'System Paths']" :key="type" class="flex items-center gap-2">
                        <input 
                          type="checkbox" 
                          :value="type" 
                          v-model="guardsConfig.piiMasker.maskTypes" 
                          class="w-3.5 h-3.5 border border-zinc-800 rounded bg-zinc-950 text-zinc-200 focus:ring-0 focus:outline-none shrink-0"
                        />
                        <span class="text-[9px] text-zinc-450 font-medium select-none">{{ type }}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- D. Egress Censor Rate Limiter -->
                <div v-else-if="selectedGuardKey === 'rateLimiter'" class="space-y-4">
                  <div class="space-y-1">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">Ceiling Tokens Per Minute (TPM)</label>
                    <input type="number" v-model.number="guardsConfig.rateLimiter.maxTokensPerMinute" class="w-full text-[10px] bg-zinc-950 border border-zinc-900 text-zinc-150 p-2 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none" />
                  </div>

                  <div class="space-y-1">
                    <label class="text-[10px] font-semibold text-zinc-350 font-push">Daily Budget Hard Limit (USD)</label>
                    <div class="flex items-center gap-2 bg-zinc-950 border border-zinc-900 rounded p-1">
                      <span class="text-xs text-zinc-600 pl-1">$</span>
                      <input type="number" step="0.5" v-model.number="guardsConfig.rateLimiter.costCeilingPerDay" class="w-full text-[10px] bg-transparent border-none text-zinc-150 p-1 focus:outline-none" />
                    </div>
                  </div>
                </div>

              </div>
            </div>

            <!-- Compiled JSON details -->
            <div class="md:col-span-3 flex flex-col justify-between space-y-3 text-left">
              <div class="text-xs font-semibold uppercase tracking-wider text-zinc-355 font-push">
                Compiled Gateway JSON
              </div>
              <div class="cyber-card rounded p-3 font-mono text-[9px] text-zinc-400 border border-zinc-900 bg-zinc-950 flex-1 overflow-y-auto max-h-56">
                <pre class="leading-relaxed">{{ compiledJson }}</pre>
              </div>

              <div class="space-y-2">
                <button 
                  @click="deployGateway" 
                  :disabled="isDeploying"
                  class="w-full flex items-center justify-center gap-2 bg-zinc-100 hover:bg-white text-zinc-950 font-bold text-[10px] py-2.5 px-4 rounded transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer font-push"
                >
                  <Play class="w-3.5 h-3.5 fill-current" />
                  <span v-if="!isDeploying">Sync Gateway Endpoints</span>
                  <span v-else>Deploying cluster...</span>
                </button>
                <div v-if="justDeployed" class="bg-zinc-900 border border-zinc-850 rounded p-2 text-[9px] text-zinc-450 text-center">
                  Gateway endpoints successfully updated.
                </div>
              </div>
            </div>

            <!-- Integration snippets -->
            <div class="md:col-span-4 flex flex-col justify-start space-y-2.5 text-left">
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
              <div class="relative flex-1 min-h-[140px] flex flex-col justify-between">
                <button 
                  @click="copySnippet"
                  class="absolute top-2 right-2 text-zinc-500 hover:text-white p-1 rounded hover:bg-zinc-800 transition-colors cursor-pointer"
                  title="Copy snippet"
                >
                  <Check v-if="copied" class="w-3 h-3 text-zinc-300" />
                  <Copy v-else class="w-3 h-3" />
                </button>
                
                <div class="cyber-card rounded p-3 font-mono text-[9px] text-zinc-450 bg-zinc-950/80 border border-zinc-800 text-left overflow-x-auto flex-1 max-h-48">
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

<style scoped>
/* Minimal override for PrimeVue Slider */
:deep(.p-slider) {
  background: #27272a !important; /* zinc-800 */
  height: 3px !important;
}
:deep(.p-slider-range) {
  background: #71717a !important; /* zinc-500 */
}
:deep(.p-slider-handle) {
  background: #f4f4f5 !important; /* zinc-100 */
  border: 1px solid #71717a !important;
  width: 10px !important;
  height: 10px !important;
  margin-top: -3.5px !important;
  margin-left: -5px !important;
  box-shadow: none !important;
}
</style>
