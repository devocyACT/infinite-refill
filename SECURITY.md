# 安全政策

## 🔒 支持的版本

当前支持安全更新的版本：

| 版本 | 支持状态 |
| --- | --- |
| 1.0.x | ✅ 支持 |
| < 1.0 | ❌ 不支持 |

## 🐛 报告安全漏洞

我们非常重视安全问题。如果你发现了安全漏洞，请**不要**公开披露。

### 报告流程

1. **私密报告**
   - 发送邮件至：[security@example.com]（请替换为实际邮箱）
   - 或使用 GitHub Security Advisories（推荐）

2. **包含信息**
   - 漏洞描述
   - 影响范围
   - 复现步骤
   - 可能的修复方案（如果有）

3. **响应时间**
   - 我们会在 48 小时内确认收到
   - 在 7 天内提供初步评估
   - 根据严重程度制定修复计划

### 安全漏洞等级

- **严重**: 远程代码执行、权限提升
- **高**: 信息泄露、拒绝服务
- **中**: 配置问题、输入验证
- **低**: 信息披露、最佳实践

## 🛡️ 安全最佳实践

### 配置安全

1. **保护敏感信息**
   ```bash
   # 不要提交 .env 文件到版本控制
   # 使用环境变量或密钥管理服务
   ```

2. **文件权限**
   ```bash
   # 限制配置文件权限
   chmod 600 .env
   chmod 700 accounts/
   ```

3. **密钥管理**
   - 定期轮换 USER_KEY
   - 不要在日志中输出密钥
   - 使用强密钥

### 网络安全

1. **使用 HTTPS**
   - 确保 SERVER_URL 使用 HTTPS
   - 验证 SSL 证书

2. **代理配置**
   - 使用可信的代理服务器
   - 避免使用公共代理

3. **防火墙**
   - 限制出站连接
   - 只允许必要的端口

### Docker 安全

1. **镜像安全**
   ```bash
   # 使用官方基础镜像
   # 定期更新镜像
   docker pull infinite-refill:latest
   ```

2. **容器隔离**
   ```bash
   # 使用只读卷
   -v ./accounts:/data/accounts:ro

   # 限制资源
   --memory="256m" --cpus="0.5"
   ```

3. **网络隔离**
   ```yaml
   # docker-compose.yml
   networks:
     - refill-network
   ```

### 运行时安全

1. **最小权限原则**
   - 不要以 root 用户运行
   - 使用专用用户账号

2. **日志安全**
   - 定期清理日志文件
   - 不要记录敏感信息
   - 使用 `-v` 参数时注意日志内容

3. **监控**
   - 监控异常行为
   - 设置告警
   - 定期审查日志

## 🔐 已知安全考虑

### 账号文件

- 账号文件包含敏感的访问令牌
- 存储在本地文件系统
- 建议加密存储（未来功能）

### 网络通信

- 与 ChatGPT API 的通信使用 HTTPS
- 与 Topup 服务器的通信使用 HTTPS
- 支持代理配置

### 依赖安全

- 定期更新 Go 依赖
- 使用 `go mod tidy` 清理未使用的依赖
- 运行 `go list -m all` 检查依赖

## 📋 安全检查清单

部署前请确认：

- [ ] 已设置强密钥
- [ ] 配置文件权限正确
- [ ] 使用 HTTPS 连接
- [ ] 启用了防火墙
- [ ] 定期备份账号文件
- [ ] 监控系统已配置
- [ ] 日志轮转已设置
- [ ] Docker 容器使用非 root 用户
- [ ] 网络隔离已配置
- [ ] 定期更新软件版本

## 🔄 安全更新

我们会通过以下方式发布安全更新：

1. **GitHub Security Advisories**
2. **Release Notes**
3. **CHANGELOG.md**

订阅 GitHub Releases 以接收更新通知。

## 📚 相关资源

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Best Practices](https://golang.org/doc/security/)
- [Docker Security](https://docs.docker.com/engine/security/)

## 🙏 致谢

感谢所有负责任地披露安全问题的研究人员。

---

最后更新：2026-03-08
