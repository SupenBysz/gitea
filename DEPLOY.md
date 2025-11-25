# Gitea 自动部署配置指南

本文档说明如何配置 Gitea 的自动部署功能，实现从 Release 发布到生产服务器的自动化部署。

## 概述

自动部署工作流会在以下情况触发：
- ✅ Release 发布成功后自动触发
- ✅ 手动触发（可指定版本号）

部署流程包括：
1. 📥 从 Release 下载 linux-amd64 产物
2. 💾 自动备份当前版本（保留最近 3 个备份）
3. ⏸️ 停止 Gitea 服务
4. 📦 替换二进制文件和资源文件
5. 🔐 设置文件权限（cap_net_bind_service）
6. ▶️ 启动 Gitea 服务
7. 🏥 健康检查（最多重试 30 次）
8. ⏪ 失败时自动回滚到上一版本

## 前置条件

### 1. 服务器要求

- **操作系统**: Linux (推荐 Ubuntu 20.04+ 或 CentOS 8+)
- **架构**: x86_64 (amd64)
- **服务管理**: systemd
- **用户**: gitea 用户（需要 sudo 权限）

### 2. 服务器配置

#### 2.1 创建 Gitea 用户（如果不存在）

```bash
# 创建 gitea 用户和组
sudo adduser --system --group --disabled-password --home /data/gitea gitea

# 创建必要的目录
sudo mkdir -p /data/gitea
sudo mkdir -p /data/gitea/backups
sudo chown -R gitea:gitea /data/gitea
```

#### 2.2 配置 systemd 服务

创建 `/etc/systemd/system/gitea.service`：

```ini
[Unit]
Description=Gitea (Git with a cup of tea)
After=network.target

[Service]
Type=simple
User=gitea
Group=gitea
WorkingDirectory=/data/gitea
ExecStart=/data/gitea/gitea web --config /data/gitea/custom/conf/app.ini
Restart=always
RestartSec=5s
Environment=USER=gitea HOME=/data/gitea GITEA_WORK_DIR=/data/gitea

# 安全加固
ProtectSystem=full
PrivateTmp=yes
NoNewPrivileges=true

[Install]
WantedBy=multi-user.target
```

重载 systemd 并启用服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable gitea
sudo systemctl start gitea
```

#### 2.3 配置 sudo 权限

为 gitea 用户配置免密码 sudo 权限（用于部署时重启服务）。

创建 `/etc/sudoers.d/gitea`：

```bash
# Gitea 用户允许执行的命令（无需密码）
gitea ALL=(ALL) NOPASSWD: /bin/systemctl start gitea
gitea ALL=(ALL) NOPASSWD: /bin/systemctl stop gitea
gitea ALL=(ALL) NOPASSWD: /bin/systemctl restart gitea
gitea ALL=(ALL) NOPASSWD: /bin/systemctl status gitea
gitea ALL=(ALL) NOPASSWD: /bin/cp
gitea ALL=(ALL) NOPASSWD: /bin/chown
gitea ALL=(ALL) NOPASSWD: /bin/chmod
gitea ALL=(ALL) NOPASSWD: /usr/sbin/setcap
```

设置正确的权限：

```bash
sudo chmod 440 /etc/sudoers.d/gitea
```

#### 2.4 配置 SSH 访问

**方式 1: SSH 密钥认证（推荐）**

在服务器上为 gitea 用户添加部署密钥：

```bash
# 在本地生成 SSH 密钥对（如果没有）
ssh-keygen -t ed25519 -C "gitea-deploy" -f ~/.ssh/gitea_deploy_key

# 将公钥添加到服务器
ssh-copy-id -i ~/.ssh/gitea_deploy_key.pub gitea@10.23.73.251

