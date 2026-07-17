<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  Shield,
  Terminal,
  Cpu,
  User,
  LogOut,
  Sliders,
  TrendingUp,
  ShieldAlert,
  Activity,
  Zap,
  Trash2,
  EyeOff,
  Flame,
  DollarSign,
  Bug
} from '@lucide/vue'

const router = useRouter()

const handleLogout = () => {
  localStorage.removeItem('cerberus_auth')
  router.push({ name: 'login' })
}

const totalRequests = ref(0)
const intrusionsBlocked = ref(0)
const averageLatency = ref(0.0)
const apiBudgetSaved = ref(0.00)

const injectionCount = ref(0)
const commandCount = ref(0)
const leakCount = ref(0)

const totalMitigated = computed(() => injectionCount.value + commandCount.value + leakCount.value)

interface BadActor {
  ip: string
  count: number
}
const badActors = ref<BadActor[]>([])

const threatTrend = ref<number[]>([0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0])

interface SecurityEvent {
  id: string
  timestamp: string
  endpoint: string
  clientIp: string
  method: string
  action: 'allow' | 'block' | 'flag'
  score: number
  categories: string[]
  matchedRules: string[]
  direction: 'inbound' | 'outbound'
  trustLevel: 'trusted' | 'semi_trusted' | 'untrusted' | 'default'
  latency: number
  inputString: string
  outputString?: string
  isFlash?: boolean
}

const events = ref<SecurityEvent[]>([])

const clearLogs = () => {
  events.value = []
}

const svgWidth = 460
const svgHeight = 120
const hoveredPoint = ref<any>(null)

const chartPoints = computed(() => {
  if (threatTrend.value.length === 0) return []
  const maxVal = Math.max(...threatTrend.value, 4)
  return threatTrend.value.map((val, idx) => {
    const x = (idx / (threatTrend.value.length - 1)) * svgWidth
    const y = svgHeight - (val / maxVal) * (svgHeight - 20) - 10

    const secondsAgo = (threatTrend.value.length - 1 - idx) * 5
    let timeLabel = `${secondsAgo}s ago`
    if (secondsAgo === 0) timeLabel = 'Now'

    return {
      x,
      y,
      val,
      timeLabel,
      idx
    }
  })
})

const linePath = computed(() => {
  if (chartPoints.value.length === 0) return ''
  const points = chartPoints.value.map(p => `${p.x},${p.y}`)
  return `M ${points.join(' L ')}`
})

const areaPath = computed(() => {
  if (chartPoints.value.length === 0) return ''
  const points = chartPoints.value.map(p => `${p.x},${p.y}`)
  return `M 0,${svgHeight} L ${points.join(' L ')} L ${svgWidth},${svgHeight} Z`
})

const selectedEvent = ref<SecurityEvent | null>(null)

const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    selectedEvent.value = null
  }
}

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeyDown)
})
</script>

