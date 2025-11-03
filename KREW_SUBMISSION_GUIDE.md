# kubectl-triage Krew 提交指南

本文档提供了将 kubectl-triage 提交到 Krew 插件索引的完整指南。

## ✅ 准备工作检查清单

基于 Krew 官方文档的所有要求，kubectl-triage 已完成：

### 1. 命名规范 ✅
- ✅ 使用 kebab-case（小写 + 连字符）
- ✅ 具体明确："triage" 是医学术语，指快速诊断分类
- ✅ 唯一性：区别于其他日志/诊断工具
- ✅ 无通用动词或名词
- ✅ 未使用 "kube-" 或 "kubernetes-" 前缀
- ✅ 不是资源首字母缩写

**结论**: 插件名称 "triage" 完全符合 Krew 命名标准

### 2. 源代码和许可证 ✅
- ✅ 源代码公开：https://github.com/Lc-Lin/kubectl-triage
- ✅ 开源许可证：Apache License 2.0
- ✅ LICENSE 文件包含在发布归档中
- ✅ 安装时提取 LICENSE 文件

### 3. 语义化版本 ✅
- ✅ 使用 git tag: `v0.1.0`
- ✅ GitHub Release 已发布
- ✅ 版本格式正确（带 v 前缀）

### 4. 插件清单 ✅
文件位置：`deploy/krew/plugin.yaml`

**必填字段**：
- ✅ `apiVersion`: krew.googlecontainertools.github.com/v1alpha2
- ✅ `kind`: Plugin
- ✅ `metadata.name`: triage
- ✅ `spec.version`: "v0.1.0"
- ✅ `spec.shortDescription`: 简洁明了
- ✅ `spec.description`: 详细功能说明，包含特性列表
- ✅ `spec.platforms`: Linux, macOS, Windows 配置完整

**可选字段**：
- ✅ `homepage`: https://github.com/Lc-Lin/kubectl-triage
- ✅ `caveats`: 简洁的使用说明

**平台配置**：
```yaml
platforms:
  - Linux amd64   ✅ SHA256 已填写
  - macOS amd64   ✅ SHA256 已填写
  - Windows amd64 ✅ SHA256 已填写
```

**文件配置**：
- ✅ 提取二进制文件
- ✅ 提取 LICENSE 文件
- ✅ bin 字段正确指定可执行文件

### 5. 最佳实践 ✅
- ✅ 使用 Go 编写（推荐语言）
- ✅ 使用 client-go 和 cli-runtime
- ✅ 支持常见 kubectl 标志：
  - ✅ `-h`/`--help`
  - ✅ `-n`/`--namespace`
  - ✅ `--kubeconfig`
  - ✅ `--context`
- ✅ 帮助信息显示 `kubectl` 前缀
- ✅ 支持云提供商认证（client-go/plugin）

---

## 📋 Krew 提交步骤

### 步骤 1: Fork krew-index 仓库

1. 访问 https://github.com/kubernetes-sigs/krew-index
2. 点击右上角的 **Fork** 按钮
3. Fork 到你的 GitHub 账号

### 步骤 2: 克隆你的 Fork

```bash
git clone https://github.com/YOUR_USERNAME/krew-index.git
cd krew-index
```

### 步骤 3: 创建插件清单

```bash
# 复制插件清单到 plugins 目录
cp ~/Documents/code/kubectl-triage/deploy/krew/plugin.yaml plugins/triage.yaml

# 验证文件
cat plugins/triage.yaml
```

### 步骤 4: 提交更改

```bash
# 添加文件
git add plugins/triage.yaml

# 提交（使用规范的提交信息）
git commit -m "Add kubectl-triage plugin

kubectl-triage is a fast diagnostic tool for failed Kubernetes pods.

It provides 5-second diagnostic snapshots by intelligently aggregating:
- Pod status and container states
- Critical events (Warning/Error only)
- Previous crash logs (if container restarted)
- Current container logs

Only failed/restarted containers are shown by default.

Key features:
- Smart health detection (catches flapping pods)
- Intelligent container filtering (RestartCount > 0)
- Signal-over-noise event filtering
- Parallel log collection for speed
- Beautiful color-coded output

GitHub: https://github.com/Lc-Lin/kubectl-triage
License: Apache-2.0
Version: v0.1.0
"

# 推送到你的 fork
git push origin main
```

### 步骤 5: 创建 Pull Request

1. 访问 https://github.com/kubernetes-sigs/krew-index
2. 点击 **New Pull Request**
3. 点击 **compare across forks**
4. 选择：
   - base repository: `kubernetes-sigs/krew-index`
   - base: `main`
   - head repository: `YOUR_USERNAME/krew-index`
   - compare: `main`