# 或者手动添加
sudo -u gitea mkdir -p /data/gitea/.ssh
sudo -u gitea touch /data/gitea/.ssh/authorized_keys
sudo chmod 700 /data/gitea/.ssh
sudo chmod 600 /data/gitea/.ssh/authorized_keys
# 将公钥内容追加到 authorized_keys
```

**方式 2: 密码认证**

如果不使用密钥，确保 gitea 用户设置了密码：

```bash
sudo passwd gitea
```

### 3. Gitea Repository 配置

#### 3.1 配置 Secrets

在 Gitea 仓库中配置以下 Secrets（Settings → Secrets → Actions）：

| Secret 名称 | 说明 | 是否必需 |
|------------|------|---------|
| `DEPLOY_SSH_KEY` | SSH 私钥（如使用密钥认证） | 推荐 |
| `DEPLOY_PASSWORD` | SSH 密码（如使用密码认证） | 可选 |

**配置 DEPLOY_SSH_KEY**：

1. 复制私钥内容：
   ```bash
   cat ~/.ssh/gitea_deploy_key
   ```

2. 在 Gitea 中添加 Secret：
   - Name: `DEPLOY_SSH_KEY`
   - Value: 粘贴私钥的完整内容（包括 `-----BEGIN OPENSSH PRIVATE KEY-----` 和 `-----END OPENSSH PRIVATE KEY-----`）

#### 3.2 调整环境变量（可选）

如果您的服务器配置不同，可以修改 `.gitea/workflows/deploy.yml` 中的环境变量：

```yaml
env:
  DEPLOY_HOST: '10.23.73.251'      # 服务器 IP 或域名
  DEPLOY_USER: 'gitea'             # SSH 用户名
  DEPLOY_PATH: '/data/gitea'       # Gitea 安装路径
  BACKUP_PATH: '/data/gitea/backups'  # 备份目录
  SERVICE_NAME: 'gitea'            # systemd 服务名称
  HEALTH_CHECK_URL: 'http://10.23.73.251:3000'  # 健康检查 URL
  MAX_BACKUPS: 3                   # 保留的备份数量
```

## 使用方法

### 自动部署

当 Release 发布成功后，自动部署工作流会自动触发：

1. 通过 CI/CD 创建 Release（`.gitea/workflows/release.yml`）
2. Release 发布成功后，自动触发部署工作流（`.gitea/workflows/deploy.yml`）
3. 工作流自动下载 linux-amd64 产物并部署到生产服务器
4. 部署成功后自动进行健康检查

### 手动部署

如需手动部署特定版本：

1. 进入 Gitea Actions 页面
2. 选择 `Deploy to Production` 工作流
3. 点击 `Run workflow`
4. 输入参数：
   - **version**: 要部署的版本号（如：`v1.25.2.1125-Kysion`）
   - **skip_backup**: 是否跳过备份（仅测试用，默认 false）
5. 点击 `Run workflow` 开始部署

### 查看部署日志

1. 进入 Gitea Actions 页面
2. 选择对应的部署工作流运行记录
3. 查看各个步骤的详细日志

## 部署流程详解

### 1. 下载产物

从 Gitea Release 下载对应版本的 linux-amd64 产物：

```bash
# 通过 Gitea API 获取 Release 信息
# 下载 gitea-vX.X.X.MMDD-Kysion-linux-amd64.tar.gz
# 解压产物
```

### 2. 备份当前版本

自动备份当前运行的版本：

```bash
/data/gitea/backups/
├── gitea-backup-20251125-143022/
│   ├── gitea              # 二进制文件
│   ├── custom/            # 自定义配置
│   └── VERSION            # 版本信息
├── gitea-backup-20251124-093015/
└── gitea-backup-20251123-101203/
```

备份策略：
- ✅ 备份二进制文件
- ✅ 备份 custom/ 目录（配置文件）
- ✅ 保留最近 3 个备份
- ✅ 自动清理旧备份

### 3. 停止服务

使用 systemd 停止 Gitea 服务：

```bash
sudo systemctl stop gitea
```

等待 2 秒确保服务完全停止。

### 4. 替换文件

替换二进制文件和资源文件：

```bash
# 替换二进制文件
sudo cp gitea /data/gitea/gitea
sudo chown gitea:gitea /data/gitea/gitea
sudo chmod +x /data/gitea/gitea

