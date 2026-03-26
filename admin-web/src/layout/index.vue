<template>
  <el-container class="layout">
    <!-- 侧边栏 -->
    <el-aside v-if="!isMobile" :width="isCollapsed ? '64px' : '220px'" class="sidebar">
      <div class="logo" @click="router.push('/dashboard')">
        <el-icon :size="28"><School /></el-icon>
        <span v-show="!isCollapsed" class="logo-text">知外助手</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        :collapse="isCollapsed"
        :collapse-transition="false"
        background-color="#0f172a"
        text-color="#94a3b8"
        active-text-color="#60a5fa"
        router
      >
        <template v-for="route in menuRoutes" :key="route.path">
          <el-menu-item :index="'/' + route.path">
            <el-icon><component :is="route.meta?.icon" /></el-icon>
            <template #title>{{ route.meta?.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-aside>

    <el-drawer
      v-model="mobileMenuVisible"
      :with-header="false"
      size="250px"
      direction="ltr"
      class="mobile-drawer"
    >
      <div class="logo mobile-logo" @click="handleMobileNavigate('/dashboard')">
        <el-icon :size="24"><School /></el-icon>
        <span class="logo-text">知外助手</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        background-color="#0f172a"
        text-color="#94a3b8"
        active-text-color="#60a5fa"
        router
      >
        <template v-for="route in menuRoutes" :key="route.path">
          <el-menu-item :index="'/' + route.path" @click="mobileMenuVisible = false">
            <el-icon><component :is="route.meta?.icon" /></el-icon>
            <template #title>{{ route.meta?.title }}</template>
          </el-menu-item>
        </template>
      </el-menu>
    </el-drawer>

    <!-- 右侧内容区 -->
    <el-container class="main-container">
      <!-- 顶栏 -->
      <el-header class="header">
        <div class="header-left">
          <el-icon class="collapse-btn" @click="handleMenuToggle">
            <Expand v-if="isCollapsed" />
            <Fold v-else />
          </el-icon>
          <el-breadcrumb v-if="!isMobile" separator="/">
            <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-if="currentRoute.meta?.title !== '仪表盘'">
              {{ currentRoute.meta?.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
          <div v-else class="mobile-title">
            {{ currentRoute.meta?.title || '仪表盘' }}
          </div>
        </div>
        <div class="header-right">
          <el-dropdown trigger="click" @command="handleCommand">
            <div class="user-info">
              <el-avatar :size="32" :src="authStore.user?.avatar">
                {{ authStore.user?.realName?.[0] || authStore.user?.username?.[0] || 'A' }}
              </el-avatar>
              <span class="user-name">{{ authStore.user?.realName || authStore.user?.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item disabled>
                  <el-tag size="small" :type="authStore.isSuperAdmin ? 'danger' : 'primary'">
                    {{ authStore.isSuperAdmin ? '超级管理员' : '管理员' }}
                  </el-tag>
                </el-dropdown-item>
                <el-dropdown-item divided command="logout">
                  <el-icon><SwitchButton /></el-icon>退出登录
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>

      <!-- 主内容 -->
      <el-main class="main-content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const isCollapsed = ref(false)
const isMobile = ref(false)
const mobileMenuVisible = ref(false)

const updateViewport = () => {
  isMobile.value = window.innerWidth < 992
  if (isMobile.value) {
    isCollapsed.value = true
  } else {
    mobileMenuVisible.value = false
  }
}

const currentRoute = computed(() => route)
const activeMenu = computed(() => {
  const path = route.path
  // 子路由高亮父菜单
  if (path.startsWith('/announcements')) return '/announcements'
  return path
})

const menuRoutes = computed(() => {
  const mainRoute = router.options.routes.find((r) => r.path === '/')
  return mainRoute?.children?.filter((r) => !r.meta?.hidden) || []
})

const handleMenuToggle = () => {
  if (isMobile.value) {
    mobileMenuVisible.value = !mobileMenuVisible.value
    return
  }
  isCollapsed.value = !isCollapsed.value
}

const handleMobileNavigate = (path: string) => {
  router.push(path)
  mobileMenuVisible.value = false
}

const handleCommand = (command: string) => {
  if (command === 'logout') {
    authStore.logout()
  }
}

onMounted(() => {
  updateViewport()
  window.addEventListener('resize', updateViewport)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', updateViewport)
})
</script>

<style lang="scss" scoped>
.layout {
  height: 100vh;
}

.sidebar {
  background-color: #0f172a;
  transition: width 0.3s;
  overflow-x: hidden;

  .el-menu {
    border-right: none;
  }
}

.logo {
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  color: #fff;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);

  .logo-text {
    font-size: 18px;
    font-weight: 700;
    white-space: nowrap;
  }
}

.main-container {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: #fff;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  padding: 0 20px;
  height: 60px;
  z-index: 10;
}

.mobile-title {
  font-size: 15px;
  font-weight: 600;
  color: #334155;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collapse-btn {
  font-size: 20px;
  cursor: pointer;
  color: #64748b;
  transition: color 0.2s;

  &:hover {
    color: #3b82f6;
  }
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 8px;
  transition: background-color 0.2s;

  &:hover {
    background-color: #f1f5f9;
  }
}

.user-name {
  font-size: 14px;
  color: #475569;
  font-weight: 500;
}

.main-content {
  background-color: #f1f5f9;
  overflow-y: auto;
  padding: 20px;
}

:deep(.mobile-drawer) {
  .el-drawer {
    background: #0f172a;
  }

  .el-drawer__body {
    padding: 0;
  }

  .el-menu {
    border-right: none;
  }
}

.mobile-logo {
  justify-content: flex-start;
  padding: 0 18px;
}

@media (max-width: 991px) {
  .header {
    padding: 0 12px;
    height: 56px;
  }

  .header-left {
    gap: 10px;
  }

  .user-name {
    display: none;
  }

  .main-content {
    padding: 12px;
  }
}
</style>
