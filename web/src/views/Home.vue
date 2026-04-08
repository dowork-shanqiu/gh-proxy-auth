<script setup lang="ts">
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { Copy, Check, Github, Download, Terminal, Link } from 'lucide-vue-next'

const auth = useAuthStore()

const inputUrl = ref('')
const copied = ref('')

const domain = computed(() => {
  return window.location.origin
})

const tokenPlaceholder = 'xxxxx'

const isValidGithubUrl = computed(() => {
  if (!inputUrl.value) return false
  const patterns = [
    /^(?:https?:\/\/)?github\.com\/.+?\/.+?\/(?:releases|archive)\/.*/i,
    /^(?:https?:\/\/)?github\.com\/.+?\/.+?\/(?:blob|raw)\/.*/i,
    /^(?:https?:\/\/)?github\.com\/.+?\/.+?\/(?:info|git-).*/i,
    /^(?:https?:\/\/)?raw\.(?:githubusercontent|github)\.com\/.+?\/.+?\/.+?\/.+/i,
    /^(?:https?:\/\/)?gist\.(?:githubusercontent|github)\.com\/.+?\/.+?\/.+/i,
    /^(?:https?:\/\/)?github\.com\/.+?\/.+?\/tags.*/i,
  ]
  return patterns.some(p => p.test(inputUrl.value))
})

const normalizedUrl = computed(() => {
  if (!inputUrl.value) return ''
  let url = inputUrl.value.trim()
  if (!url.startsWith('http://') && !url.startsWith('https://')) {
    url = 'https://' + url
  }
  return url
})

const proxyUrl = computed(() => {
  if (!isValidGithubUrl.value) return ''
  return `${domain.value}/${normalizedUrl.value}`
})

const commands = computed(() => {
  if (!proxyUrl.value) return []
  const url = proxyUrl.value

  // Check if it's a git clone URL
  const isGitUrl = /\/(?:info|git-)/.test(normalizedUrl.value)
  const isArchiveOrRelease = /\/(?:releases|archive)\//.test(normalizedUrl.value)

  const cmds = []

  if (isGitUrl) {
    // Extract the repo URL (before /info/ or /git-)
    const repoUrl = normalizedUrl.value.replace(/\/(?:info|git-).*$/, '')
    const proxyRepoUrl = `${domain.value}/${repoUrl}`
    cmds.push({
      label: 'Git Clone',
      icon: Terminal,
      command: `git -c http.extraHeader="X-XN-Token: ${tokenPlaceholder}" clone ${proxyRepoUrl}`,
    })
  }

  if (!isGitUrl) {
    cmds.push({
      label: 'curl 下载',
      icon: Download,
      command: `curl -H "X-XN-Token: ${tokenPlaceholder}" -L -O ${url}`,
    })
    cmds.push({
      label: 'wget 下载',
      icon: Download,
      command: `wget --header="X-XN-Token: ${tokenPlaceholder}" ${url}`,
    })
  }

  cmds.push({
    label: '代理链接',
    icon: Link,
    command: `${url}${url.includes('?') ? '&' : '?'}token=${tokenPlaceholder}`,
  })

  return cmds
})

async function copyToClipboard(text: string, id: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = id
    setTimeout(() => (copied.value = ''), 2000)
  } catch {
    // Fallback
    const el = document.createElement('textarea')
    el.value = text
    document.body.appendChild(el)
    el.select()
    document.execCommand('copy')
    document.body.removeChild(el)
    copied.value = id
    setTimeout(() => (copied.value = ''), 2000)
  }
}
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <!-- Hero -->
    <div class="text-center mb-10">
      <div class="flex items-center justify-center gap-3 mb-3">
        <Github class="w-10 h-10" />
        <h1 class="text-3xl font-bold">GitHub 代理加速</h1>
      </div>
      <p class="text-[hsl(var(--muted-foreground))]">输入 GitHub 文件链接，生成加速下载链接</p>
    </div>

    <!-- Input -->
    <div class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-6 mb-6">
      <label class="block text-sm font-medium mb-2">GitHub 链接</label>
      <div class="flex gap-3">
        <input
          v-model="inputUrl"
          type="text"
          class="flex-1 px-4 py-2.5 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))] focus:ring-offset-1"
          placeholder="输入 GitHub 文件链接，例如：https://github.com/user/repo/releases/download/v1.0/file.zip"
        />
      </div>

      <div class="mt-3 text-xs text-[hsl(var(--muted-foreground))]">
        支持: releases, archive, blob/raw, raw.githubusercontent.com, gist, tags, git clone
      </div>
    </div>

    <!-- Results -->
    <div v-if="inputUrl && !isValidGithubUrl" class="bg-white rounded-lg border border-orange-200 shadow-sm p-4 mb-6">
      <p class="text-sm text-orange-600">请输入有效的 GitHub 链接</p>
    </div>

    <div v-if="isValidGithubUrl" class="space-y-4">
      <div v-for="(cmd, idx) in commands" :key="idx" class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-4">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-2">
            <component :is="cmd.icon" class="w-4 h-4 text-[hsl(var(--muted-foreground))]" />
            <span class="text-sm font-medium">{{ cmd.label }}</span>
          </div>
          <button @click="copyToClipboard(cmd.command, `cmd-${idx}`)" class="flex items-center gap-1 px-2.5 py-1 rounded text-xs border border-[hsl(var(--border))] hover:bg-[hsl(var(--secondary))] transition-colors">
            <Check v-if="copied === `cmd-${idx}`" class="w-3 h-3 text-green-500" />
            <Copy v-else class="w-3 h-3" />
            {{ copied === `cmd-${idx}` ? '已复制' : '复制' }}
          </button>
        </div>
        <div class="bg-[hsl(var(--secondary))] rounded-md p-3 font-mono text-xs break-all leading-relaxed">
          {{ cmd.command }}
        </div>
      </div>

      <div class="bg-blue-50 rounded-lg border border-blue-200 p-4">
        <p class="text-sm text-blue-700">
          💡 提示: 将 <code class="bg-blue-100 px-1 rounded">{{ tokenPlaceholder }}</code> 替换为您在
          <router-link to="/profile" class="underline font-medium">个人中心</router-link>
          中生成的 Token。
        </p>
      </div>
    </div>
  </div>
</template>
