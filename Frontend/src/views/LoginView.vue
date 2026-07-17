<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Terminal } from '@lucide/vue'

const router = useRouter()
const isLoading = ref(false)
const terminalLogs = ref<string[]>([])

const mouseX = ref(0)
const mouseY = ref(0)
const orbX = ref(0)
const orbY = ref(0)
const canvasRef = ref<HTMLCanvasElement | null>(null)

// Sandbox click ripple easter egg state
interface ClickRipple {
  x: number
  y: number
  radius: number
  maxRadius: number
  speed: number
  alpha: number
}
const clickRipples = ref<ClickRipple[]>([])

// Reactive sandbox card flash outline feedback state
const sandboxFlash = ref<'block' | 'flag' | null>(null)

onMounted(() => {
  // Initialize positions to center of screen
  mouseX.value = window.innerWidth / 2
  mouseY.value = window.innerHeight / 2
  orbX.value = window.innerWidth / 2
  orbY.value = window.innerHeight / 2

  // Initial terminal status message
  terminalLogs.value.push('Cerberus Control Plane offline. Waiting for session authorization...')

  // Initialize high-performance Ripple Canvas Grid
  const canvas = canvasRef.value
  if (canvas) {
    const ctx = canvas.getContext('2d')
    if (ctx) {
      const dpr = window.devicePixelRatio || 1
      let width = window.innerWidth
      let height = window.innerHeight
      
      canvas.width = width * dpr
      canvas.height = height * dpr
      ctx.scale(dpr, dpr)

      const handleResize = () => {
        if (!canvas) return
        width = window.innerWidth
        height = window.innerHeight
        canvas.width = width * dpr
        canvas.height = height * dpr
        ctx.scale(dpr, dpr)
      }
      window.addEventListener('resize', handleResize)

      const spacing = 24 // Space between dots
      const maxDist = 220 // Area of ripple influence around light ball

      const updateLoop = () => {
        // Gravitational lag easing calculation
        const ease = 0.05
        orbX.value = orbX.value + (mouseX.value - orbX.value) * ease
        orbY.value = orbY.value + (mouseY.value - orbY.value) * ease

        // Update active click ripples
        for (let i = clickRipples.value.length - 1; i >= 0; i--) {
          const rip = clickRipples.value[i]
          rip.radius += rip.speed
          rip.alpha = 1 - rip.radius / rip.maxRadius
          if (rip.radius >= rip.maxRadius) {
            clickRipples.value.splice(i, 1)
          }
        }

        ctx.clearRect(0, 0, width, height)

        const startX = (width % spacing) / 2
        const startY = (height % spacing) / 2

        for (let x0 = startX; x0 < width; x0 += spacing) {
          for (let y0 = startY; y0 < height; y0 += spacing) {
            const dx = x0 - orbX.value
            const dy = y0 - orbY.value
            const dist = Math.sqrt(dx * dx + dy * dy)

            let x = x0
            let y = y0
            let opacity = 0.08
            let dotSize = 0.7
            let r = 63
            let g = 63
            let b = 70

            // Hover Spotlight Repulsion & Color Blends
            if (dist < maxDist) {
              const force = (maxDist - dist) / maxDist

              // Ripple effect: repel dots outward relative to the distance of floating orb
              const push = force * 5
              x = x0 + (dx / (dist || 1)) * push
              y = y0 + (dy / (dist || 1)) * push

              // Expand and highlight dots near the gravitating core
              opacity = 0.08 + force * 0.14
              dotSize = 0.7 + force * 0.3

              // Soft crimson tone shift
              r = Math.round(63 + force * 60)
              g = Math.round(63 - force * 10)
              b = Math.round(70 - force * 15)
            }

            // Click Ripples Repulsion & Glowing Wave effect
            for (let i = 0; i < clickRipples.value.length; i++) {
              const rip = clickRipples.value[i]
              const dxClick = x0 - rip.x
              const dyClick = y0 - rip.y
              const distClick = Math.sqrt(dxClick * dxClick + dyClick * dyClick)
              const diff = Math.abs(distClick - rip.radius)

              if (diff < 24) {
                const force = (1 - diff / 24) * rip.alpha
                
                // Repel dots along the expanding click ring wave vector
                const pushClick = force * 4
                x += (dxClick / (distClick || 1)) * pushClick
                y += (dyClick / (distClick || 1)) * pushClick

                // Pulse parameters
                dotSize += force * 0.5
                opacity = Math.min(0.6, opacity + force * 0.25)
                
                // Shift colors to bright core red on wave pass
                r = Math.min(239, Math.round(r + force * 120))
                g = Math.max(30, Math.round(g - force * 20))
                b = Math.max(35, Math.round(b - force * 25))
              }
            }

            ctx.fillStyle = `rgba(${r}, ${g}, ${b}, ${opacity})`
            ctx.beginPath()
            ctx.arc(x, y, dotSize, 0, Math.PI * 2)
            ctx.fill()
          }
        }

        requestAnimationFrame(updateLoop)
      }
      requestAnimationFrame(updateLoop)
    }
  }
  
  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})

