<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterView } from 'vue-router'
import type { RouteLocationNormalizedLoaded } from 'vue-router'
import Toast from 'openvue/toast'
import ConfirmDialog from 'openvue/confirmdialog'
import AppHeader from '@/components/layout/AppHeader.vue'
import AppFooter from '@/components/layout/AppFooter.vue'
import { useThemeStore } from '@/stores/theme'

const theme = useThemeStore()

function viewKey(route: RouteLocationNormalizedLoaded): string {
  if (route.name === 'play' || route.name === 'game') return route.fullPath
  if (route.name === 'login' || route.name === 'register') return 'auth'

  return String(route.name)
}

onMounted(() => theme.init())
</script>

<template>
  <div class="flex min-h-dvh flex-col">
    <AppHeader />

    <main class="flex-1">
      <RouterView v-slot="{ Component, route }">
        <Transition name="page" mode="out-in">
          <component :is="Component" :key="viewKey(route)" />
        </Transition>
      </RouterView>
    </main>

    <AppFooter />

    <Toast position="bottom-right" />
    <ConfirmDialog />
  </div>
</template>

<style scoped>
.page-enter-active,
.page-leave-active {
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(7px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
