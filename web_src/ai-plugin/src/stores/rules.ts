import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Rule, PaginatedResponse } from '@/types'
import {
  getRuleList,
  getRule,
  createRule,
  updateRule,
  deleteRule,
  toggleRule,
  type RuleListParams,
  type CreateRuleData,
} from '@/api'

export const useRulesStore = defineStore('rules', () => {
  // State
  const rules = ref<Rule[]>([])
  const currentRule = ref<Rule | null>(null)
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  async function fetchRules(params?: RuleListParams) {
    loading.value = true
    error.value = null
    try {
      const response: PaginatedResponse<Rule> = await getRuleList({
        page: page.value,
        page_size: pageSize.value,
        ...params,
      })
      rules.value = response.items
      total.value = response.total
    } catch (e) {
      error.value = '获取规则列表失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchRule(id: string) {
    loading.value = true
    error.value = null
    try {
      currentRule.value = await getRule(id)
    } catch (e) {
      error.value = '获取规则详情失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function addRule(data: CreateRuleData) {
    loading.value = true
    error.value = null
    try {
      const newRule = await createRule(data)
      rules.value.unshift(newRule)
      total.value++
      return newRule
    } catch (e) {
      error.value = '创建规则失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function editRule(id: string, data: Partial<CreateRuleData>) {
    loading.value = true
    error.value = null
    try {
      const updatedRule = await updateRule(id, data)
      const index = rules.value.findIndex((r) => r.id === id)
      if (index !== -1) {
        rules.value[index] = updatedRule
      }
      if (currentRule.value?.id === id) {
        currentRule.value = updatedRule
      }
      return updatedRule
    } catch (e) {
      error.value = '更新规则失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function removeRule(id: string) {
    loading.value = true
    error.value = null
    try {
      await deleteRule(id)
      rules.value = rules.value.filter((r) => r.id !== id)
      total.value--
      if (currentRule.value?.id === id) {
        currentRule.value = null
      }
    } catch (e) {
      error.value = '删除规则失败'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function toggleRuleEnabled(id: string, enabled: boolean) {
    error.value = null
    try {
      const updatedRule = await toggleRule(id, enabled)
      const index = rules.value.findIndex((r) => r.id === id)
      if (index !== -1) {
        rules.value[index] = updatedRule
      }
      if (currentRule.value?.id === id) {
        currentRule.value = updatedRule
      }
      return updatedRule
    } catch (e) {
      error.value = '切换规则状态失败'
      throw e
    }
  }

  function setPage(p: number) {
    page.value = p
  }

  function setPageSize(size: number) {
    pageSize.value = size
  }

  function clearCurrentRule() {
    currentRule.value = null
  }

  return {
    // State
    rules,
    currentRule,
    total,
    page,
    pageSize,
    loading,
    error,
    // Actions
    fetchRules,
    fetchRule,
    addRule,
    editRule,
    removeRule,
    toggleRuleEnabled,
    setPage,
    setPageSize,
    clearCurrentRule,
  }
})
