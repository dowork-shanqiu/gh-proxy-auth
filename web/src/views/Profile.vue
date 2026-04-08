<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import { Key, Shield, Lock, Plus, Trash2, Eye, Clock, Infinity, Copy, Check, ChevronDown, ChevronUp } from 'lucide-vue-next'

// Profile data
const profile = ref<any>({})
const activeTab = ref('security')

// Password change
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordMsg = ref({ text: '', type: '' })
const passwordLoading = ref(false)

// TOTP
const totpSetup = ref<{ secret: string; qr_image: string } | null>(null)
const totpCode = ref('')
const totpMsg = ref({ text: '', type: '' })
const totpLoading = ref(false)

// Passkeys
const passkeys = ref<any[]>([])
const passkeyName = ref('')
const passkeyMsg = ref({ text: '', type: '' })
const passkeyLoading = ref(false)

// MFA Priority
const mfaPriority = ref('passkey')
const mfaMsg = ref({ text: '', type: '' })

// Tokens
const tokens = ref<any[]>([])
const showCreateToken = ref(false)
const newTokenName = ref('')
const newTokenExpireNum = ref(0)
const newTokenExpireUnit = ref('hour')
const tokenMsg = ref({ text: '', type: '' })
const tokenLoading = ref(false)
const expandedToken = ref<number | null>(null)
const tokenLogs = ref<any>({})

const copied = ref('')

onMounted(() => {
  fetchProfile()
  fetchTokens()
  fetchPasskeys()
})

async function fetchProfile() {
  try {
    const res = await api.get('/user/profile')
    profile.value = res.data
    mfaPriority.value = res.data.mfa_priority || 'passkey'
  } catch {
    // ignore
  }
}

async function fetchTokens() {
  try {
    const res = await api.get('/tokens')
    tokens.value = res.data || []
  } catch {
    // ignore
  }
}

async function fetchPasskeys() {
  try {
    const res = await api.get('/user/passkeys')
    passkeys.value = res.data || []
  } catch {
    // ignore
  }
}

// Password
async function changePassword() {
  passwordMsg.value = { text: '', type: '' }
  if (newPassword.value !== confirmPassword.value) {
    passwordMsg.value = { text: '两次输入的密码不一致', type: 'error' }
    return
  }
  passwordLoading.value = true
  try {
    await api.put('/user/password', {
      old_password: oldPassword.value,
      new_password: newPassword.value,
    })
    passwordMsg.value = { text: '密码修改成功', type: 'success' }
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
  } catch (err: any) {
    passwordMsg.value = { text: err.response?.data?.error || '修改失败', type: 'error' }
  } finally {
    passwordLoading.value = false
  }
}

// TOTP
async function setupTOTP() {
  totpMsg.value = { text: '', type: '' }
  totpLoading.value = true
  try {
    const res = await api.post('/user/totp/setup')
    totpSetup.value = res.data
  } catch (err: any) {
    totpMsg.value = { text: err.response?.data?.error || '设置失败', type: 'error' }
  } finally {
    totpLoading.value = false
  }
}

async function enableTOTP() {
  totpMsg.value = { text: '', type: '' }
  totpLoading.value = true
  try {
    await api.post('/user/totp/enable', { code: totpCode.value })
    totpMsg.value = { text: 'TOTP 已启用', type: 'success' }
    totpSetup.value = null
    totpCode.value = ''
    fetchProfile()
  } catch (err: any) {
    totpMsg.value = { text: err.response?.data?.error || '启用失败', type: 'error' }
  } finally {
    totpLoading.value = false
  }
}

async function disableTOTP() {
  if (!confirm('确定要关闭 TOTP 吗？')) return
  try {
    await api.delete('/user/totp')
    totpMsg.value = { text: 'TOTP 已关闭', type: 'success' }
    totpSetup.value = null
    fetchProfile()
  } catch (err: any) {
    totpMsg.value = { text: err.response?.data?.error || '关闭失败', type: 'error' }
  }
}