const showStickyHeader = ref(false)
const handleScroll = () => {
  showStickyHeader.value = window.scrollY > 400
}

const handleCanvasClick = (e: MouseEvent) => {
  // Push a new expanding ripple into grid workspace
  clickRipples.value.push({
    x: e.clientX,
    y: e.clientY,
    radius: 0,
    maxRadius: 280,
    speed: 7,
    alpha: 1
  })
}

const handleGithubLogin = () => {
  isLoading.value = true
  terminalLogs.value.push('Authenticating with GitHub OAuth API...')
  
  setTimeout(() => {
    terminalLogs.value.push('OAuth Token verified. Handshake completed.')
    terminalLogs.value.push('Initializing developer control plane session...')
    
    setTimeout(() => {
      localStorage.setItem('cerberus_auth', 'true')
      router.push({ name: 'builder' })
    }, 500)
  }, 800)
}

const handleMouseMove = (e: MouseEvent) => {
  mouseX.value = e.clientX
  mouseY.value = e.clientY
}

const scrollToInfo = () => {
  document.getElementById('info-section')?.scrollIntoView({ behavior: 'smooth' })
}

// Sandbox Interactive Playpen States
const customInput = ref('')
const sandboxOutput = ref('')
const sandboxVerdict = ref<any>({
  action: 'allow',
  score: 0.00,
  categories: [],
  matchedRules: [],
  direction: 'inbound',
  trustLevel: 'trusted'
})

const activeScenario = computed(() => {
  return {
    name: 'Interactive Playground',
    input: customInput.value,
    output: sandboxOutput.value,
    verdict: sandboxVerdict.value
  }
})

const isTestingRealProxy = ref(false)
const sandboxError = ref('')

const runRealProxyCheck = async (promptText: string) => {
  if (!promptText) return
  isTestingRealProxy.value = true
  sandboxError.value = ''
  
  try {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json'
    }
    // Optional client key if known
    const storedClientKey = localStorage.getItem('cerberus_last_client_key')
    if (storedClientKey) {
      headers['Authorization'] = `Bearer ${storedClientKey}`
    }

    const res = await fetch('/v1/chat/completions', {
      method: 'POST',
      headers,
      body: JSON.stringify({
        model: 'gpt-4o',
        messages: [{ role: 'user', content: promptText }]
      })
    })

    if (res.status === 403) {
      const data = await res.json()
      sandboxVerdict.value = {
        action: data.action || 'block',
        score: data.score || 0.99,
        categories: data.categories || ['prompt_injection'],
        matchedRules: data.matched_rules || ['rule_intercepted'],
        direction: data.direction || 'inbound',
        trustLevel: data.trust_level || 'untrusted'
      }
      sandboxOutput.value = '[CONNECTION TERMINATED] Threat intercepted by Ingress Sentry.'
      triggerFlash(sandboxVerdict.value.action)
    } else if (res.status === 401) {
      sandboxError.value = 'Go Gateway requires a valid client key (Auth enabled).'
      sandboxVerdict.value = {
        action: 'block',
        score: 0.95,
        categories: ['auth_failure'],
        matchedRules: ['invalid_authorization_token'],
        direction: 'inbound',
        trustLevel: 'untrusted'
      }
      sandboxOutput.value = '[UNAUTHORIZED] Access denied by Cerberus Core.'
    } else if (res.ok) {
      const data = await res.json()
      if (data.choices && data.choices[0] && data.choices[0].message) {
        sandboxOutput.value = data.choices[0].message.content
      } else {
        sandboxOutput.value = 'Forwarded to upstream successfully. Output format unrecognized.'
      }
      sandboxVerdict.value = {
        action: 'allow',
        score: 0.01,
        categories: [],
        matchedRules: [],
        direction: 'inbound',
        trustLevel: 'trusted'
      }
    } else {
      const outboundHeader = res.headers.get('X-Cerberus-Outbound')
      if (outboundHeader === 'block') {
        sandboxVerdict.value = {
          action: 'block',
          score: 0.99,
          categories: ['context_leak'],
          matchedRules: ['system_prompt_echo'],
          direction: 'outbound',
          trustLevel: 'untrusted'
        }
        sandboxOutput.value = '[EXECUTION BLOCKED] Response output blocked by Egress Censor.'
        triggerFlash('block')
      } else if (outboundHeader === 'flag') {
        sandboxVerdict.value = {
          action: 'flag',
          score: 0.65,
          categories: ['context_leak'],
          matchedRules: ['rule_credential_exposure'],
          direction: 'outbound',
          trustLevel: 'semi_trusted'
        }
        sandboxOutput.value = 'Processing stream data... Credentials masked securely: [REDACTED_SECURE]'
        triggerFlash('flag')
      } else {
        sandboxVerdict.value = {
          action: 'allow',
          score: 0.05,
          categories: [],
          matchedRules: [],
          direction: 'inbound',
          trustLevel: 'default'
        }
        sandboxOutput.value = `Passed scan successfully. Upstream status: ${res.status}.`
      }
    }
  } catch (err: any) {
    console.error('Proxy check failed:', err)
    sandboxError.value = 'Go Gateway is offline. Start backend server on port 8080.'
    runLocalSimulation(promptText)
  } finally {
    isTestingRealProxy.value = false
  }
}

