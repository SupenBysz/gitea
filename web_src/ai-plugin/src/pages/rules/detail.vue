<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useRulesStore } from '@/stores'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'

const route = useRoute()
const router = useRouter()
const rulesStore = useRulesStore()

const ruleId = computed(() => route.params.id as string)

const operatorLabels: Record<string, string> = {
  equals: '等于',
  not_equals: '不等于',
  contains: '包含',
  not_contains: '不包含',
  starts_with: '开头是',
  ends_with: '结尾是',
  matches: '正则匹配',
  greater_than: '大于',
  less_than: '小于',
  in: '在列表中',
  not_in: '不在列表中',
}

const actionLabels: Record<string, string> = {
  block: '阻止',
  warn: '警告',
  remind: '提醒',
  notify: '通知',
  auto_fix: '自动修复',
  llm_analyze: 'LLM 分析',
}

function formatTime(time: string) {
  return dayjs(time).format('YYYY-MM-DD HH:mm:ss')
}

function goBack() {
  router.push('/rules')
}

async function handleToggle() {
  if (!rulesStore.currentRule) return
  try {
    await rulesStore.toggleRuleEnabled(ruleId.value, !rulesStore.currentRule.enabled)
    ElMessage.success(`规则已${rulesStore.currentRule.enabled ? '禁用' : '启用'}`)
  } catch {
    // Error handled in store
  }
}

onMounted(async () => {
  await rulesStore.fetchRule(ruleId.value)
})
</script>

<template>
  <div class="rule-detail-page">
    <div class="page-header">
      <el-button type="text" @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回列表
      </el-button>
    </div>

    <el-skeleton :loading="rulesStore.loading" animated>
      <template #template>
        <div class="dashboard-card">
          <el-skeleton-item variant="h1" style="width: 30%" />
          <el-skeleton-item variant="text" style="width: 60%; margin-top: 16px" />
        </div>
      </template>
      <template #default>
        <div v-if="rulesStore.currentRule" class="rule-detail">
          <!-- Basic Info -->
          <div class="dashboard-card">
            <div class="dashboard-card__header">
              <div class="rule-title">
                <h2>{{ rulesStore.currentRule.name }}</h2>
                <el-tag :type="rulesStore.currentRule.enabled ? 'success' : 'info'" size="large">
                  {{ rulesStore.currentRule.enabled ? '已启用' : '已禁用' }}
                </el-tag>
              </div>
              <div>
                <el-button @click="handleToggle">
                  {{ rulesStore.currentRule.enabled ? '禁用规则' : '启用规则' }}
                </el-button>
                <el-button type="primary" disabled>编辑规则</el-button>
              </div>
            </div>
            <div class="dashboard-card__body">
              <el-descriptions :column="2" border>
                <el-descriptions-item label="规则ID">
                  {{ rulesStore.currentRule.id }}
                </el-descriptions-item>
                <el-descriptions-item label="优先级">
                  {{ rulesStore.currentRule.priority }}
                </el-descriptions-item>
                <el-descriptions-item label="描述" :span="2">
                  {{ rulesStore.currentRule.description }}
                </el-descriptions-item>
                <el-descriptions-item label="事件类型" :span="2">
                  <el-tag
                    v-for="type in rulesStore.currentRule.event_types"
                    :key="type"
                    style="margin-right: 8px"
                    :class="`event-type-tag ${type}`"
                  >
                    {{ type }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="创建时间">
                  {{ formatTime(rulesStore.currentRule.created_at) }}
                </el-descriptions-item>
                <el-descriptions-item label="更新时间">
                  {{ formatTime(rulesStore.currentRule.updated_at) }}
                </el-descriptions-item>
                <el-descriptions-item label="创建者">
                  {{ rulesStore.currentRule.created_by }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </div>

          <!-- Conditions -->
          <div class="dashboard-card mt-20">
            <div class="dashboard-card__header">
              <span class="title">触发条件</span>
            </div>
            <div class="dashboard-card__body">
              <div v-if="rulesStore.currentRule.conditions.length === 0" class="empty-hint">
                暂无触发条件
              </div>
              <el-timeline v-else>
                <el-timeline-item
                  v-for="(condition, index) in rulesStore.currentRule.conditions"
                  :key="index"
                  :type="index === 0 ? 'primary' : 'info'"
                >
                  <div class="condition-item">
                    <span v-if="index > 0" class="logic-label">
                      {{ condition.logic || 'AND' }}
                    </span>
                    <code class="field">{{ condition.field }}</code>
                    <span class="operator">{{ operatorLabels[condition.operator] || condition.operator }}</span>
                    <code class="value">{{ condition.value }}</code>
                  </div>
                </el-timeline-item>
              </el-timeline>
            </div>
          </div>

          <!-- Actions -->
          <div class="dashboard-card mt-20">
            <div class="dashboard-card__header">
              <span class="title">执行动作</span>
            </div>
            <div class="dashboard-card__body">
              <div v-if="rulesStore.currentRule.actions.length === 0" class="empty-hint">
                暂无执行动作
              </div>
              <el-table v-else :data="rulesStore.currentRule.actions" style="width: 100%">
                <el-table-column label="动作类型" width="120">
                  <template #default="{ row }">
                    <el-tag :type="row.type === 'block' ? 'danger' : row.type === 'warn' ? 'warning' : 'info'">
                      {{ actionLabels[row.type] || row.type }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="配置">
                  <template #default="{ row }">
                    <pre class="config-preview">{{ JSON.stringify(row.config, null, 2) }}</pre>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </div>

        <el-empty v-else description="规则不存在" />
      </template>
    </el-skeleton>
  </div>
</template>

<style scoped>
.rule-title {
  display: flex;
  align-items: center;
  gap: 12px;

  h2 {
    margin: 0;
    font-size: 20px;
  }
}

.empty-hint {
  color: #909399;
  text-align: center;
  padding: 20px;
}

.condition-item {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.logic-label {
  color: #e6a23c;
  font-weight: 600;
  margin-right: 8px;
}

.field {
  background-color: #ecf5ff;
  color: #409eff;
  padding: 2px 8px;
  border-radius: 4px;
}

.operator {
  color: #606266;
}

.value {
  background-color: #f0f9eb;
  color: #67c23a;
  padding: 2px 8px;
  border-radius: 4px;
}

.config-preview {
  background-color: #f5f7fa;
  padding: 8px 12px;
  border-radius: 4px;
  margin: 0;
  font-size: 12px;
  max-height: 100px;
  overflow: auto;
}
</style>