// Passkeys
async function registerPasskey() {
  passkeyMsg.value = { text: '', type: '' }
  passkeyLoading.value = true
  try {
    const beginRes = await api.post('/user/passkey/begin-register')
    const options = beginRes.data

    options.publicKey.challenge = base64URLToBuffer(options.publicKey.challenge)
    options.publicKey.user.id = base64URLToBuffer(options.publicKey.user.id)
    if (options.publicKey.excludeCredentials) {
      options.publicKey.excludeCredentials = options.publicKey.excludeCredentials.map((c: any) => ({
        ...c,
        id: base64URLToBuffer(c.id),
      }))
    }

    const credential = await navigator.credentials.create(options) as PublicKeyCredential
    if (!credential) {
      passkeyMsg.value = { text: '注册被取消', type: 'error' }
      passkeyLoading.value = false
      return
    }

    const response = credential.response as AuthenticatorAttestationResponse

    await api.post(`/user/passkey/finish-register?name=${encodeURIComponent(passkeyName.value || 'My Passkey')}`, {
      id: credential.id,
      rawId: bufferToBase64URL(credential.rawId),
      type: credential.type,
      response: {
        attestationObject: bufferToBase64URL(response.attestationObject),
        clientDataJSON: bufferToBase64URL(response.clientDataJSON),
      },
    })

    passkeyMsg.value = { text: 'Passkey 注册成功', type: 'success' }
    passkeyName.value = ''
    fetchPasskeys()
    fetchProfile()
  } catch (err: any) {
    passkeyMsg.value = { text: err.response?.data?.error || '注册失败', type: 'error' }
  } finally {
    passkeyLoading.value = false
  }
}

async function deletePasskey(id: number) {
  if (!confirm('确定要删除这个 Passkey 吗？')) return
  try {
    await api.delete(`/user/passkey/${id}`)
    fetchPasskeys()
    fetchProfile()
  } catch {
    // ignore
  }
}

// MFA Priority
async function updateMFAPriority() {
  mfaMsg.value = { text: '', type: '' }
  try {
    await api.put('/user/mfa-priority', { priority: mfaPriority.value })
    mfaMsg.value = { text: '验证优先级已更新', type: 'success' }
    fetchProfile()
  } catch (err: any) {
    mfaMsg.value = { text: err.response?.data?.error || '更新失败', type: 'error' }
  }
}

// Token
async function createToken() {
  tokenMsg.value = { text: '', type: '' }
  tokenLoading.value = true
  try {
    await api.post('/tokens', {
      name: newTokenName.value,
      expire_num: newTokenExpireNum.value,
      expire_unit: newTokenExpireUnit.value,
    })
    tokenMsg.value = { text: 'Token 创建成功', type: 'success' }
    showCreateToken.value = false
    newTokenName.value = ''
    newTokenExpireNum.value = 0
    newTokenExpireUnit.value = 'hour'
    fetchTokens()
  } catch (err: any) {
    tokenMsg.value = { text: err.response?.data?.error || '创建失败', type: 'error' }
  } finally {
    tokenLoading.value = false
  }
}

async function deleteToken(id: number) {
  if (!confirm('确定要删除这个 Token 吗？')) return
  try {
    await api.delete(`/tokens/${id}`)
    fetchTokens()
  } catch {
    // ignore
  }
}

async function toggleTokenLogs(id: number) {
  if (expandedToken.value === id) {
    expandedToken.value = null
    return
  }
  expandedToken.value = id
  if (!tokenLogs.value[id]) {
    try {
      const res = await api.get(`/tokens/${id}/logs`)
      tokenLogs.value[id] = res.data
    } catch {
      // ignore
    }
  }
}

async function copyToClipboard(text: string, id: string) {
  try {
    await navigator.clipboard.writeText(text)
  } catch {
    const el = document.createElement('textarea')
    el.value = text
    document.body.appendChild(el)
    el.select()
    document.execCommand('copy')
    document.body.removeChild(el)
  }
  copied.value = id
  setTimeout(() => (copied.value = ''), 2000)
}

function formatDate(dateStr: string) {
  if (!dateStr) return '永不过期'
  return new Date(dateStr).toLocaleString('zh-CN')
}

function isExpired(dateStr: string | null) {
  if (!dateStr) return false
  return new Date(dateStr) < new Date()
}

function base64URLToBuffer(base64url: string): ArrayBuffer {
  const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/')
  const pad = base64.length % 4
  const padded = pad ? base64 + '='.repeat(4 - pad) : base64
  const binary = atob(padded)
  const buffer = new ArrayBuffer(binary.length)
  const view = new Uint8Array(buffer)
  for (let i = 0; i < binary.length; i++) view[i] = binary.charCodeAt(i)
  return buffer
}