const runLocalSimulation = (promptText: string) => {
  const text = promptText.toLowerCase()
  let action: 'allow' | 'block' | 'flag' = 'allow'
  let score = 0.01
  let categories: string[] = []
  let matchedRules: string[] = []
  let direction: 'inbound' | 'outbound' = 'inbound'
  let textTrust: 'trusted' | 'semi_trusted' | 'untrusted' | 'default' = 'default'
  let output = promptText

  if (text.includes('ignore') || text.includes('jailbreak') || text.includes('instruction')) {
    action = 'block'
    score = 0.95
    categories = ['prompt_injection']
    matchedRules = ['rule_obfuscation_patterns']
    textTrust = 'untrusted'
    output = '[CONNECTION TERMINATED] Jailbreak vector intercepted by Ingress Sentry.'
  } else if (text.includes('rm ') || text.includes('sudo') || text.includes('drop table')) {
    action = 'block'
    score = 0.99
    categories = ['destructive_command']
    matchedRules = ['rule_dangerous_system_call']
    textTrust = 'untrusted'
    output = '[EXECUTION BLOCKED] System execution terminated by Runtime Firewall.'
  } else if (text.includes('sk_') || text.includes('key') || text.includes('token')) {
    action = 'flag'
    score = 0.70
    categories = ['context_leak']
    matchedRules = ['rule_credential_exposure']
    direction = 'outbound'
    textTrust = 'semi_trusted'
    output = 'Processing stream data... Credentials masked securely: [REDACTED_SECURE]'
  } else {
    output = `Standard pass-through query processed. Response forwarded.`
  }

  sandboxVerdict.value = { action, score, categories, matchedRules, direction, trustLevel: textTrust }
  sandboxOutput.value = output
}

let debounceTimer: any = null
watch(customInput, (newVal) => {
  clearTimeout(debounceTimer)
  if (newVal) {
    debounceTimer = setTimeout(() => {
      runRealProxyCheck(newVal)
    }, 450)
  } else {
    sandboxOutput.value = ''
    sandboxVerdict.value = {
      action: 'allow',
      score: 0.00,
      categories: [],
      matchedRules: [],
      direction: 'inbound',
      trustLevel: 'trusted'
    }
  }
})

// Trigger a soft outline warning flash inside sandbox card when threat is caught
const triggerFlash = (action: 'allow' | 'block' | 'flag') => {
  if (action === 'block' || action === 'flag') {
    sandboxFlash.value = action
    setTimeout(() => {
      sandboxFlash.value = null
    }, 600)
  }
}

watch(() => activeScenario.value.verdict.action, (newAction) => {
  triggerFlash(newAction)
})

