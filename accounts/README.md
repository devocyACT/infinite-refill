# 账号文件目录

此目录用于存储账号 JSON 文件。

## ⚠️ 重要提示

- **不要提交账号文件到版本控制**
- 账号文件包含敏感的访问令牌
- 此目录已在 `.gitignore` 中排除

## 📁 文件格式

账号文件为 JSON 格式，示例：

```json
{
  "type": "codex",
  "access_token": "Bearer_token_here",
  "account_id": "user-xxx",
  "email": "example@example.com"
}
```

## 🔒 安全建议

1. **设置正确的文件权限**
   ```bash
   chmod 700 accounts/
   chmod 600 accounts/*.json
   ```

2. **定期备份**
   ```bash
   tar -czf accounts-backup-$(date +%Y%m%d).tar.gz accounts/
   ```

3. **加密存储**（推荐）
   ```bash
   # 使用 gpg 加密备份
   gpg -c accounts-backup.tar.gz
   ```

## 📊 文件管理

- 文件名通常为 MD5 哈希值或 account_id
- 程序会自动管理文件的创建和删除
- 清理操作会先备份到 `out/清理-*/backup/`

## 🚫 不要做的事

- ❌ 不要手动编辑账号文件
- ❌ 不要分享账号文件
- ❌ 不要提交到 Git
- ❌ 不要上传到公共服务器

## ✅ 可以做的事

- ✅ 定期备份
- ✅ 监控文件数量
- ✅ 查看文件修改时间
- ✅ 使用程序提供的清理功能

## 📝 查看账号统计

```bash
# 查看账号数量
ls -1 accounts/*.json | wc -l

# 查看最近修改的账号
ls -lt accounts/*.json | head -5

# 查看账号文件大小
du -sh accounts/
```

---

**注意**: 此 README 文件会被提交到版本控制，但账号 JSON 文件不会。
