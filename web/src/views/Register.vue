<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/api'
import { Github, Eye, EyeOff } from 'lucide-vue-next'

const router = useRouter()
const auth = useAuthStore()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)
const showPassword = ref(false)

async function handleRegister() {
  error.value = ''

  if (password.value !== confirmPassword.value) {
    error.value = '两次输入的密码不一致'
    return
  }

  loading.value = true
  try {
    const res = await api.post('/auth/register', {
      username: username.value,
      password: password.value,
    })
    auth.setAuth(res.data.token, res.data.user)
    router.push('/')
  } catch (err: any) {
    error.value = err.response?.data?.error || '注册失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <div class="flex items-center justify-center gap-2 mb-2">
          <Github class="w-8 h-8" />
          <h1 class="text-2xl font-bold">GH Proxy Auth</h1>
        </div>
        <p class="text-sm text-[hsl(var(--muted-foreground))]">GitHub 代理加速服务</p>
      </div>

      <div class="bg-white rounded-lg border border-[hsl(var(--border))] shadow-sm p-6">
        <h2 class="text-lg font-semibold mb-4">注册</h2>

        <div v-if="error" class="mb-4 p-3 rounded-md bg-red-50 text-red-600 text-sm">{{ error }}</div>

        <form @submit.prevent="handleRegister" class="space-y-4">
          <div>
            <label class="block text-sm font-medium mb-1.5">用户名</label>
            <input v-model="username" type="text" required minlength="3" class="w-full px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))] focus:ring-offset-1" placeholder="至少3个字符" />
          </div>

          <div>
            <label class="block text-sm font-medium mb-1.5">密码</label>
            <div class="relative">
              <input v-model="password" :type="showPassword ? 'text' : 'password'" required minlength="6" class="w-full px-3 py-2 pr-10 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))] focus:ring-offset-1" placeholder="至少6个字符" />
              <button type="button" @click="showPassword = !showPassword" class="absolute right-3 top-1/2 -translate-y-1/2 text-[hsl(var(--muted-foreground))]">
                <Eye v-if="!showPassword" class="w-4 h-4" />
                <EyeOff v-else class="w-4 h-4" />
              </button>
            </div>
          </div>

          <div>
            <label class="block text-sm font-medium mb-1.5">确认密码</label>
            <input v-model="confirmPassword" :type="showPassword ? 'text' : 'password'" required class="w-full px-3 py-2 rounded-md border border-[hsl(var(--input))] bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-[hsl(var(--ring))] focus:ring-offset-1" placeholder="再次输入密码" />
          </div>

          <button type="submit" :disabled="loading" class="w-full py-2 px-4 rounded-md bg-[hsl(var(--primary))] text-[hsl(var(--primary-foreground))] text-sm font-medium hover:opacity-90 disabled:opacity-50 transition-opacity">
            {{ loading ? '注册中...' : '注册' }}
          </button>
        </form>

        <div class="mt-4 text-center">
          <router-link to="/login" class="text-sm text-[hsl(var(--primary))] hover:underline">已有账号？登录</router-link>
        </div>
      </div>
    </div>
  </div>
</template>
