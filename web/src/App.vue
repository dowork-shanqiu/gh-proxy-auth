<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'
import { computed } from 'vue'
import { LogOut, User, Shield, Home, Github } from 'lucide-vue-next'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const showNav = computed(() => {
  return !['login', 'register'].includes(route.name as string)
})

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="min-h-screen bg-[hsl(var(--background))]">
    <!-- Navigation -->
    <nav v-if="showNav && auth.isLoggedIn" class="border-b border-[hsl(var(--border))] bg-white/80 backdrop-blur-sm sticky top-0 z-50">
      <div class="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex items-center justify-between h-14">
          <div class="flex items-center gap-6">
            <router-link to="/" class="flex items-center gap-2 text-lg font-semibold text-[hsl(var(--foreground))]">
              <Github class="w-5 h-5" />
              GH Proxy
            </router-link>
            <div class="flex items-center gap-1">
              <router-link to="/" class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm transition-colors" :class="route.name === 'home' ? 'bg-[hsl(var(--secondary))] text-[hsl(var(--secondary-foreground))]' : 'text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))]'">
                <Home class="w-4 h-4" />
                首页
              </router-link>
              <router-link to="/profile" class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm transition-colors" :class="route.name === 'profile' ? 'bg-[hsl(var(--secondary))] text-[hsl(var(--secondary-foreground))]' : 'text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))]'">
                <User class="w-4 h-4" />
                个人中心
              </router-link>
              <router-link v-if="auth.isAdmin" to="/admin" class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm transition-colors" :class="route.name === 'admin' ? 'bg-[hsl(var(--secondary))] text-[hsl(var(--secondary-foreground))]' : 'text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--foreground))]'">
                <Shield class="w-4 h-4" />
                管理后台
              </router-link>
            </div>
          </div>
          <div class="flex items-center gap-3">
            <span class="text-sm text-[hsl(var(--muted-foreground))]">{{ auth.user?.username }}</span>
            <button @click="handleLogout" class="flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm text-[hsl(var(--muted-foreground))] hover:text-[hsl(var(--destructive))] transition-colors">
              <LogOut class="w-4 h-4" />
              退出
            </button>
          </div>
        </div>
      </div>
    </nav>

    <router-view />
  </div>
</template>
