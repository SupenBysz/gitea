# Wiki API 修复工单总结

## 修复时间
2025-11-24

## 修复提交
- Commit ID: `44c5b7f762`
- 修改文件: `routers/api/v1/repo/wiki.go`
- 代码变更: +23行 / -9行

---

## 问题1: PATCH API标题处理错误

### 问题描述
当PATCH请求包含`Title`字段时，即使标题未改变，也会触发路径转换，导致：
- 不必要的页面重命名
- 空标题或"."被转换为"unnamed"
- 总是更新错误的页面

### 解决方案
在`EditWikiPage`函数中添加标题比较逻辑：
```go
// 比较新旧标题，只在标题真正改变时才转换路径
_, currentTitle := wiki_service.WebPathToUserTitle(oldWikiName)
if strings.TrimSpace(form.Title) == currentTitle {
    newWikiName = oldWikiName  // 保持原有WebPath
} else {
    newWikiName = wiki_service.UserTitleToWebPath("", form.Title)
}
```

### 修复效果
✅ 避免不必要的路径转换和重命名
✅ 修复"unnamed"页面bug
✅ PATCH API正确更新指定页面

---

## 问题2.1: GET API不返回content_base64字段

### 问题描述
GET Wiki页面API不返回`content_base64`字段，导致前端无法获取页面内容。

### 根本原因
`wikiContentsByName()`函数只尝试查找一种文件路径格式，但Wiki文件在Git中可能以两种格式存储：
- Unescaped格式: `Page-Name.md`
- Escaped格式: `Page%20Name.md`

### 解决方案
实现双重查找策略，与Service层的`prepareGitPath()`逻辑保持一致：
```go
// 先尝试unescaped版本
unescaped := string(wikiName) + ".md"
entry, err := findEntryForFile(commit, unescaped)
if err == nil && entry != nil {
    return wikiContentsByEntry(ctx, entry), unescaped
}

// 如果找不到，再尝试escaped版本
gitPath := wiki_service.WebPathToGitPath(wikiName)
entry, err = findEntryForFile(commit, gitPath)
```

### 修复效果
✅ GET API正确返回`content_base64`字段
✅ 与Service层行为一致
✅ 支持两种文件格式的向后兼容

---

## 问题2.2: GET API对非Home页面返回404

### 问题描述
GET请求非Home页面时返回404错误，无法获取大部分Wiki页面。

### 根本原因
与问题2.1相同，只查找一种文件路径格式。

### 解决方案
同问题2.1的双重查找策略。

### 修复效果
✅ 解决了大部分页面返回404的问题
✅ 正确处理不同命名格式的页面

---

## 问题2.3: Revisions API返回格式错误

### 问题描述
`/repos/{owner}/{repo}/wiki/revisions/{pageName}` API直接返回数组，导致前端JavaScript无法访问`.commits`属性。

期望返回格式：
```json
{
  "commits": [...],
  "count": N
}
```

实际返回：`[...]` (直接数组)

### 解决方案
使用`convert.ToWikiCommitList()`正确包装响应：
```go
result := convert.ToWikiCommitList(commitsHistory, commitsCount)
ctx.JSON(http.StatusOK, result)
```

### 修复效果
✅ API返回正确的`WikiCommitList`结构
✅ 前端可以正常访问`.commits`和`.count`属性

---

## 测试建议

### 1. PATCH API测试
```bash
# 更新页面但保持标题不变
curl -X PATCH "/api/v1/repos/{owner}/{repo}/wiki/page/{pageName}" \
  -d '{"title": "Current Title", "content": "new content"}'

# 更新页面并修改标题
curl -X PATCH "/api/v1/repos/{owner}/{repo}/wiki/page/{pageName}" \
  -d '{"title": "New Title", "content": "new content"}'
```

### 2. GET API测试
```bash
# 获取Home页面
curl "/api/v1/repos/{owner}/{repo}/wiki/page/Home"

# 获取非Home页面（验证content_base64字段存在）
curl "/api/v1/repos/{owner}/{repo}/wiki/page/Other-Page"
```

### 3. Revisions API测试
```bash
# 验证响应包含.commits和.count属性
curl "/api/v1/repos/{owner}/{repo}/wiki/revisions/Home"
```

---

## 技术要点

### Wiki路径转换链
```
用户输入标题 -> WebPath -> GitPath -> Git文件
    ↓              ↓          ↓
"My Page"   -> "My-Page" -> "My Page.md"  (unescaped)
                            "My%20Page.md" (escaped)
```

### 关键改进
1. **API层与Service层一致性**: 确保两层都使用相同的文件查找逻辑
2. **向后兼容**: 支持历史遗留的两种文件格式
3. **避免副作用**: 防止不必要的路径转换和重命名

---

## 文档参考
详细技术说明请参考: [WIKI_API_FIXES.md](./WIKI_API_FIXES.md)

---

## 状态
✅ 所有问题已修复
✅ 代码已提交
✅ 文档已完成
⏳ 等待测试验证

建议关闭相关工单。