function bufferToBase64URL(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '')
}
</script>

<template>
  <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <h1 class="text-2xl font-bold mb-6">个人中心</h1>

    <!-- Tabs -->
    <div class="flex gap-1 mb-6 border-b border-[hsl(var(--border))]">
      <button @click="activeTab = 'security'" class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px" :class="activeTab === 'security' ? 'border-[hsl(var(--primary))] text-[hsl(var(--foreground))]' : 'border-transparent text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))]'">
        <Shield class="w-4 h-4 inline mr-1.5" />安全设置
      </button>
      <button @click="activeTab = 'tokens'" class="px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px" :class="activeTab === 'tokens' ? 'border-[hsl(var(--primary))] text-[hsl(var(--foreground))]' : 'border-transparent text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))]'">
        <Key class="w-4 h-4 inline mr-1.5" />Token 管理
      </button>
    </div>

    <!-- Security Tab -->
    <div v-if="activeTab === 'security'" class="space-y-6">
      <!-- Password -->
      <div class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-6">
        <h3 class="text-base font-semibold mb-4 flex items-center gap-2">
          <Lock class="w-4 h-4" /> 修改密码
        </h3>
        <div v-if="passwordMsg.text" class="mb-4 p-3 rounded-md text-sm" :class="passwordMsg.type === 'error' ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600'">{{ passwordMsg.text }}</div>
        <form @submit.prevent="changePassword" class="space-y-3 max-w-md">
          <div>
            <label class="block text-sm mb-1">当前密码</label>
            <input v-model="oldPassword" type="password" required class="w-full px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]" />
          </div>
          <div>
            <label class="block text-sm mb-1">新密码</label>
            <input v-model="newPassword" type="password" required minlength="6" class="w-full px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]" />
          </div>
          <div>
            <label class="block text-sm mb-1">确认新密码</label>
            <input v-model="confirmPassword" type="password" required class="w-full px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]" />
          </div>
          <button type="submit" :disabled="passwordLoading" class="px-4 py-2 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90 disabled:opacity-50">
            {{ passwordLoading ? '保存中...' : '修改密码' }}
          </button>
        </form>
      </div>

      <!-- TOTP -->
      <div class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-6">
        <h3 class="text-base font-semibold mb-4 flex items-center gap-2">
          <Shield class="w-4 h-4" /> TOTP 验证器
          <span v-if="profile.totp_enabled" class="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-700">已启用</span>
          <span v-else class="text-xs px-2 py-0.5 rounded-full bg-gray-100 text-gray-500">未启用</span>
        </h3>
        <div v-if="totpMsg.text" class="mb-4 p-3 rounded-md text-sm" :class="totpMsg.type === 'error' ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600'">{{ totpMsg.text }}</div>

        <template v-if="!profile.totp_enabled">
          <template v-if="!totpSetup">
            <p class="text-sm text-[hsl(var(--muted-foreground))] mb-3">启用 TOTP 可为您的账户增加额外的安全保护</p>
            <button @click="setupTOTP" :disabled="totpLoading" class="px-4 py-2 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90 disabled:opacity-50">
              设置 TOTP
            </button>
          </template>
          <template v-else>
            <div class="flex flex-col sm:flex-row gap-6">
              <div class="flex-shrink-0">
                <img :src="'data:image/png;base64,' + totpSetup.qr_image" alt="TOTP QR Code" class="w-48 h-48 border rounded-lg" />
              </div>
              <div class="flex-1 space-y-3">
                <p class="text-sm text-[hsl(var(--muted-foreground))]">使用验证器 App（如 Google Authenticator）扫描二维码，或手动输入密钥：</p>
                <div class="flex items-center gap-2">
                  <code class="bg-[hsl(var(--secondary))] px-3 py-1.5 rounded text-xs font-mono break-all">{{ totpSetup.secret }}</code>
                  <button @click="copyToClipboard(totpSetup.secret, 'totp-secret')" class="flex-shrink-0 p-1.5 rounded border hover:bg-[hsl(var(--secondary))]">
                    <Check v-if="copied === 'totp-secret'" class="w-3.5 h-3.5 text-green-500" />
                    <Copy v-else class="w-3.5 h-3.5" />
                  </button>
                </div>
                <form @submit.prevent="enableTOTP" class="flex gap-2">
                  <input v-model="totpCode" type="text" maxlength="6" required placeholder="输入验证码" class="w-32 px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm text-center tracking-widest focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]" />
                  <button type="submit" :disabled="totpLoading" class="px-4 py-2 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90 disabled:opacity-50">
                    启用
                  </button>
                </form>
              </div>
            </div>
          </template>
        </template>
        <template v-else>
          <p class="text-sm text-[hsl(var(--muted-foreground))] mb-3">TOTP 验证器已启用，登录时将需要额外验证</p>
          <button @click="disableTOTP" class="px-4 py-2 rounded-md border border-[hsl(var(--destructive))] text-[hsl(var(--destructive))] text-sm font-medium hover:bg-red-50">
            关闭 TOTP
          </button>
        </template>
      </div>

      <!-- Passkeys -->
      <div class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-6">
        <h3 class="text-base font-semibold mb-4 flex items-center gap-2">
          <Key class="w-4 h-4" /> Passkey 管理
          <span class="text-xs px-2 py-0.5 rounded-full bg-blue-100 text-blue-700">{{ passkeys.length }} 个</span>
        </h3>
        <div v-if="passkeyMsg.text" class="mb-4 p-3 rounded-md text-sm" :class="passkeyMsg.type === 'error' ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600'">{{ passkeyMsg.text }}</div>

        <div v-if="passkeys.length" class="space-y-2 mb-4">
          <div v-for="pk in passkeys" :key="pk.id" class="flex items-center justify-between p-3 rounded-md border border-[hsl(var(--border))] bg-[hsl(var(--secondary))]">
            <div>
              <span class="text-sm font-medium">{{ pk.name }}</span>
              <span class="text-xs text-[hsl(var(--muted-foreground))] ml-2">{{ formatDate(pk.created_at) }}</span>
            </div>
            <button @click="deletePasskey(pk.id)" class="p-1.5 rounded text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--destructive))] hover:bg-red-50">
              <Trash2 class="w-4 h-4" />
            </button>
          </div>
        </div>

        <div class="flex items-center gap-2">
          <input v-model="passkeyName" type="text" placeholder="Passkey 名称（可选）" class="w-48 px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]" />
          <button @click="registerPasskey" :disabled="passkeyLoading" class="px-4 py-2 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90 disabled:opacity-50">
            <Plus class="w-4 h-4 inline mr-1" />{{ passkeyLoading ? '注册中...' : '添加 Passkey' }}
          </button>
        </div>
      </div>

      <!-- MFA Priority -->
      <div v-if="profile.totp_enabled && passkeys.length > 0" class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-6">
        <h3 class="text-base font-semibold mb-4">验证优先级</h3>
        <div v-if="mfaMsg.text" class="mb-4 p-3 rounded-md text-sm" :class="mfaMsg.type === 'error' ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600'">{{ mfaMsg.text }}</div>
        <p class="text-sm text-[hsl(var(--muted-foreground))] mb-3">设置登录时的首选验证方式</p>
        <div class="flex items-center gap-3">
          <select v-model="mfaPriority" class="px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]">
            <option value="passkey">Passkey 优先</option>
            <option value="totp">TOTP 优先</option>
          </select>
          <button @click="updateMFAPriority" class="px-4 py-2 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90">
            保存
          </button>
        </div>
      </div>
    </div>

    <!-- Tokens Tab -->
    <div v-if="activeTab === 'tokens'" class="space-y-4">
      <div class="flex items-center justify-between">
        <p class="text-sm text-[hsl(var(--muted-foreground))]">管理用于 GitHub 代理下载的访问令牌</p>
        <button @click="showCreateToken = !showCreateToken" class="flex items-center gap-1.5 px-4 py-2 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90">
          <Plus class="w-4 h-4" /> 创建 Token
        </button>
      </div>

      <div v-if="tokenMsg.text" class="p-3 rounded-md text-sm" :class="tokenMsg.type === 'error' ? 'bg-red-50 text-red-600' : 'bg-green-50 text-green-600'">{{ tokenMsg.text }}</div>

      <!-- Create Token Form -->
      <div v-if="showCreateToken" class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-6">
        <h3 class="text-base font-semibold mb-4">创建新 Token</h3>
        <form @submit.prevent="createToken" class="space-y-3">
          <div>
            <label class="block text-sm mb-1">名称</label>
            <input v-model="newTokenName" type="text" placeholder="Token 名称（可选）" class="w-full max-w-md px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]" />
          </div>
          <div>
            <label class="block text-sm mb-1">有效期</label>
            <div class="flex items-center gap-2">
              <input v-model.number="newTokenExpireNum" type="number" min="0" class="w-24 px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]" />
              <select v-model="newTokenExpireUnit" class="px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))]">
                <option value="hour">小时</option>
                <option value="day">天</option>
              </select>
              <span class="text-xs text-[hsl(var(--muted-foreground))]">设为 0 表示永不过期</span>
            </div>
          </div>
          <div class="flex gap-2">
            <button type="submit" :disabled="tokenLoading" class="px-4 py-2 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90 disabled:opacity-50">
              {{ tokenLoading ? '创建中...' : '创建' }}
            </button>
            <button type="button" @click="showCreateToken = false" class="px-4 py-2 rounded-md border border-[hsl(var(--border))] text-sm hover:bg-[hsl(var(--secondary))]">
              取消
            </button>
          </div>
        </form>
      </div>

      <!-- Token List -->
      <div v-if="tokens.length === 0 && !showCreateToken" class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-8 text-center">
        <Key class="w-8 h-8 mx-auto mb-3 text-[hsl(var(--muted-foreground))]" />
        <p class="text-sm text-[hsl(var(--muted-foreground))]">暂无 Token，点击"创建 Token"生成新的访问令牌</p>
      </div>

      <div v-for="token in tokens" :key="token.id" class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm">
        <div class="p-4">
          <div class="flex items-center justify-between">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1">
                <span class="text-sm font-medium">{{ token.name }}</span>
                <span v-if="token.expires_at && isExpired(token.expires_at)" class="text-xs px-1.5 py-0.5 rounded-full bg-red-100 text-red-600">已过期</span>
                <span v-else-if="!token.expires_at" class="text-xs px-1.5 py-0.5 rounded-full bg-green-100 text-green-700 flex items-center gap-0.5"><Infinity class="w-3 h-3" /> 永不过期</span>
                <span v-else class="text-xs px-1.5 py-0.5 rounded-full bg-blue-100 text-blue-700 flex items-center gap-0.5"><Clock class="w-3 h-3" /> {{ formatDate(token.expires_at) }}</span>
              </div>
              <div class="flex items-center gap-2">
                <code class="text-xs font-mono bg-[hsl(var(--secondary))] px-2 py-1 rounded truncate max-w-xs">{{ token.token }}</code>
                <button @click="copyToClipboard(token.token, `token-${token.id}`)" class="flex-shrink-0 p-1 rounded border hover:bg-[hsl(var(--secondary))]">
                  <Check v-if="copied === `token-${token.id}`" class="w-3 h-3 text-green-500" />
                  <Copy v-else class="w-3 h-3" />
                </button>
              </div>
            </div>
            <div class="flex items-center gap-1 ml-3">
              <button @click="toggleTokenLogs(token.id)" class="p-1.5 rounded text-[hsl(var(--muted-foreground))] hover:bg-[hsl(var(--secondary))]" title="查看下载记录">
                <ChevronDown v-if="expandedToken !== token.id" class="w-4 h-4" />
                <ChevronUp v-else class="w-4 h-4" />
              </button>
              <button @click="deleteToken(token.id)" class="p-1.5 rounded text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--destructive))] hover:bg-red-50" title="删除">
                <Trash2 class="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <!-- Download Logs -->
        <div v-if="expandedToken === token.id" class="border-t border-[hsl(var(--border))] p-4">
          <h4 class="text-sm font-medium mb-2">下载记录</h4>
          <div v-if="!tokenLogs[token.id]?.data?.length" class="text-xs text-[hsl(var(--muted-foreground))]">暂无下载记录</div>
          <div v-else class="space-y-1.5">
            <div v-for="log in tokenLogs[token.id].data" :key="log.id" class="text-xs p-2 rounded bg-[hsl(var(--secondary))]">
              <div class="text-[hsl(var(--muted-foreground))]">{{ formatDate(log.created_at) }} · {{ log.ip }}</div>
              <div class="font-mono break-all mt-0.5">{{ log.url }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