const animationClass = computed(() => {
  const action = activeScenario.value.verdict.action
  if (action === 'block') return 'animate-block'
  if (action === 'flag') return 'animate-flag'
  return 'animate-allow'
})
</script>

<template>
  <div 
    @mousemove="handleMouseMove"
    @click="handleCanvasClick"
    class="min-h-screen flex flex-col bg-zinc-950 text-zinc-100 font-sans relative overflow-y-auto scroll-smooth"
  >
    <!-- Sticky Top Navigation Header (Appears on Scroll) -->
    <header 
      class="fixed top-0 left-0 w-full h-14 border-b border-zinc-900 bg-zinc-950/80 backdrop-blur-md z-40 transition-all duration-300 flex items-center justify-between px-8"
      :class="[showStickyHeader ? 'translate-y-0 opacity-100' : '-translate-y-full opacity-0']"
    >
      <!-- Left Side: Brand Logo & Title -->
      <div class="flex items-center gap-2.5">
        <span class="font-bold tracking-wider text-xs text-white font-push">Cerberus</span>
      </div>

      <!-- Right Side: Continue with GitHub CTA -->
      <button 
        @click="handleGithubLogin" 
        :disabled="isLoading"
        class="flex items-center justify-center gap-2 bg-zinc-100 hover:bg-white text-zinc-950 font-bold text-[10px] py-1.5 px-4 rounded transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer border border-zinc-200 font-push"
      >
        <svg v-if="!isLoading" class="w-3 h-3 fill-current" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/>
        </svg>
        <span v-if="!isLoading">Connect GitHub</span>
        <span v-else class="flex items-center gap-1.5">
          <svg class="animate-spin h-3 w-3 text-zinc-950" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          Syncing...
        </span>
      </button>
    </header>

    <!-- High-Performance Canvas Dot Matrix Ripple Background -->
    <canvas ref="canvasRef" class="fixed inset-0 w-full h-full z-0 pointer-events-none"></canvas>

    <!-- Wide Cursor Follower Glow (Instant soft illumination wrapper) -->
    <div 
      class="absolute inset-0 pointer-events-none z-0 transition-opacity duration-300"
      :style="{
        background: `radial-gradient(600px circle at ${mouseX}px ${mouseY}px, rgba(239, 68, 68, 0.03), transparent 80%)`
      }"
    ></div>

    <!-- Gravitating Orb Core (Fluid red lag lighting) -->
    <div 
      class="absolute inset-0 pointer-events-none z-0 transition-opacity duration-300"
      :style="{
        background: `radial-gradient(180px circle at ${orbX}px ${orbY}px, rgba(239, 68, 68, 0.12), transparent 75%)`
      }"
    ></div>

    <!-- Fixed Glowing Red Ball Spots (Aesthetic Accents) -->
    <div class="fixed top-[15%] right-[10%] w-[450px] h-[450px] bg-red-600/5 rounded-full blur-[140px] pointer-events-none z-0 animate-pulse-slow"></div>
    <div class="fixed bottom-[20%] left-[5%] w-[350px] h-[350px] bg-red-800/4 rounded-full blur-[110px] pointer-events-none z-0"></div>
    <div class="fixed top-[60%] right-[45%] w-[250px] h-[250px] bg-red-500/6 rounded-full blur-[90px] pointer-events-none z-0"></div>

    <!-- Hero Screen Section -->
    <section class="min-h-screen flex flex-col items-center justify-center px-6 relative z-10 text-center space-y-8">
      
      <div class="space-y-3.5 select-none">
        <h1 class="text-6xl md:text-7xl font-bold tracking-tight text-white leading-none font-push">
          Cerberus
        </h1>
        
        <div class="space-y-3">
          <h2 class="text-xs md:text-sm font-semibold tracking-[0.25em] text-zinc-400 uppercase font-push">
            Build Fast. Secure All.
          </h2>
          <p class="text-[11px] text-zinc-500 max-w-sm mx-auto leading-relaxed">
            Zero-Trust Reverse Proxy & Real-Time Guardrails for AI Agents & MCP Endpoints
          </p>
        </div>
      </div>

      <!-- Action Button Group -->
      <div class="flex flex-col sm:flex-row gap-4 items-center justify-center w-full max-w-md mx-auto relative z-20">
        <!-- Button 1: Auth -->
        <button 
          @click="handleGithubLogin" 
          :disabled="isLoading"
          class="w-full sm:w-auto min-w-[180px] flex items-center justify-center gap-2.5 bg-zinc-100 hover:bg-white text-zinc-950 font-bold text-xs py-3 px-6 rounded transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer border border-zinc-200"
        >
          <svg class="w-4 h-4 fill-current" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 .297c-6.63 0-12 5.373-12 12 0 5.303 3.438 9.8 8.205 11.385.6.113.82-.258.82-.577 0-.285-.01-1.04-.015-2.04-3.338.724-4.042-1.61-4.042-1.61C4.422 18.07 3.633 17.7 3.633 17.7c-1.087-.744.084-.729.084-.729 1.205.084 1.838 1.236 1.838 1.236 1.07 1.835 2.809 1.305 3.495.998.108-.776.417-1.305.76-1.605-2.665-.3-5.466-1.332-5.466-5.93 0-1.31.465-2.38 1.235-3.22-.135-.303-.54-1.523.105-3.176 0 0 1.005-.322 3.3 1.23.96-.267 1.98-.399 3-.405 1.02.006 2.04.138 3 .405 2.28-1.552 3.285-1.23 3.285-1.23.645 1.653.24 2.873.12 3.176.765.84 1.23 1.91 1.23 3.22 0 4.61-2.805 5.625-5.475 5.92.42.36.81 1.096.81 2.22 0 1.606-.015 2.896-.015 3.286 0 .315.21.69.825.57C20.565 22.092 24 17.592 24 12.297c0-6.627-5.373-12-12-12"/>
          </svg>
          <span v-if="!isLoading">Continue with GitHub</span>
          <span v-else class="flex items-center gap-2">
            <svg class="animate-spin h-3.5 w-3.5 text-zinc-950" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Syncing Cluster...
          </span>
        </button>

        <!-- Button 2: How it works -->
        <button 
          @click="scrollToInfo"
          class="w-full sm:w-auto min-w-[180px] flex items-center justify-center border border-zinc-850 hover:border-zinc-700 bg-zinc-900/20 hover:bg-zinc-900/40 text-zinc-400 hover:text-zinc-200 font-semibold text-xs py-3 px-6 rounded transition-all cursor-pointer font-push"
        >
          How It Works
        </button>
      </div>

    </section>

    <!-- Info Detail Section below the fold -->
    <section id="info-section" class="border-t border-zinc-900 bg-zinc-950/40 relative z-10 px-6 py-20">
      <div class="max-w-5xl mx-auto space-y-12">
        
        <!-- Section Header -->
        <div class="text-left max-w-2xl space-y-1.5">
          <h2 class="text-lg font-bold text-white tracking-tight font-push">How Cerberus Works</h2>
          <p class="text-xs text-zinc-400 leading-relaxed">
            Cerberus is a security guard for your AI applications. It sits between your users and your AI model, scanning requests for safety threats (like prompt injections or data leaks) and blocking or cleaning them automatically.
          </p>
        </div>

        <!-- 3. Animated Request Flow Diagram (Visual schematic) -->
        <div class="cyber-card rounded p-6 border border-zinc-900 bg-zinc-950/10 flex flex-col space-y-4">
          <div class="flex items-center justify-between w-full max-w-2xl mx-auto relative py-6">
            <!-- Continuous Connection lines -->
            <div class="absolute left-[10%] right-[10%] top-1/2 -translate-y-1/2 h-0.5 bg-zinc-900 z-0"></div>
            
            <!-- Animated Tracer Dot with Gradient Trail -->
            <div 
              class="absolute top-1/2 -translate-y-1/2 w-2.5 h-2.5 rounded-full z-10"
              :class="[animationClass]"
            ></div>

            <!-- Node 1: Client Ingress -->
            <div class="z-20 w-28 h-10 rounded border border-zinc-900 bg-zinc-950/80 flex flex-col justify-center items-center shadow-md">
              <span class="text-[9px] text-zinc-450 uppercase font-bold tracking-wider font-push">User Request</span>
            </div>

            <!-- Node 2: Cerberus Proxy -->
            <div 
              class="z-20 w-36 h-12 rounded border flex flex-col justify-center items-center shadow-lg transition-colors duration-300"
              :class="[
                activeScenario.verdict.action === 'block' ? 'border-red-900 bg-red-950/10 shadow-[0_0_15px_rgba(239,68,68,0.08)]' :
                activeScenario.verdict.action === 'flag' ? 'border-amber-900 bg-amber-950/10 shadow-[0_0_15px_rgba(245,158,11,0.08)]' :
                'border-zinc-800 bg-zinc-950/95'
              ]"
            >
              <span class="text-[10px] text-white font-bold font-push">Cerberus Proxy</span>
              <span class="text-[8px] mt-0.5 font-bold uppercase tracking-wider font-mono" :class="{
                'text-red-400': activeScenario.verdict.action === 'block',
                'text-amber-400': activeScenario.verdict.action === 'flag',
                'text-emerald-400': activeScenario.verdict.action === 'allow'
              }">
                [{{ activeScenario.verdict.action }}]
              </span>
            </div>

            <!-- Node 3: LLM Engine -->
            <div class="z-20 w-28 h-10 rounded border border-zinc-900 bg-zinc-950/80 flex flex-col justify-center items-center shadow-md">
              <span class="text-[9px] text-zinc-450 uppercase font-bold tracking-wider font-push">Target AI Model</span>
            </div>
          </div>
        </div>

        <!-- Section Header 2 -->
        <div class="text-left max-w-2xl space-y-1.5 pt-8">
          <h2 class="text-lg font-bold text-white tracking-tight font-push">Interactive Sandbox</h2>
          <p class="text-xs text-zinc-450 leading-relaxed font-sans">
            Select a scenario below to see how Cerberus handles requests in real-time. If you are a developer, you can inspect the exact structural verdict JSON payload returned by the proxy further down.
          </p>
        </div>

        <!-- 2. Interactive Threat Playground (Interactive Sandbox) -->
        <div class="grid md:grid-cols-12 gap-8 items-start">
          
          <!-- Sandbox Playpen Interactive Panel (Left Column) -->
          <div class="md:col-span-7 space-y-3 text-left">
            <div 
              class="cyber-card rounded p-5 border bg-zinc-900/10 space-y-4 transition-all duration-500"
              :class="[
                sandboxFlash === 'block' ? 'border-red-500 shadow-[0_0_15px_rgba(239,68,68,0.15)] duration-75' :
                sandboxFlash === 'flag' ? 'border-amber-500 shadow-[0_0_15px_rgba(245,158,11,0.15)] duration-75' :
                'border-zinc-900'
              ]"
            >


              <!-- Input text box -->
              <div class="space-y-1.5">
                <label class="text-[9px] text-zinc-550 font-bold uppercase tracking-wider font-push">Prompt Sent to AI:</label>
                <textarea 
                  v-model="customInput" 
                  placeholder="Type your prompt here (e.g. 'ignore previous rules' or commands like 'sudo rm')..."
                  rows="2"
                  class="w-full text-xs bg-zinc-950 border border-zinc-900 text-zinc-300 p-2.5 rounded focus:ring-1 focus:ring-zinc-700 focus:outline-none font-mono placeholder:text-zinc-700"
                ></textarea>
              </div>

              <!-- Output Box -->
              <div class="space-y-1.5">
                <span class="text-[9px] text-zinc-550 font-bold uppercase tracking-wider font-push">Filtered Result (Forwarded to Target):</span>
                <div class="w-full text-xs bg-zinc-950/80 border border-zinc-900 text-zinc-400 p-2.5 rounded font-mono min-h-[48px] select-all">
                  {{ activeScenario.output }}
                </div>
              </div>
            </div>
          </div>

          <!-- Live Verdict JSON Payload Output (Right Column) -->
          <div class="md:col-span-5 space-y-3 text-left">
            <div 
              class="cyber-card rounded p-5 border bg-zinc-900/10 space-y-4 h-full flex flex-col justify-between transition-all duration-500"
              :class="[
                sandboxFlash === 'block' ? 'border-red-500 shadow-[0_0_15px_rgba(239,68,68,0.15)] duration-75' :
                sandboxFlash === 'flag' ? 'border-amber-500 shadow-[0_0_15px_rgba(245,158,11,0.15)] duration-75' :
                'border-zinc-900'
              ]"
            >
              <!-- Live metrics -->
              <div class="grid grid-cols-2 gap-3 text-[10px]">
                <div class="border border-zinc-900 rounded p-2 bg-zinc-950/40">
                  <span class="text-zinc-555 block text-[8px] uppercase tracking-wide font-push">Risk score (0 - 1)</span>
                  <span class="text-sm font-bold font-push mt-0.5 block" :class="{
                    'text-red-400': activeScenario.verdict.score >= 0.8,
                    'text-amber-400': activeScenario.verdict.score >= 0.4 && activeScenario.verdict.score < 0.8,
                    'text-emerald-400': activeScenario.verdict.score < 0.4
                  }">{{ activeScenario.verdict.score.toFixed(2) }}</span>
                </div>
                <div class="border border-zinc-900 rounded p-2 bg-zinc-950/40">
                  <span class="text-zinc-555 block text-[8px] uppercase tracking-wide font-push">Verdict Action</span>
                  <span class="text-sm font-bold text-zinc-300 font-push capitalize mt-0.5 block">{{ activeScenario.verdict.action }}</span>
                </div>
              </div>

              <!-- Live Go JSON output -->
              <div class="cyber-card rounded p-3 font-mono text-[9px] text-zinc-450 bg-zinc-950 border border-zinc-900 overflow-y-auto max-h-40">
                <pre class="leading-relaxed">{{ JSON.stringify(activeScenario.verdict, null, 2) }}</pre>
              </div>
            </div>
          </div>

        </div>

        <!-- Section Header 3 -->
        <div class="text-left max-w-2xl space-y-1.5 pt-8 border-t border-zinc-900">
          <h2 class="text-lg font-bold text-white tracking-tight font-push">Technical Subsystems</h2>
          <p class="text-xs text-zinc-550 leading-relaxed font-sans">
            Under the hood, Cerberus compiles developer-defined safety rules into a high-performance Go reverse proxy pipeline executing across three dedicated security layers:
          </p>
        </div>

        <!-- Section Footer Grid -->
        <div class="grid md:grid-cols-12 gap-8 items-start">
          
          <!-- Terminal logs (left column) -->
          <div class="md:col-span-5 space-y-3">
            <div class="cyber-card rounded p-4 font-mono text-[11px] text-zinc-555 border border-zinc-900 bg-zinc-900/10 h-56 overflow-y-auto text-left">
              <div class="flex items-center justify-between border-b border-zinc-900 pb-2 mb-2">
                <span class="flex items-center gap-1.5 font-bold uppercase text-[9px] tracking-wider text-zinc-500 font-push">
                  <Terminal class="w-3 h-3 text-zinc-650" />
                  Terminal Console
                </span>
                
                <span class="flex items-center gap-1.5 text-[8px] bg-red-950/30 text-red-400 px-1.5 py-0.5 rounded border border-red-900/20 font-sans">
                  <span class="w-1.5 h-1.5 rounded-full bg-red-500 animate-pulse shrink-0"></span>
                  Gateway Active
                </span>
              </div>
              <div class="space-y-1.5">
                <div v-for="(log, idx) in terminalLogs" :key="idx" class="flex gap-2">
                  <span class="text-zinc-700 select-none">$</span>
                  <span :class="{'text-zinc-300': !log.includes('Auth') && !log.includes('verified'), 'text-zinc-450': log.includes('Auth') || log.includes('verified')}">{{ log }}</span>
                </div>
                <div v-if="isLoading" class="animate-pulse flex gap-2 text-zinc-500">
                  <span class="text-zinc-700 select-none">$</span>
                  <span>Syncing with gateway core...</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Feature items list: The Three Heads of Cerberus (right column) -->
          <div class="md:col-span-7 space-y-3 text-left">
            <div class="grid gap-3 font-sans">
              <!-- Node 1 -->
              <div class="cyber-card rounded p-4 border border-zinc-900 bg-zinc-950/20 hover:border-zinc-800 transition-colors space-y-1">
                <div class="flex items-center justify-between font-push">
                  <span class="text-[10px] font-bold text-zinc-300 flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-red-500"></span>
                    Ingress Sentry (Left Head)
                  </span>
                </div>
                <p class="text-[10px] text-zinc-500 leading-relaxed">
                  Evaluates incoming JSON completions. It uses optimized regex pattern classification to parse message arrays and block prompt overrides (like "ignore previous rules") or jailbreak attempts before they strike the LLM.
                </p>
              </div>

              <!-- Node 2 -->
              <div class="cyber-card rounded p-4 border border-zinc-900 bg-zinc-950/20 hover:border-zinc-800 transition-colors space-y-1">
                <div class="flex items-center justify-between font-push">
                  <span class="text-[10px] font-bold text-zinc-300 flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-red-500"></span>
                    Runtime Firewall (Middle Head)
                  </span>
                </div>
                <p class="text-[10px] text-zinc-500 leading-relaxed">
                  Monitors and intercepts tool argument signatures and command payloads. If an agent attempts to invoke hazardous functions (e.g. system deletions, database table drops, or formatting commands), the call is instantly halted.
                </p>
              </div>

              <!-- Node 3 -->
              <div class="cyber-card rounded p-4 border border-zinc-900 bg-zinc-950/20 hover:border-zinc-800 transition-colors space-y-1">
                <div class="flex items-center justify-between font-push">
                  <span class="text-[10px] font-bold text-zinc-300 flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-red-500"></span>
                    Egress Censor (Right Head)
                  </span>
                </div>
                <p class="text-[10px] text-zinc-500 leading-relaxed">
                  Sanitizes outbound model responses and streaming event chunks (SSE). It redacts PII or secrets (like API keys, JWTs, and phone numbers) on-the-fly, and blocks prompt leaks by matching rolling 8-word system prompt shingles.
                </p>
              </div>
            </div>

          </div>

        </div>

      </div>
    </section>
  </div>
