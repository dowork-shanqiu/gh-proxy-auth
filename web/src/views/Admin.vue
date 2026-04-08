<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import { Settings, Users, FileText, ChevronLeft, ChevronRight } from 'lucide-vue-next'

const activeTab = ref('settings')

// Settings
const openRegistration = ref(false)
const settingsMsg = ref({ text: '', type: '' })
const settingsLoading = ref(false)

// Users
const users = ref<any[]>([])
const usersPage = ref(1)
const usersTotal = ref(0)

// Logs
const logs = ref<any[]>([])
const logsPage = ref(1)
const logsTotal = ref(0)

onMounted(() => {
  fetchSettings()
  fetchUsers()
  fetchLogs()
})

async function fetchSettings() {
  try {
    const res = await api.get('/admin/settings')
    openRegistration.value = res.data.open_registration
  } catch {
    // ignore
  }
}

async function updateSettings() {
  settingsMsg.value = { text: '', type: '' }
  settingsLoading.value = true
  try {
    await api.put('/admin/settings', {
      open_registration: openRegistration.value,
    })
    settingsMsg.value = { text: '设置已保存', type: 'success' }
  } catch (err: any) {
    settingsMsg.value = { text: err.response?.data?.error || '保存失败', type: 'error' }
  } finally {
    settingsLoading.value = false
  }
}

async function fetchUsers(page = 1) {
  try {
    const res = await api.get('/admin/users', { params: { page, page_size: 20 } })
    users.value = res.data.data || []
    usersTotal.value = res.data.total
    usersPage.value = page
  } catch {
    // ignore
  }
}

