<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useRulesStore } from '@/stores'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Rule, EventType } from '@/types'
import dayjs from 'dayjs'

const router = useRouter()
const rulesStore = useRulesStore()

const searchKeyword = ref('')
const filterEnabled = ref<boolean | ''>('')
const filterEventType = ref<EventType | ''>('')

const eventTypeOptions: { label: string; value: EventType | '' }[] = [
  { label: '全部类型', value: '' },
  { label: 'Push', value: 'push' },
  { label: 'Pull Request', value: 'pull_request' },
  { label: 'Issues', value: 'issues' },
  { label: 'Issue Comment', value: 'issue_comment' },
  { label: 'PR Review', value: 'pull_request_review' },
]

const filteredRules = computed(() => {
  return rulesStore.rules.filter((rule) => {
    if (searchKeyword.value && !rule.name.includes(searchKeyword.value) && !rule.description.includes(searchKeyword.value)) {
      return false
    }
    if (filterEnabled.value !== '' && rule.enabled !== filterEnabled.value) {
      return false
    }
    if (filterEventType.value && !rule.event_types.includes(filterEventType.value)) {
      return false
    }
    return true
  })
})

async function handleSearch() {
  await rulesStore.fetchRules({
    keyword: searchKeyword.value,
    enabled: filterEnabled.value === '' ? undefined : filterEnabled.value,
    event_type: filterEventType.value || undefined,
  })
}

async function handleToggle(rule: Rule) {
  try {
    await rulesStore.toggleRuleEnabled(rule.id, !rule.enabled)
    ElMessage.success(`规则已${rule.enabled ? '禁用' : '启用'}`)
  } catch {
    // Error already handled in store
  }
}

async function handleDelete(rule: Rule) {
  try {
    await ElMessageBox.confirm(
      `确定要删除规则「${rule.name}」吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    await rulesStore.removeRule(rule.id)
    ElMessage.success('规则已删除')
  } catch (e) {
    if (e !== 'cancel') {
      // Error already handled in store
    }
  }
}

function handleEdit(rule: Rule) {
  router.push(`/rules/${rule.id}`)
}

function handlePageChange(page: number) {
  rulesStore.setPage(page)
  rulesStore.fetchRules()
}

function handleSizeChange(size: number) {
  rulesStore.setPageSize(size)
  rulesStore.fetchRules()
}

function formatTime(time: string) {
  return dayjs(time).format('YYYY-MM-DD HH:mm')
}

function getActionTypes(rule: Rule) {
  return rule.actions.map((a) => a.type).join(', ')
}

onMounted(() => {
  rulesStore.fetchRules()
})
</script>

<template>
  <div class="rules-page">
    <div class="page-header">
      <h1 class="page-header__title">规则管理</h1>
      <p class="page-header__desc">管理 AI 协同工作的合规检查规则</p>
    </div>

    <!-- Filter Bar -->
    <div class="filter-bar">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索规则名称或描述"
        style="width: 240px"
        clearable
        @keyup.enter="handleSearch"
      >
        <template #prefix>
          <el-icon><Search /></el-icon>
        </template>
      </el-input>
      <el-select v-model="filterEnabled" placeholder="启用状态" style="width: 120px" clearable>
        <el-option label="全部" :value="''" />
        <el-option label="已启用" :value="true" />
        <el-option label="已禁用" :value="false" />
      </el-select>
      <el-select v-model="filterEventType" placeholder="事件类型" style="width: 160px" clearable>
        <el-option
          v-for="opt in eventTypeOptions"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-button type="primary" @click="handleSearch">
        <el-icon><Search /></el-icon>
        搜索
      </el-button>
      <div style="flex: 1" />
      <el-button type="primary" disabled>
        <el-icon><Plus /></el-icon>
        新建规则
      </el-button>
    </div>

    <!-- Rules Table -->
    <div class="dashboard-card mt-20">
      <el-table
        v-loading="rulesStore.loading"
        :data="filteredRules"
        style="width: 100%"
        row-key="id"
      >
        <el-table-column label="规则名称" min-width="180">
          <template #default="{ row }">
            <div class="rule-name">
              <span class="name">{{ row.name }}</span>
              <span class="priority">优先级: {{ row.priority }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="描述" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.description }}
          </template>
        </el-table-column>
        <el-table-column label="事件类型" width="180">
          <template #default="{ row }">
            <div class="event-types">
              <el-tag
                v-for="type in row.event_types.slice(0, 2)"
                :key="type"
                size="small"
                :class="`event-type-tag ${type}`"
                style="margin-right: 4px"
              >
                {{ type }}
              </el-tag>
              <el-tag v-if="row.event_types.length > 2" size="small" type="info">
                +{{ row.event_types.length - 2 }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="动作" width="120">
          <template #default="{ row }">
            <el-tooltip :content="getActionTypes(row)" placement="top">
              <span>{{ row.actions.length }} 个动作</span>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              :model-value="row.enabled"
              @change="handleToggle(row)"
              inline-prompt
              active-text="启用"
              inactive-text="禁用"
            />
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template #default="{ row }">
            {{ formatTime(row.updated_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- Pagination -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="rulesStore.page"
          v-model:page-size="rulesStore.pageSize"
          :total="rulesStore.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handlePageChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.filter-bar {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 16px 20px;
  background-color: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
}

.rule-name {
  display: flex;
  flex-direction: column;

  .name {
    font-weight: 600;
    color: #303133;
  }

  .priority {
    font-size: 12px;
    color: #909399;
    margin-top: 2px;
  }
}

.event-types {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.pagination-wrapper {
  display: flex;
  justify-content: flex-end;
  padding: 16px 0;
}
</style>