</template>

<style scoped>
/* Keyframe animations for request flow tracer dot with trailing shadow motion blur */
@keyframes flow-allow {
  0% { 
    left: 10%; 
    background-color: #3b82f6; 
    box-shadow: -6px 0 8px rgba(59, 130, 246, 0.5), -12px 0 12px rgba(59, 130, 246, 0.25);
    opacity: 1; 
  }
  45% { 
    left: 50%; 
    background-color: #10b981; 
    box-shadow: -6px 0 8px rgba(16, 185, 129, 0.5), -12px 0 12px rgba(16, 185, 129, 0.25);
  }
  90% { 
    left: 90%; 
    background-color: #3b82f6; 
    box-shadow: -6px 0 8px rgba(59, 130, 246, 0.5), -12px 0 12px rgba(59, 130, 246, 0.25);
    opacity: 1; 
  }
  100% { 
    left: 90%; 
    opacity: 0; 
  }
}

@keyframes flow-block {
  0% { 
    left: 10%; 
    background-color: #3b82f6; 
    box-shadow: -6px 0 8px rgba(59, 130, 246, 0.5), -12px 0 12px rgba(59, 130, 246, 0.25);
    opacity: 1; 
  }
  45% { 
    left: 50%; 
    background-color: #ef4444; 
    box-shadow: -6px 0 8px rgba(239, 68, 68, 0.5), -12px 0 12px rgba(239, 68, 68, 0.25);
  }
  50% { left: 50%; opacity: 1; }
  60% { left: 50%; opacity: 0; } /* terminates at proxy core */
  100% { left: 50%; opacity: 0; }
}

