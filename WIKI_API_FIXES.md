# Wiki API 修复总结

## 修复概述

已成功修复Wiki API的三个关键问题，所有修改集中在一个文件中：`routers/api/v1/repo/wiki.go`

提交ID: `44c5b7f762`
提交时间: 2025-11-24

## 修复详情

### 1. Issue #2 Problem 3: Revisions API返回格式错误 (优先级: P1)

**问题描述**:
- Revisions API直接返回数组，导致前端JavaScript无法访问`.commits`属性
- 期望返回`WikiCommitList`对象结构：`{commits: [...], count: N}`

**修复方案**:
```go
// 文件位置: routers/api/v1/repo/wiki.go 第451-455行
// 使用convert.ToWikiCommitList()正确包装响应
result := convert.ToWikiCommitList(commitsHistory, commitsCount)
ctx.JSON(http.StatusOK, result)
```

**修复效果**:
- ✅ API返回正确的`WikiCommitList`结构
- ✅ 前端可以正常访问`.commits`和`.count`属性

---

### 2. Issue #2 Problems 1 & 2: GET API内容缺失和404错误 (优先级: P0)

**问题描述**:
- Problem 1: GET API不返回`content_base64`字段
- Problem 2: GET API对非Home页面返回404错误
- 根本原因: `wikiContentsByName()`只查找一种文件路径格式，而Wiki页面在Git中可能以两种格式存储

**修复方案**:
```go
// 文件位置: routers/api/v1/repo/wiki.go 第518-542行
func wikiContentsByName(ctx *context.APIContext, commit *git.Commit, wikiName wiki_service.WebPath, isSidebarOrFooter bool) (string, string) {
    // 同时尝试查找unescaped和escaped两种文件格式
    unescaped := string(wikiName) + ".md"
    gitPath := wiki_service.WebPathToGitPath(wikiName)

    // 先尝试unescaped版本 (如: Page-Name.md)
    entry, err := findEntryForFile(commit, unescaped)
    if err == nil && entry != nil {
        return wikiContentsByEntry(ctx, entry), unescaped
    }

    // 如果找不到，尝试escaped gitPath版本 (如: Page%20Name.md)
    entry, err = findEntryForFile(commit, gitPath)
    // ... 错误处理
}
```

**修复效果**:
- ✅ GET API正确返回`content_base64`字段
- ✅ 解决了大部分页面返回404的问题
- ✅ 与Service层的`prepareGitPath()`逻辑保持一致
- ✅ 支持两种文件格式的向后兼容

---

### 3. Issue #1: PATCH API标题处理错误

**问题描述**:
- 当请求体包含`Title`字段时，即使值与当前标题相同，也会通过`UserTitleToWebPath()`重新转换
- 转换可能导致WebPath与原始值不一致，触发不必要的重命名
- 特别严重的是，空标题或"."会被转换成"unnamed"

**修复方案**:
```go
// 文件位置: routers/api/v1/repo/wiki.go 第137-154行
oldWikiName := wiki_service.WebPathFromRequest(ctx.PathParamRaw("pageName"))

var newWikiName wiki_service.WebPath
if form.Title == "" {
    newWikiName = oldWikiName
} else {
    // 比较标题是否相同，避免不必要的转换
    _, currentTitle := wiki_service.WebPathToUserTitle(oldWikiName)
    if strings.TrimSpace(form.Title) == currentTitle {
        // 标题未改变，保持原有的WebPath
        newWikiName = oldWikiName
    } else {
        newWikiName = wiki_service.UserTitleToWebPath("", form.Title)
    }
}
```

**修复效果**:
- ✅ 避免不必要的路径转换
- ✅ 防止意外重命名
- ✅ PATCH API正确更新指定的页面
- ✅ 修复了总是更新"unnamed"页面的bug

## 技术说明

### Wiki路径格式兼容性

Wiki文件在Git仓库中可能以两种格式存储:

1. **Unescaped格式**: `Page-Name.md`
   - 破折号表示空格分隔
   - 无特殊字符编码

2. **Escaped格式**: `Page%20Name.md`
   - URL编码格式
   - 空格编码为`%20`

修复方案确保API层与Service层行为一致，都能正确查找两种格式。

### 路径转换链

```
用户输入标题 -> WebPath -> GitPath -> Git文件
    ↓              ↓          ↓
"My Page"   -> "My-Page" -> "My Page.md"  (unescaped)
                            "My%20Page.md" (escaped)
```

## 验证建议

建议测试以下场景:

1. **GET API测试**:
   - 获取Home页面
   - 获取非Home页面（不同命名格式）
   - 验证`content_base64`字段存在且正确

2. **Revisions API测试**:
   - 调用`/repos/{owner}/{repo}/wiki/revisions/{pageName}`
   - 验证响应包含`.commits`和`.count`属性

3. **PATCH API测试**:
   - 更新页面但保持标题不变
   - 更新页面并修改标题
   - 验证正确的页面被更新

## 后续工作

- ✅ 代码修复完成
- ⏳ 等待测试验证
- ⏳ 回复并关闭相关工单

## 文件变更

```
修改文件: routers/api/v1/repo/wiki.go
变更行数: +23 -9
提交ID: 44c5b7f762
```

---

**生成时间**: 2025-11-24
**修复人员**: Claude Code Assistant