# 更新资源文件（如果包含）
sudo cp -r custom/* /data/gitea/custom/
sudo cp -r templates /data/gitea/
sudo cp -r public /data/gitea/
sudo cp -r options /data/gitea/
```

### 5. 设置权限

设置特殊权限允许非 root 用户绑定低端口（80/443）：

```bash
sudo setcap 'cap_net_bind_service=+ep' /data/gitea/gitea
```

### 6. 启动服务

使用 systemd 启动 Gitea 服务：

```bash
sudo systemctl start gitea
```

等待 5 秒让服务完全启动。

### 7. 健康检查

通过 HTTP 请求检查服务是否正常运行：

```bash
# 最多重试 30 次，每次间隔 2 秒
curl -f -s http://10.23.73.251:3000
```

健康检查通过条件：
- ✅ HTTP 请求返回成功（状态码 200-299）
- ✅ 服务正常响应

### 8. 自动回滚

如果部署失败（任何步骤出错或健康检查失败），自动回滚到上一版本：

1. 查找最新的备份
2. 停止当前服务
3. 恢复备份的二进制文件和配置
4. 重新设置权限
5. 启动服务
6. 验证回滚成功

## 安全注意事项

### 1. SSH 密钥管理

- ✅ 使用 ED25519 或 RSA 4096 密钥
- ✅ 密钥仅用于部署，不要用于其他目的
- ✅ 定期轮换密钥（建议每 6 个月）
- ✅ 私钥仅存储在 Gitea Secrets 中，不要提交到代码仓库

### 2. sudo 权限控制

- ✅ 仅授予必要的命令权限
- ✅ 使用 NOPASSWD 避免交互式密码输入
- ✅ 定期审计 sudoers 配置

### 3. 备份管理

- ✅ 定期测试备份恢复流程
- ✅ 考虑将备份同步到远程存储
- ✅ 定期检查备份目录磁盘空间

### 4. 网络安全

- ✅ 使用防火墙限制 SSH 访问源 IP
- ✅ 禁用 SSH 密码认证（仅允许密钥认证）
- ✅ 定期更新服务器安全补丁

## 故障排查

### 问题 1: SSH 连接失败

**症状**: 部署工作流报错 "Permission denied" 或 "Connection refused"

**解决方案**:
1. 检查 SSH 密钥配置是否正确
2. 确认服务器防火墙允许 SSH 连接
3. 验证 gitea 用户的 authorized_keys 权限
4. 测试 SSH 连接：`ssh -i deploy_key gitea@10.23.73.251`

### 问题 2: sudo 权限不足

**症状**: 部署脚本报错 "sudo: a password is required"

**解决方案**:
1. 检查 `/etc/sudoers.d/gitea` 文件是否存在
2. 确认文件权限为 440
3. 确认文件内容正确配置了 NOPASSWD
4. 测试 sudo 命令：`sudo -u gitea sudo systemctl status gitea`

### 问题 3: 健康检查失败

**症状**: 部署后健康检查超时或失败

**解决方案**:
1. 检查 Gitea 服务状态：`systemctl status gitea`
2. 查看 Gitea 日志：`journalctl -u gitea -n 100`
3. 确认健康检查 URL 是否正确
4. 手动访问健康检查 URL：`curl http://10.23.73.251:3000`

### 问题 4: 回滚失败

**症状**: 部署失败但回滚也失败

**解决方案**:
1. 手动连接到服务器
2. 检查备份目录：`ls -la /data/gitea/backups/`
3. 手动恢复最新备份
4. 重启服务：`sudo systemctl restart gitea`

### 问题 5: 磁盘空间不足

**症状**: 备份或部署失败，提示磁盘空间不足

**解决方案**:
1. 检查磁盘空间：`df -h /data/gitea`
2. 清理旧备份：手动删除或减少 MAX_BACKUPS 数量
3. 清理日志文件：`journalctl --vacuum-time=7d`

## 高级配置

### 配置部署通知

可以在 deploy.yml 中添加通知步骤，发送部署结果到邮件、Slack、钉钉等：

```yaml
- name: 发送部署通知
  if: always()
  run: |
    # 发送邮件通知
    # 或调用 Webhook
    # 或发送钉钉消息
```

### 配置蓝绿部署

如需实现蓝绿部署（零停机时间），需要：

1. 使用反向代理（如 Nginx）
2. 配置两套 Gitea 实例
3. 修改部署脚本实现实例切换

### 配置多服务器部署

如需部署到多台服务器，可以使用 matrix 策略：

```yaml
strategy:
  matrix:
    server:
      - host: 10.23.73.251
        name: production-1
      - host: 10.23.73.252
        name: production-2
```

## 参考资料

- [Gitea Actions 文档](https://docs.gitea.io/en-us/actions/)
- [systemd 服务管理](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
- [SSH 密钥认证配置](https://www.ssh.com/academy/ssh/public-key-authentication)
- [Linux Capabilities](https://man7.org/linux/man-pages/man7/capabilities.7.html)

## 更新记录

- 2025-11-25: 初始版本，支持自动部署到生产服务器
