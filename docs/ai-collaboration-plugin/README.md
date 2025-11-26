# AI 协同工作规范插件 - 技术设计文档

## 项目概述

AI 协同工作规范插件是为 Gitea 平台开发的自动化合规检查与智能分析系统，旨在确保 AI 员工在代码协作过程中遵循既定的工作规范。

## 文档目录

### 1. 技术架构设计

📁 [architecture/README.md](./architecture/README.md)

- 系统架构概览
- 技术栈选型
- 部署架构设计
- 核心模块设计
  - 规则引擎 (Rule Engine)
  - 事件监听器 (Event Listener)
  - LLM 网关 (LLM Gateway)
  - 智能分析模块 (Smart Analysis)
  - 通知模块 (Notification)
  - 自动化模块 (Automation)
- 数据流设计
- 安全设计
- 性能设计
- 可观测性设计
- 扩展性设计
- 版本规划

### 2. 模块接口规范

📁 [interfaces/README.md](./interfaces/README.md)

- 规则引擎接口 (RuleEngine)
- 事件监听器接口 (EventListener)
- LLM 网关接口 (LLMGateway)
- 智能分析接口 (AnalysisService)
- 通知模块接口 (NotificationService)
- 自动化模块接口 (AutomationService)
- 审计日志接口 (AuditService)
- 缓存管理接口 (CacheService)
- 错误处理规范
- 使用示例

### 3. 数据库模型设计

📁 [database/README.md](./database/README.md)

- ER 图
- 表结构定义
  - 规则管理表
  - 事件处理表
  - 通知管理表
  - LLM 管理表
  - 定时任务表
  - 用户与权限表
  - 审计日志表
  - 时序数据表 (TimescaleDB)
- Go 模型定义 (GORM)
- 数据迁移方案
- 索引策略
- 数据保留策略

### 4. API 接口文档

📁 [api/README.md](./api/README.md)

- 规则管理 API
- 事件管理 API
- 评估结果 API
- LLM 网关 API
- 智能分析 API
- 通知管理 API
- 定时任务 API
- 审计日志 API
- 仪表盘 API
- 用户管理 API
- 系统配置 API
- 错误码参考

## 核心功能

### 规则引擎

- 支持 YAML 格式的规则配置
- 灵活的条件表达式（AND/OR 组合）
- 多种动作类型：block、warn、remind、notify、auto_fix
- 规则优先级管理
- 规则验证与测试

### 事件监听

- 支持所有 Gitea Webhook 事件类型
- Webhook 签名验证
- 异步事件处理
- 事件重放功能

### LLM 网关

- 多模型支持（Claude、GPT 等）
- 智能路由与负载均衡
- 流式响应支持
- 使用量统计与配额管理
- 故障转移机制

### 智能分析

- 代码审查分析
- 评论质量分析
- 依赖安全检测
- PR 摘要生成
- 冲突分析建议

### 通知系统

- 多渠道支持：Telegram、Gitea 评论、Webhook、Email、Slack
- 模板引擎（Go template）
- 通知历史记录
- 重试机制

### 仪表盘

- 实时合规率展示
- 事件流监控
- 统计趋势图表
- WebSocket 实时更新

## 技术栈

| 分类 | 技术 |
|------|------|
| 后端语言 | Go 1.21+ |
| Web 框架 | Gin |
| ORM | GORM |
| 前端框架 | Vue 3 + TypeScript + Vite |
| UI 组件 | Element Plus |
| 状态管理 | Pinia |
| 主数据库 | PostgreSQL 15+ |
| 时序数据库 | TimescaleDB |
| 缓存 | Redis 7+ |
| 文件存储 | MinIO |
| 消息队列 | Redis Streams |
| 配置格式 | YAML |

## 版本计划

### MVP (v0.1.0)
- [x] 基础 Webhook 事件监听
- [x] 简单规则引擎（YAML 配置）
- [x] Telegram 通知
- [x] 基础仪表盘

### v0.2.0
- [ ] LLM 网关（Claude 支持）
- [ ] 智能代码审查
- [ ] 评论质量分析
- [ ] 多渠道通知

### v1.0.0
- [ ] 完整规则引擎
- [ ] 多 LLM 支持
- [ ] 自动化工作流
- [ ] 管理后台
- [ ] 完整审计日志

## 相关工单

- [Issue #14](https://gitea.ktyun.cc/Kysion/entai-gitea-custom/issues/14) - AI 协同工作规范插件技术架构设计

## 维护信息

- **文档版本**: 1.0.0
- **最后更新**: 2025-01-15
- **作者**: ArchitectAi
- **分支**: `feat/issue-14-architecture-design`
