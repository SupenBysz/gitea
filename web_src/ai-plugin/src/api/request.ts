import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse } from '@/types'

const BASE_URL = '/api/v1/ai-plugin'

const instance: AxiosInstance = axios.create({
  baseURL: BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
instance.interceptors.request.use(
  (config) => {
    // Get token from meta tag or localStorage
    const token = document.querySelector('meta[name="csrf-token"]')?.getAttribute('content')
      || localStorage.getItem('ai-plugin-token')
    if (token) {
      config.headers['X-CSRF-Token'] = token
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
instance.interceptors.response.use(
  (response: AxiosResponse<ApiResponse<unknown>>) => {
    const { data } = response
    if (data.code === 0) {
      return response
    }
    // Business error
    ElMessage.error(data.message || '请求失败')
    return Promise.reject(new Error(data.message))
  },
  (error) => {
    // Network or server error
    const status = error.response?.status
    let message = '网络错误，请稍后重试'

    if (status === 401) {
      message = '登录已过期，请重新登录'
      // Redirect to login
      window.location.href = '/user/login'
    } else if (status === 403) {
      message = '权限不足'
    } else if (status === 404) {
      message = '资源不存在'
    } else if (status === 500) {
      message = '服务器错误'
    } else if (error.message === 'Network Error') {
      message = '网络连接失败'
    } else if (error.code === 'ECONNABORTED') {
      message = '请求超时'
    }

    ElMessage.error(message)
    return Promise.reject(error)
  }
)

// Generic request methods
export async function get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  const response = await instance.get<ApiResponse<T>>(url, config)
  return response.data.data
}

export async function post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  const response = await instance.post<ApiResponse<T>>(url, data, config)
  return response.data.data
}

export async function put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  const response = await instance.put<ApiResponse<T>>(url, data, config)
  return response.data.data
}

export async function del<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
  const response = await instance.delete<ApiResponse<T>>(url, config)
  return response.data.data
}

export async function patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
  const response = await instance.patch<ApiResponse<T>>(url, data, config)
  return response.data.data
}

export default instance