@keyframes flow-flag {
  0% { 
    left: 10%; 
    background-color: #3b82f6; 
    box-shadow: -6px 0 8px rgba(59, 130, 246, 0.5), -12px 0 12px rgba(59, 130, 246, 0.25);
    opacity: 1; 
  }
  45% { 
    left: 50%; 
    background-color: #f59e0b; 
    box-shadow: -6px 0 8px rgba(245, 158, 11, 0.5), -12px 0 12px rgba(245, 158, 11, 0.25);
  }
  90% { 
    left: 90%; 
    background-color: #f59e0b; 
    box-shadow: -6px 0 8px rgba(245, 158, 11, 0.5), -12px 0 12px rgba(245, 158, 11, 0.25);
    opacity: 1; 
  }
  100% { 
    left: 90%; 
    opacity: 0; 
  }
}

.animate-allow {
  animation: flow-allow 3s infinite linear;
}

.animate-block {
  animation: flow-block 3s infinite linear;
}

.animate-flag {
  animation: flow-flag 3s infinite linear;
}

/* Slow breathing background pulse */
@keyframes pulse-slow {
  0%, 100% { opacity: 0.8; transform: scale(1); }
  50% { opacity: 1; transform: scale(1.04); }
}
.animate-pulse-slow {
  animation: pulse-slow 14s infinite ease-in-out;
}
</style>