async function fetchLogs(page = 1) {
  try {
    const res = await api.get('/admin/logs', { params: { page, page_size: 20 } })
    logs.value = res.data.data || []
    logsTotal.value = res.data.total
    logsPage.value = page
  } catch {
    // ignore
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}
</script>

<template>
  <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold mb-6">管理后台</h1>

    <!-- Tabs -->
    <div class="flex gap-1 mb-6 border-b border-[hsl(var(--border))]">
      <button @click="activeTab = 'settings'" class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px" :class="activeTab === 'settings' ? 'border-[hsl(var(--primary))] text-[hsl(var(--foreground))]' : 'border-transparent text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))]'">
        <Settings class="w-4 h-4 inline mr-1.5" />系统设置
      </button>
      <button @click="activeTab = 'users'; fetchUsers()" class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px" :class="activeTab === 'users' ? 'border-[hsl(var(--primary))] text-[hsl(var(--foreground))]' : 'border-transparent text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))]'">
        <Users class="w-4 h-4 inline mr-1.5" />用户管理
      </button>
      <button @click="activeTab = 'logs'; fetchLogs()" class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px" :class="activeTab === 'logs' ? 'border-[hsl(var(--primary))] text-[hsl(var(--foreground))]' : 'border-transparent text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))]'">
        <FileText class="w-4 h-4 inline mr-1.5" />下载记录
      </button>
    </div>

    <!-- Settings -->
    <div v-if="activeTab === 'settings'" class="space-y-6">
      <div class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-6">
        <h3 class="text-base font-semibold mb-4">注册设置</h3>
        <div v-if="settingsMsg.text" class="mb-4 p-3 rounded-md text-sm" :class="settingsMsg.type === 'error' ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600'">{{ settingsMsg.text }}</div>

        <div class="flex items-center justify-between max-w-md">
          <div>
            <div class="text-sm font-medium">开放注册</div>
            <div class="text-xs text-[hsl(var(--muted-foreground))]">允许新用户自行注册账号</div>
          </div>
          <label class="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" v-model="openRegistration" class="sr-only peer">
            <div class="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-[hsl(var(--ring))] rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[hsl(var(--primary))]"></div>
          </label>
        </div>

        <button @click="updateSettings" :disabled="settingsLoading" class="mt-4 px-4 py-2 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90 disabled:opacity-50">
          {{ settingsLoading ? '保存中...' : '保存设置' }}
        </button>
      </div>
    </div>

    <!-- Users -->
    <div v-if="activeTab === 'users'">
      <div class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm overflow-hidden">
        <table class="w-full text-sm">
          <thead class="bg-[hsl(var(--secondary))]">
            <tr>
              <th class="text-left px-4 py-3 font-medium">ID</th>
              <th class="text-left px-4 py-3 font-medium">用户名</th>
              <th class="text-left px-4 py-3 font-medium">角色</th>
              <th class="text-left px-4 py-3 font-medium">TOTP</th>
              <th class="text-left px-4 py-3 font-medium">注册时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[hsl(var(--border))]">
            <tr v-for="user in users" :key="user.id" class="hover:bg-[hsl(var(--secondary)_/_0.5)]">
              <td class="px-4 py-3">{{ user.id }}</td>
              <td class="px-4 py-3 font-medium">{{ user.username }}</td>
              <td class="px-4 py-3">
                <span v-if="user.is_admin" class="text-xs px-2 py-0.5 rounded-full bg-purple-100 text-purple-700">管理员</span>
                <span v-else class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">普通用户</span>
              </td>
              <td class="px-4 py-3">
                <span v-if="user.totp_enabled" class="text-xs text-green-600">✓</span>
                <span v-else class="text-xs text-gray-400">-</span>
              </td>
              <td class="px-4 py-3 text-[hsl(var(--muted-foreground))]">{{ formatDate(user.created_at) }}</td>
            </tr>
          </tbody>
        </table>
        <div v-if="usersTotal > 20" class="flex items-center justify-between px-4 py-3 border-t border-[hsl(var(--border))]">
          <span class="text-xs text-[hsl(var(--muted-foreground))]">共 {{ usersTotal }} 条</span>
          <div class="flex items-center gap-1">
            <button @click="fetchUsers(usersPage - 1)" :disabled="usersPage <= 1" class="p-1.5 rounded hover:bg-[hsl(var(--secondary))] disabled:opacity-30">
              <ChevronLeft class="w-4 h-4" />
            </button>
            <span class="text-xs px-2">{{ usersPage }}</span>
            <button @click="fetchUsers(usersPage + 1)" :disabled="usersPage * 20 >= usersTotal" class="p-1.5 rounded hover:bg-[hsl(var(--secondary))] disabled:opacity-30">
              <ChevronRight class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Download Logs -->
    <div v-if="activeTab === 'logs'">
      <div class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm overflow-hidden">
        <table class="w-full text-sm">
          <thead class="bg-[hsl(var(--secondary))]">
            <tr>
              <th class="text-left px-4 py-3 font-medium">时间</th>
              <th class="text-left px-4 py-3 font-medium">用户</th>
              <th class="text-left px-4 py-3 font-medium">Token</th>
              <th class="text-left px-4 py-3 font-medium">下载链接</th>
              <th class="text-left px-4 py-3 font-medium">IP</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[hsl(var(--border))]">
            <tr v-for="log in logs" :key="log.id" class="hover:bg-[hsl(var(--secondary)_/_0.5)]">
              <td class="px-4 py-3 whitespace-nowrap text-[hsl(var(--muted-foreground))]">{{ log.created_at }}</td>
              <td class="px-4 py-3 font-medium">{{ log.username }}</td>
              <td class="px-4 py-3">
                <span class="text-xs font-mono bg-[hsl(var(--secondary))] px-1.5 py-0.5 rounded">{{ log.token_name }}</span>
              </td>
              <td class="px-4 py-3 max-w-xs truncate font-mono text-xs">{{ log.url }}</td>
              <td class="px-4 py-3 text-[hsl(var(--muted-foreground))]">{{ log.ip }}</td>
            </tr>
          </tbody>
        </table>
        <div v-if="logs.length === 0" class="p-8 text-center text-sm text-[hsl(var(--muted-foreground))]">暂无下载记录</div>
        <div v-if="logsTotal > 20" class="flex items-center justify-between px-4 py-3 border-t border-[hsl(var(--border))]">
          <span class="text-xs text-[hsl(var(--muted-foreground))]">共 {{ logsTotal }} 条</span>
          <div class="flex items-center gap-1">
            <button @click="fetchLogs(logsPage - 1)" :disabled="logsPage <= 1" class="p-1.5 rounded hover:bg-[hsl(var(--secondary))] disabled:opacity-30">
              <ChevronLeft class="w-4 h-4" />
            </button>
            <span class="text-xs px-2">{{ logsPage }}</span>
            <button @click="fetchLogs(logsPage + 1)" :disabled="logsPage * 20 >= logsTotal" class="p-1.5 rounded hover:bg-[hsl(var(--secondary))] disabled:opacity-30">
              <ChevronRight class="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
