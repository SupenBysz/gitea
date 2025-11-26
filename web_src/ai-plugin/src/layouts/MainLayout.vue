<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  Odometer,
  List,
  Document,
  Setting,
  Fold,
  Expand,
  User,
  SwitchButton,
} from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const isCollapsed = ref(false)

const menuItems = [
  { path: '/dashboard', title: '仪表盘', icon: Odometer },
  { path: '/rules', title: '规则管理', icon: List },
  { path: '/events', title: '事件日志', icon: Document },
  { path: '/settings', title: '系统设置', icon: Setting },
]

const activeMenu = computed(() => route.path)

function toggleSidebar() {
  isCollapsed.value = !isCollapsed.value
}

function handleMenuSelect(path: string) {
  router.push(path)
}

function handleLogout() {
  window.location.href = '/user/logout'
}
</script>

<template>
  <div class="ai-plugin-layout">
    <!-- Sidebar -->
    <aside class="ai-plugin-layout__sidebar" :class="{ collapsed: isCollapsed }">
      <div class="sidebar-menu">
        <div class="logo">
          <el-icon class="logo-icon"><Odometer /></el-icon>
          <span v-show="!isCollapsed" class="logo-text">AI 协同规范</span>
        </div>
        <el-menu
          :default-active="activeMenu"
          :collapse="isCollapsed"
          :collapse-transition="false"
          background-color="#304156"
          text-color="#bfcbd9"
          active-text-color="#409EFF"
          @select="handleMenuSelect"
        >
          <el-menu-item
            v-for="item in menuItems"
            :key="item.path"
            :index="item.path"
          >
            <el-icon><component :is="item.icon" /></el-icon>
            <template #title>{{ item.title }}</template>
          </el-menu-item>
        </el-menu>
      </div>
    </aside>

    <!-- Main Content -->
    <div class="ai-plugin-layout__main">
      <!-- Header -->
      <header class="ai-plugin-layout__header">
        <el-icon
          class="collapse-btn"
          :size="20"
          style="cursor: pointer"
          @click="toggleSidebar"
        >
          <component :is="isCollapsed ? Expand : Fold" />
        </el-icon>

        <el-breadcrumb separator="/" style="margin-left: 20px">
          <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item v-if="route.meta.title">
            {{ route.meta.title }}
          </el-breadcrumb-item>
        </el-breadcrumb>

        <div style="flex: 1" />

        <el-dropdown trigger="click">
          <div class="user-info" style="display: flex; align-items: center; cursor: pointer">
            <el-avatar :size="32" :icon="User" />
            <span style="margin-left: 8px">管理员</span>
          </div>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleLogout">
                <el-icon><SwitchButton /></el-icon>
                退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </header>

      <!-- Page Content -->
      <main class="ai-plugin-layout__content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <keep-alive>
              <component :is="Component" />
            </keep-alive>
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
.collapse-btn:hover {
  color: var(--ai-plugin-primary);
}

.user-info:hover {
  color: var(--ai-plugin-primary);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