5. 填写 PR 信息：

**标题**：
```
Add kubectl-triage plugin
```

**描述**：
```markdown
## Plugin Information

- **Name**: kubectl-triage
- **Version**: v0.1.0
- **Homepage**: https://github.com/Lc-Lin/kubectl-triage
- **License**: Apache-2.0

## Description

kubectl-triage provides fast diagnostic snapshots for failed Kubernetes pods. It's designed as a "first responder" tool that aggregates critical information in one command.

## Features

- 5-second diagnostic snapshots
- Smart health detection (3-condition check)
- Intelligent container filtering (only shows failed/restarted containers)
- Signal-over-noise event filtering (Warning/Error only)
- Parallel log collection
- Beautiful color-coded output

## Testing

- [x] Plugin tested locally with --manifest flag
- [x] All platforms (Linux, macOS, Windows) have valid SHA256 checksums
- [x] Source code is open source (Apache-2.0)
- [x] LICENSE file included in all archives
- [x] Binary tested and working correctly

## Checklist

- [x] Plugin name follows naming guidelines
- [x] Manifest follows required format
- [x] All required fields present
- [x] Platform configurations complete
- [x] SHA256 checksums verified
- [x] Semantic versioning used
- [x] Documentation complete
```

6. 点击 **Create Pull Request**

---

## 📝 PR 审核流程

### 预期审核内容

Krew 维护者会检查：

1. **命名规范**
   - 插件名称是否符合规范
   - 是否有名称冲突

2. **清单格式**
   - YAML 格式是否正确
   - 所有必填字段是否存在
   - 字段内容是否合理

3. **SHA256 校验和**
   - 校验和是否正确
   - URI 是否可访问

4. **许可证**
   - 是否包含开源许可证
   - LICENSE 文件是否提取

5. **描述质量**
   - shortDescription 是否简洁
   - description 是否清晰有用

### 审核时间

- 通常 1-3 天
- 可能需要进行修改
- 维护者会在 PR 中提供反馈

### 常见问题

**Q: 如果需要修改怎么办？**
A: 在你的 fork 中修改，然后推送，PR 会自动更新。

**Q: SHA256 不匹配怎么办？**
A: 重新下载归档文件，重新计算 SHA256：
```bash
shasum -a 256 kubectl-triage_darwin_amd64.tar.gz
```

**Q: 插件名称被拒绝怎么办？**
A: 根据反馈选择新名称，更新清单和仓库。

---

## ✅ 提交后验证

PR 合并后（通常 1-3 天），用户可以通过以下方式安装：

```bash
# 更新 krew 索引
kubectl krew update

# 搜索插件
kubectl krew search triage

# 安装插件
kubectl krew install triage

# 使用插件
kubectl triage <pod-name>
```

---

## 🎯 kubectl-triage 符合性总结

| 检查项 | 状态 | 说明 |
|--------|------|------|
| 命名规范 | ✅ | "triage" 符合所有命名规则 |
| 源代码公开 | ✅ | GitHub 公开仓库 |
| 开源许可证 | ✅ | Apache-2.0 |
| LICENSE 文件 | ✅ | 包含在归档中 |
| 语义化版本 | ✅ | v0.1.0 |
| 清单格式 | ✅ | 所有必填字段完整 |
| 平台支持 | ✅ | Linux, macOS, Windows |
| SHA256 校验和 | ✅ | 所有平台已填写 |
| 最佳实践 | ✅ | 使用 Go, client-go |
| kubectl 标志 | ✅ | 支持标准标志 |
| 文档 | ✅ | 完整的 README 和 USAGE |

**结论**: kubectl-triage 已完全准备好提交到 Krew！

---

## 📚 参考资源

- **Krew 官方文档**: https://krew.sigs.k8s.io/docs/developer-guide/
- **插件命名指南**: https://krew.sigs.k8s.io/docs/developer-guide/develop/naming-guide/
- **插件清单格式**: https://krew.sigs.k8s.io/docs/developer-guide/plugin-manifest/
- **本地测试指南**: https://krew.sigs.k8s.io/docs/developer-guide/testing-locally/
- **最佳实践**: https://krew.sigs.k8s.io/docs/developer-guide/develop/best-practices/
- **krew-index 仓库**: https://github.com/kubernetes-sigs/krew-index

---

## 💡 提交建议

1. **时机**: 选择工作日提交，维护者活跃时间响应更快
2. **沟通**: 在 PR 中积极回应维护者的反馈
3. **耐心**: 审核过程可能需要几天，保持耐心
4. **跟进**: PR 合并后，关注 GitHub Discussions 和 issues

---

**准备完毕！现在可以提交到 Krew 了！** 🚀