<template>
  <div class="h-screen flex bg-zinc-950 text-zinc-100 font-sans relative overflow-hidden">
    <div class="absolute inset-0 minimal-dashed z-0 opacity-15 pointer-events-none"></div>

    <!-- Navigation Sidebar -->
    <aside class="w-64 border-r border-zinc-900 bg-zinc-900/20 backdrop-blur-md flex flex-col justify-between z-10 shrink-0">
      <div>
        <!-- Brand Header -->
        <div class="h-16 flex items-center gap-3 px-6 border-b border-zinc-900">
          <span class="font-bold tracking-wider text-xs text-white font-push">Cerberus</span>
        </div>

        <!-- Navigation Links -->
        <nav class="p-4 space-y-1">
          <router-link
            :to="{ name: 'builder' }"
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold text-zinc-400 hover:bg-zinc-900/60 hover:text-white transition-colors"
          >
            <Sliders class="w-3.5 h-3.5" />
            Manage Proxy
          </router-link>

          <router-link
            :to="{ name: 'testzone' }"
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold text-zinc-400 hover:bg-zinc-900/60 hover:text-white transition-colors"
          >
            <Bug class="w-3.5 h-3.5" />
            Rule Set & Playground
          </router-link>

          <router-link
            :to="{ name: 'analytics' }"
            class="flex items-center gap-3 px-3 py-2 rounded text-xs font-semibold bg-zinc-900 text-white border border-zinc-800"
          >
            <Cpu class="w-3.5 h-3.5" />
            Threat Intel
            <span class="ml-auto text-[9px] bg-zinc-800 text-zinc-355 px-1.5 py-0.5 rounded">Live</span>
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
    <main class="flex-1 flex flex-col z-10 min-w-0 overflow-y-auto relative">

      <!-- Top Header Bar -->
      <header class="h-16 border-b border-zinc-900 bg-zinc-950/20 backdrop-blur-md flex items-center justify-between px-8 shrink-0">
        <div>
          <h2 class="text-sm font-bold text-white font-push">
            Threat Intelligence Center
          </h2>
          <p class="text-[10px] text-zinc-500">Real-time reverse proxy telemetry and attack mitigation logs</p>
        </div>

        <!-- Simulation Controllers -->
        <div class="flex items-center gap-2">

          <button
            @click="clearLogs"
            class="flex items-center gap-1.5 text-[10px] font-semibold px-2.5 py-1.5 rounded border border-zinc-800 bg-zinc-900/50 text-zinc-450 hover:text-white transition-all cursor-pointer font-push"
          >
            <Trash2 class="w-3 h-3" />
            Clear Feed
          </button>
        </div>
      </header>

      <!-- Dashboard Telemetry Cards -->
      <div class="p-8 space-y-6">

        <!-- Value Realization Counters -->
        <div class="grid sm:grid-cols-2 lg:grid-cols-4 gap-6 text-left">

          <!-- Card 1: Total Requests -->
          <div class="cyber-card rounded p-4 border border-zinc-900 bg-zinc-900/5 flex items-center justify-between">
            <div class="space-y-1">
              <span class="text-[10px] font-semibold tracking-wider text-zinc-550 flex items-center gap-1 font-push">
                <TrendingUp class="w-3.5 h-3.5 text-zinc-650" />
                TOTAL ROUTED
              </span>
              <h3 class="text-2xl font-bold text-white font-push tracking-tight">{{ totalRequests.toLocaleString() }}</h3>
              <p class="text-[9px] text-zinc-550 font-medium">Agent completions proxied</p>
            </div>
            <div class="w-9 h-9 rounded bg-zinc-900 border border-zinc-800 flex items-center justify-center">
              <Activity class="w-4 h-4 text-zinc-400" />
            </div>
          </div>

          <!-- Card 2: Attacks Deflected -->
          <div class="cyber-card rounded p-4 border border-zinc-900 bg-zinc-900/5 flex items-center justify-between">
            <div class="space-y-1">
              <span class="text-[10px] font-semibold tracking-wider text-zinc-550 flex items-center gap-1 font-push">
                <ShieldAlert class="w-3.5 h-3.5 text-zinc-600" />
                ATTACKS DEFLECTED
              </span>
              <h3 class="text-2xl font-bold text-zinc-200 font-push tracking-tight">{{ intrusionsBlocked }}</h3>
              <p class="text-[9px] text-zinc-550 font-medium">100% Mitigated Rate</p>
            </div>
            <div class="w-9 h-9 rounded bg-zinc-900 border border-zinc-800 flex items-center justify-center">
              <Flame class="w-4 h-4 text-zinc-400" />
            </div>
          </div>

          <!-- Card 3: Overhead Latency -->
          <div class="cyber-card rounded p-4 border border-zinc-900 bg-zinc-900/5 flex items-center justify-between">
            <div class="space-y-1">
              <span class="text-[10px] font-semibold tracking-wider text-zinc-550 flex items-center gap-1 font-push">
                <Zap class="w-3.5 h-3.5 text-zinc-650" />
                OVERHEAD LATENCY
              </span>
              <div class="flex items-baseline gap-2">
                <h3 class="text-2xl font-bold text-white font-push tracking-tight">{{ averageLatency }}ms</h3>
                <span class="text-[8px] bg-emerald-950/30 text-emerald-400 border border-emerald-900/30 px-1 py-0.5 rounded font-sans uppercase tracking-wide">Under 2ms</span>
              </div>
              <p class="text-[9px] text-zinc-555">WASM compiler latency filter</p>
            </div>
            <div class="w-9 h-9 rounded bg-zinc-900 border border-zinc-800 flex items-center justify-center">
              <Cpu class="w-4 h-4 text-zinc-400" />
            </div>
          </div>

          <!-- Card 4: API Budget Saved -->
          <div class="cyber-card rounded p-4 border border-zinc-900 bg-zinc-900/5 flex items-center justify-between">
            <div class="space-y-1">
              <span class="text-[10px] font-semibold tracking-wider text-zinc-550 flex items-center gap-1 font-push">
                <DollarSign class="w-3.5 h-3.5 text-zinc-600" />
                LLM BUDGET SAVED
              </span>
              <h3 class="text-2xl font-bold text-emerald-450 font-push tracking-tight">${{ apiBudgetSaved.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) }}</h3>
              <p class="text-[9px] text-zinc-550 font-medium">Unused prompt tokens saved</p>
            </div>
            <div class="w-9 h-9 rounded bg-zinc-900 border border-zinc-800 flex items-center justify-center">
              <Terminal class="w-4 h-4 text-zinc-450" />
            </div>
          </div>

        </div>

        <!-- Grafana Visual Charts Grid -->
        <div class="grid lg:grid-cols-12 gap-6 text-left">

          <!-- Panel 1: Threat Mitigation Timeline Area Chart (7 Columns) -->
          <div class="lg:col-span-7 cyber-card rounded border border-zinc-900 bg-zinc-900/10 p-5 flex flex-col space-y-4">
            <div class="flex items-center justify-between border-b border-zinc-900 pb-2.5">
              <h4 class="text-[10px] font-bold text-zinc-450 uppercase tracking-widest font-push">Threat Mitigation Timeline</h4>
              <span class="text-[8px] text-zinc-600 font-mono">RATE: 5S_INTERVALS</span>
            </div>

            <!-- SVG Line Graph Container -->
            <div class="relative w-full h-[120px] bg-zinc-950/40 rounded border border-zinc-900 overflow-hidden">
              <!-- Grid gridlines -->
              <div class="absolute inset-0 flex flex-col justify-between pointer-events-none opacity-20 py-2">
                <div class="border-b border-zinc-800 w-full"></div>
                <div class="border-b border-zinc-800 w-full"></div>
                <div class="border-b border-zinc-800 w-full"></div>
              </div>

              <!-- SVG Render -->
              <svg class="w-full h-full" viewBox="0 0 460 120" preserveAspectRatio="none">
                <defs>
                  <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="#ef4444" stop-opacity="0.25" />
                    <stop offset="100%" stop-color="#ef4444" stop-opacity="0.0" />
                  </linearGradient>
                </defs>
                <!-- Area Fill -->
                <path :d="areaPath" fill="url(#areaGradient)" />
                <!-- Polyline path -->
                <path :d="linePath" stroke="#ef4444" stroke-width="1.5" fill="none" stroke-linecap="round" stroke-linejoin="round" />

                <!-- Glowing indicator for hovered point -->
                <circle
                  v-if="hoveredPoint"
                  :cx="hoveredPoint.x"
                  :cy="hoveredPoint.y"
                  r="3.5"
                  class="fill-red-500 stroke-zinc-950 stroke-[1.5]"
                />

                <!-- Faint small dots for all points so they stand out like a real chart -->
                <circle
                  v-for="p in chartPoints"
                  :key="'dot-' + p.idx"
                  :cx="p.x"
                  :cy="p.y"
                  r="1.5"
                  class="fill-red-500/50 pointer-events-none"
                />

                <!-- Invisible helper circles for easy mouse hovering -->
                <circle
                  v-for="p in chartPoints"
                  :key="'hover-' + p.idx"
                  :cx="p.x"
                  :cy="p.y"
                  r="12"
                  class="fill-transparent stroke-transparent cursor-crosshair"
                  @mouseover="hoveredPoint = p"
                  @mouseleave="hoveredPoint = null"
                />

                <!-- Inline responsive SVG tooltip -->
                <g v-if="hoveredPoint" class="pointer-events-none">
                  <rect
                    :x="hoveredPoint.x > svgWidth - 105 ? hoveredPoint.x - 100 : hoveredPoint.x + 10"
                    :y="hoveredPoint.y - 34"
                    width="90"
                    height="28"
                    rx="3"
                    fill="#18181b"
                    stroke="#3f3f46"
                    stroke-width="1"
                  />
                  <text
                    :x="hoveredPoint.x > svgWidth - 105 ? hoveredPoint.x - 94 : hoveredPoint.x + 16"
                    :y="hoveredPoint.y - 23"
                    fill="#f4f4f5"
                    font-size="7"
                    font-family="monospace"
                    font-weight="bold"
                  >
                    Threats: {{ hoveredPoint.val }}
                  </text>
                  <text
                    :x="hoveredPoint.x > svgWidth - 105 ? hoveredPoint.x - 94 : hoveredPoint.x + 16"
                    :y="hoveredPoint.y - 13"
                    fill="#a1a1aa"
                    font-size="6"
                    font-family="monospace"
                  >
                    Time: {{ hoveredPoint.timeLabel }}
                  </text>
                </g>
              </svg>
            </div>
          </div>

          <!-- Panel 2: Threat Distribution Meters (5 Columns) -->
          <div class="lg:col-span-5 cyber-card rounded border border-zinc-900 bg-zinc-900/10 p-5 flex flex-col justify-between">
            <div class="flex items-center justify-between border-b border-zinc-900 pb-2.5">
              <h4 class="text-[10px] font-bold text-zinc-450 uppercase tracking-widest font-push">Classification Share</h4>
              <span class="text-[8px] text-zinc-600 font-mono">TOTAL: {{ totalMitigated }}</span>
            </div>

            <!-- Categorized Meter lists -->
            <div class="space-y-3.5 py-2">
              <!-- Prompt Injections -->
              <div class="space-y-1">
                <div class="flex items-center justify-between text-[10px] font-medium font-mono text-zinc-350">
                  <span class="flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-red-500"></span>
                    Prompt Injections
                  </span>
                  <span>{{ injectionCount }} ({{ Math.round((injectionCount / totalMitigated) * 100) }}%)</span>
                </div>
                <div class="h-1.5 bg-zinc-950 border border-zinc-900 rounded-full overflow-hidden">
                  <div class="h-full bg-red-600 transition-all duration-500" :style="{ width: (injectionCount / totalMitigated) * 100 + '%' }"></div>
                </div>
              </div>

              <!-- Destructive Commands -->
              <div class="space-y-1">
                <div class="flex items-center justify-between text-[10px] font-medium font-mono text-zinc-350">
                  <span class="flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
                    Destructive Commands
                  </span>
                  <span>{{ commandCount }} ({{ Math.round((commandCount / totalMitigated) * 100) }}%)</span>
                </div>
                <div class="h-1.5 bg-zinc-950 border border-zinc-900 rounded-full overflow-hidden">
                  <div class="h-full bg-amber-500 transition-all duration-500" :style="{ width: (commandCount / totalMitigated) * 100 + '%' }"></div>
                </div>
              </div>

              <!-- Context/PII Leaks -->
              <div class="space-y-1">
                <div class="flex items-center justify-between text-[10px] font-medium font-mono text-zinc-350">
                  <span class="flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-zinc-400"></span>
                    Context / PII Leaks
                  </span>
                  <span>{{ leakCount }} ({{ Math.round((leakCount / totalMitigated) * 100) }}%)</span>
                </div>
                <div class="h-1.5 bg-zinc-950 border border-zinc-900 rounded-full overflow-hidden">
                  <div class="h-full bg-zinc-450 transition-all duration-500" :style="{ width: (leakCount / totalMitigated) * 100 + '%' }"></div>
                </div>
              </div>
            </div>
          </div>

        </div>

        <!-- Bad actors IP list panel and transaction logs -->
        <div class="grid lg:grid-cols-12 gap-6 items-start text-left">

          <!-- Bad Actors List (4 Columns) -->
          <div class="lg:col-span-4 cyber-card rounded border border-zinc-900 bg-zinc-900/10 p-5 space-y-4">
            <div class="flex items-center justify-between border-b border-zinc-900 pb-2.5">
              <h4 class="text-[10px] font-bold text-zinc-450 uppercase tracking-widest font-push">Bad Actor Source IPs</h4>
              <span class="text-[8px] text-zinc-650 font-mono">MITIGATED</span>
            </div>

            <div class="space-y-3">
              <div
                v-for="actor in badActors"
                :key="actor.ip"
                class="flex items-center justify-between border-b border-zinc-900/50 pb-2 text-[10px] font-mono"
              >
                <span class="text-zinc-300 font-bold">{{ actor.ip }}</span>
                <div class="flex items-center gap-2">
                  <span class="text-red-400 font-semibold">{{ actor.count }} attacks</span>
                  <div class="w-16 h-1 bg-zinc-950 border border-zinc-900 rounded-full overflow-hidden">
                    <div class="h-full bg-red-650" :style="{ width: Math.min(100, (actor.count / 60) * 100) + '%' }"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Transaction logs (8 Columns) -->
          <div class="lg:col-span-8 space-y-2.5">
            <div class="flex items-center justify-between">
              <h4 class="text-xs font-bold uppercase tracking-wider text-zinc-400 font-push">Security Transaction Logs</h4>
              <span class="text-[9px] text-zinc-500 font-mono">Displaying last 50 transactions</span>
            </div>

            <!-- Modern Log Stream Feed List (No horizontal overflow) -->
            <div class="cyber-card rounded border border-zinc-900 bg-zinc-900/10 divide-y divide-zinc-900 overflow-hidden">
              <div
                v-for="evt in events"
                :key="evt.id"
                @click="selectedEvent = evt"
                class="group flex flex-col sm:flex-row sm:items-center justify-between p-3.5 bg-zinc-950/20 hover:bg-zinc-900/30 transition-all cursor-pointer select-none gap-3"
                :class="[
                  evt.isFlash && evt.action === 'block' ? 'animate-threat-flash-red' :
                  evt.isFlash && evt.action === 'flag' ? 'animate-threat-flash-amber' :
                  evt.isFlash ? 'animate-threat-flash-zinc' : '',
                  evt.action === 'block' ? 'border-l-2 border-l-red-500' :
                  evt.action === 'flag' ? 'border-l-2 border-l-amber-500' :
                  'border-l-2 border-l-zinc-700'
                ]"
              >
                <!-- 1. Verdict & Risk Score -->
                <div class="flex items-center gap-2.5 shrink-0 sm:w-36">
                  <span
                    v-if="evt.action === 'block'"
                    class="px-2 py-0.5 rounded border border-red-955 bg-red-950/20 text-red-400 font-semibold text-[9px] inline-flex items-center gap-1 font-mono uppercase"
                  >
                    <ShieldAlert class="w-2.5 h-2.5 shrink-0" />
                    BLOCKED
                  </span>
                  <span
                    v-else-if="evt.action === 'flag'"
                    class="px-2 py-0.5 rounded border border-amber-955 bg-amber-950/20 text-amber-450 font-semibold text-[9px] inline-flex items-center gap-1 font-mono uppercase"
                  >
                    <EyeOff class="w-2.5 h-2.5 shrink-0" />
                    FLAGGED
                  </span>
                  <span
                    v-else
                    class="px-2 py-0.5 rounded border border-zinc-900 bg-zinc-950 text-zinc-500 text-[9px] inline-flex items-center gap-1 font-mono uppercase"
                  >
                    <Shield class="w-2.5 h-2.5 text-zinc-650 shrink-0" />
                    ALLOWED
                  </span>

                  <span class="text-[10px] font-mono" :class="{
                    'text-red-400 font-bold': evt.score >= 0.8,
                    'text-amber-400': evt.score >= 0.4 && evt.score < 0.8,
                    'text-zinc-600': evt.score < 0.4
                  }">
                    {{ evt.score.toFixed(2) }}
                  </span>
                </div>

                 <!-- 2. Connection Meta Detail (Taller block, vertical stacking) -->
                 <div class="flex flex-col text-left shrink-0 sm:w-52 font-mono space-y-1 py-1">
                   <!-- Row 1: Method & Endpoint -->
                   <div class="flex items-center gap-1.5 text-[10px] font-bold text-zinc-200">
                     <span class="px-1.5 py-0.5 rounded text-[8px] font-semibold tracking-wide bg-zinc-900 border border-zinc-800 text-zinc-400 uppercase">
                       {{ evt.method }}
                     </span>
                     <span class="truncate max-w-[130px]" :title="evt.endpoint">{{ evt.endpoint }}</span>
                   </div>
                   <!-- Row 2: Origin Client IP -->
                   <div class="text-[9px] text-zinc-400 font-semibold flex items-center gap-1">
                     <span class="text-zinc-650 text-[8px] uppercase tracking-wider font-bold">Origin:</span>
                     <span>{{ evt.clientIp }}</span>
                   </div>
                   <!-- Row 3: Timestamp & Trust Status -->
                   <div class="flex items-center gap-1.5 text-[8px] text-zinc-550">
                     <span>{{ evt.timestamp }}</span>
                     <span>•</span>
                     <div class="flex items-center gap-1">
                       <span class="w-1 h-1 rounded-full" :class="{
                         'bg-emerald-500': evt.trustLevel === 'trusted',
                         'bg-amber-500': evt.trustLevel === 'semi_trusted',
                         'bg-rose-500': evt.trustLevel === 'untrusted',
                         'bg-zinc-600': evt.trustLevel === 'default'
                       }"></span>
                       <span class="capitalize">{{ evt.trustLevel.replace('_', ' ') }}</span>
                     </div>
                   </div>
                 </div>

                <!-- 3. Triggered Rules & Categories -->
                <div class="flex flex-wrap gap-1 items-center text-left sm:w-44 shrink-0 font-mono">
                  <span v-if="evt.categories.length === 0" class="text-zinc-600 text-[10px] font-mono">-</span>
                  <template v-else>
                    <span
                      v-for="cat in evt.categories"
                      :key="cat"
                      class="px-1.5 py-0.5 rounded bg-zinc-900 text-zinc-350 text-[8px] border border-zinc-850"
                    >
                      {{ cat }}
                    </span>
                    <span
                      v-for="rule in evt.matchedRules"
                      :key="rule"
                      class="text-[8px] text-zinc-600 max-w-[90px] truncate"
                      :title="rule"
                    >
                      ({{ rule }})
                    </span>
                  </template>
                </div>

                <!-- 4. Latency / Connection Speed -->
                <div class="shrink-0 text-left sm:w-16 font-mono text-[9px] text-zinc-450">
                  {{ evt.latency }} ms
                </div>

                <!-- 5. Dynamic Payload Details -->
                <div class="flex-1 text-left font-mono text-[10px] text-zinc-500 truncate pr-2">
                  <span class="text-zinc-600 font-normal">Payload: </span>
                  <span :class="{'text-red-400 font-medium': evt.action === 'block', 'text-amber-400': evt.action === 'flag'}">
                    {{ evt.inputString }}
                  </span>
                </div>
              </div>

              <!-- Empty state -->
              <div v-if="events.length === 0" class="px-5 py-12 text-center text-zinc-600 font-mono text-xs">
                No active connection logs found.
              </div>
            </div>
          </div>

        </div>

      </div>

      <!-- Payload Diff Explorer side-by-side modal -->
      <div
        v-if="selectedEvent"
        class="fixed inset-0 z-50 flex items-center justify-center p-6 bg-zinc-950/80 backdrop-blur-sm"
        @click.self="selectedEvent = null"
      >
        <div class="cyber-card rounded-lg border border-zinc-800 bg-zinc-950 max-w-3xl w-full max-h-[85vh] flex flex-col shadow-2xl overflow-hidden text-left relative z-50">

          <!-- Modal Header -->
          <div class="p-4 border-b border-zinc-900 flex items-center justify-between bg-zinc-900/30">
            <div class="space-y-0.5">
              <span class="text-[9px] text-zinc-500 font-mono tracking-wider font-semibold">TRANSACTION ID: {{ selectedEvent.id }}</span>
              <h3 class="text-xs font-bold text-white font-push flex items-center gap-2">
                Payload Diff Explorer
                <span
                  v-if="selectedEvent.action === 'block'"
                  class="px-2 py-0.5 rounded bg-red-950/30 text-red-400 border border-red-900/20 text-[9px] font-mono font-bold"
                >
                  BLOCKED
                </span>
                <span
                  v-else-if="selectedEvent.action === 'flag'"
                  class="px-2 py-0.5 rounded bg-amber-950/30 text-amber-400 border border-amber-900/20 text-[9px] font-mono font-bold"
                >
                  SANITIZED
                </span>
                <span
                  v-else
                  class="px-2 py-0.5 rounded bg-emerald-950/30 text-emerald-400 border border-emerald-900/20 text-[9px] font-mono font-bold"
                >
                  PASSED
                </span>
              </h3>
            </div>

            <button
              @click="selectedEvent = null"
              class="text-zinc-500 hover:text-white transition-colors cursor-pointer text-xs font-mono font-bold hover:bg-zinc-900 px-2 py-1 rounded"
            >
              CLOSE [ESC]
            </button>
          </div>

          <!-- Modal Content -->
          <div class="p-6 overflow-y-auto space-y-5 flex-1 bg-zinc-950 select-text">
            <!-- Meta Grid -->
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-[10px]">
              <div class="border border-zinc-900 rounded p-2.5 bg-zinc-900/20">
                <span class="text-zinc-550 block text-[8px] uppercase font-push">Threat Score</span>
                <span class="text-xs font-mono font-bold" :class="{
                  'text-red-400': selectedEvent.score >= 0.8,
                  'text-amber-400': selectedEvent.score >= 0.4 && selectedEvent.score < 0.8,
                  'text-emerald-400': selectedEvent.score < 0.4
                }">{{ selectedEvent.score.toFixed(2) }}</span>
              </div>
              <div class="border border-zinc-900 rounded p-2.5 bg-zinc-900/20">
                <span class="text-zinc-550 block text-[8px] uppercase font-push">Proxy Overhead</span>
                <span class="text-xs font-mono font-bold text-zinc-300">{{ selectedEvent.latency }} ms</span>
              </div>
              <div class="border border-zinc-900 rounded p-2.5 bg-zinc-900/20">
                <span class="text-zinc-550 block text-[8px] uppercase font-push">Client Trust</span>
                <span class="text-xs font-mono font-bold text-zinc-300 capitalize">{{ selectedEvent.trustLevel }}</span>
              </div>
              <div class="border border-zinc-900 rounded p-2.5 bg-zinc-900/20">
                <span class="text-zinc-550 block text-[8px] uppercase font-push">Triggered Filters</span>
                <span class="text-xs font-mono font-bold text-zinc-300 truncate block">
                  {{ selectedEvent.categories.join(', ') || 'None' }}
                </span>
              </div>
            </div>

            <!-- Payload Diff Side-By-Side Grid -->
            <div class="grid md:grid-cols-2 gap-6">
              <!-- Left side: Input -->
              <div class="space-y-2 flex flex-col">
                <div class="flex items-center justify-between">
                  <span class="text-[9px] font-bold text-zinc-400 uppercase tracking-wider font-push">Raw User Input Payload</span>
                  <span class="text-[8px] text-zinc-650 font-mono">DIRECTION: INBOUND</span>
                </div>
                <div class="flex-1 min-h-[140px] rounded border border-zinc-900 bg-zinc-950 p-3.5 font-mono text-[10px] text-zinc-300 whitespace-pre-wrap leading-relaxed select-text" :class="[selectedEvent.action === 'block' ? 'border-red-955 ring-1 ring-red-950/20' : '']">
                  {{ selectedEvent.inputString }}
                </div>
              </div>

              <!-- Right side: Output -->
              <div class="space-y-2 flex flex-col">
                <div class="flex items-center justify-between">
                  <span class="text-[9px] font-bold text-zinc-400 uppercase tracking-wider font-push">Sanitized Target Output</span>
                  <span class="text-[8px] text-zinc-650 font-mono">DIRECTION: OUTBOUND</span>
                </div>
                <div class="flex-1 min-h-[140px] rounded border border-zinc-900 bg-zinc-955 p-3.5 font-mono text-[10px] whitespace-pre-wrap leading-relaxed select-text" :class="[
                  selectedEvent.action === 'block' ? 'border-red-955 bg-red-950/5 text-red-400/90 font-semibold' :
                  selectedEvent.action === 'flag' ? 'border-amber-955 bg-amber-950/5 text-amber-400/90' :
                  'text-zinc-400'
                ]">
                  <span v-if="selectedEvent.action === 'block'">
                    [CONNECTION TERMINATED] Active rule block triggered. Request payload rejected by Ingress Sentry.

                    Reason: Matched signatures [{{ selectedEvent.matchedRules.join(', ') }}] with high threat score {{ selectedEvent.score }}.
                  </span>
                  <span v-else-if="selectedEvent.action === 'flag'">
                    {{ selectedEvent.outputString || selectedEvent.inputString }}
                  </span>
                  <span v-else>
                    {{ selectedEvent.inputString }}
                  </span>
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

@keyframes threatRedFlash {
  0% { background-color: rgba(239, 68, 68, 0.25); border-color: rgba(239, 68, 68, 0.4); }
  100% { background-color: transparent; }
}

@keyframes threatAmberFlash {
  0% { background-color: rgba(245, 158, 11, 0.25); border-color: rgba(245, 158, 11, 0.35); }
  100% { background-color: transparent; }
}

@keyframes threatZincFlash {
  0% { background-color: rgba(63, 63, 70, 0.25); }
  100% { background-color: transparent; }
}

.animate-threat-flash-red {
  animation: threatRedFlash 1.5s ease-out;
}

.animate-threat-flash-amber {
  animation: threatAmberFlash 1.5s ease-out;
}

.animate-threat-flash-zinc {
  animation: threatZincFlash 1.2s ease-out;
}
</style>
