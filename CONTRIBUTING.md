# 贡献指南

感谢你考虑为 Infinite Refill 做出贡献！

## 🤝 如何贡献

### 报告 Bug

如果你发现了 bug，请创建一个 issue，包含：

1. **清晰的标题**
2. **详细的描述**
3. **复现步骤**
4. **期望行为**
5. **实际行为**
6. **环境信息**（操作系统、Go 版本等）
7. **日志输出**（使用 `-v` 参数）

### 提出新功能

如果你有新功能的想法：

1. 先创建一个 issue 讨论
2. 说明功能的用途和价值
3. 提供使用场景示例
4. 等待维护者反馈

### 提交代码

1. **Fork 项目**
   ```bash
   git clone https://github.com/devocyACT/infinite-refill.git
   cd infinite-refill/refill
   ```

2. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **编写代码**
   - 遵循现有代码风格
   - 添加必要的注释
   - 保持中文日志消息
   - 编写单元测试

4. **测试代码**
   ```bash
   make test
   make build
   ./refill check
   ```

5. **提交更改**
   ```bash
   git add .
   git commit -m "feat: 添加新功能描述"
   ```

6. **推送分支**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **创建 Pull Request**
   - 清晰描述更改内容
   - 关联相关 issue
   - 等待代码审查

## 📝 代码规范

### Go 代码风格

- 使用 `gofmt` 格式化代码
- 遵循 [Effective Go](https://golang.org/doc/effective_go.html)
- 变量和函数使用驼峰命名
- 包名使用小写单词

### 日志消息

- 使用中文日志消息
- 保持简洁清晰
- 包含必要的上下文信息

示例：
```go
logger.Info("开始探测 %d 个账号（并发数=%d）", len(accounts), parallel)
logger.Error("加载账号失败：%v", err)
```

### 提交消息

使用约定式提交（Conventional Commits）：

- `feat:` 新功能
- `fix:` Bug 修复
- `docs:` 文档更新
- `style:` 代码格式（不影响功能）
- `refactor:` 重构
- `test:` 测试相关
- `chore:` 构建/工具相关

示例：
```
feat: 添加 Webhook 通知支持
fix: 修复 Worker Pool 超时问题
docs: 更新 Docker 使用文档
```

## 🧪 测试

### 运行测试

```bash
# 运行所有测试
make test

# 运行特定包的测试
go test -v ./internal/account/

# 生成覆盖率报告
make coverage
```

### 编写测试

- 为新功能编写单元测试
- 测试文件命名：`*_test.go`
- 测试函数命名：`TestXxx`
- 使用表驱动测试

示例：
```go
func TestAccount_IsExpired(t *testing.T) {
    tests := []struct {
        name     string
        modTime  time.Time
        days     int
        expected bool
    }{
        {
            name:     "未过期",
            modTime:  time.Now().AddDate(0, 0, -10),
            days:     30,
            expected: false,
        },
        // 更多测试案例...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

## 📚 文档

### 更新文档

如果你的更改影响了用户使用方式：

1. 更新 `README.md`
2. 更新相关的指南文档
3. 添加示例代码
4. 更新 `CHANGELOG.md`

### 文档风格

- 使用中文编写
- 保持简洁清晰
- 提供代码示例
- 包含实际用例

## 🔍 代码审查

Pull Request 会经过以下审查：

1. **代码质量**
   - 是否遵循代码规范
   - 是否有适当的错误处理
   - 是否有必要的注释

2. **功能完整性**
   - 是否实现了预期功能
   - 是否有边界情况处理
   - 是否有单元测试

3. **向后兼容性**
   - 是否破坏现有 API
   - 是否影响现有配置
   - 是否需要迁移指南

4. **文档完整性**
   - 是否更新了相关文档
   - 是否添加了使用示例
   - 是否更新了 CHANGELOG

## 🎯 开发环境

### 必需工具

- Go 1.22+
- Git
- Make
- Docker（可选）

### 推荐工具

- VS Code + Go 扩展
- golangci-lint（代码检查）
- delve（调试器）

### 设置开发环境

```bash
# 克隆项目
git clone https://github.com/devocyACT/infinite-refill.git
cd infinite-refill/refill

# 安装依赖
go mod download

# 构建
make build

# 运行测试
make test

# 本地运行
export SERVER_URL="https://your-server.com"
export USER_KEY="your-key"
export ACCOUNTS_DIR="./accounts"
./refill check
```

## 🐛 调试

### 启用详细日志

```bash
./refill -v run
```

### 使用调试器

```bash
dlv debug ./cmd/refill -- run
```

### 查看探测报告

```bash
cat out/probe_report_*.jsonl | jq .
```

## 📋 发布流程

（仅限维护者）

1. 更新版本号
2. 更新 CHANGELOG.md
3. 创建 Git tag
4. 构建多平台二进制
5. 构建 Docker 镜像
6. 发布 GitHub Release

## 💬 社区

- **Issues**: 报告 bug 和提出功能请求
- **Discussions**: 一般性讨论和问答
- **Pull Requests**: 代码贡献

## 📜 行为准则

- 尊重所有贡献者
- 保持友好和专业
- 接受建设性批评
- 关注项目目标

## ❓ 问题？

如果你有任何问题：

1. 查看现有文档
2. 搜索已有 issues
3. 创建新 issue 提问

感谢你的贡献！🎉
