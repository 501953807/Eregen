# Eregen 已知问题与解决方案

## 问题记录

### [2026-08-02] 数据库文件被误提交到 git
- **现象**: .db 文件被追踪，导致仓库体积过大
- **根因**: .gitignore 缺少 *.db 规则
- **解决**: 已更新 .gitignore 并从 git 历史中移除
- **状态**: ✅ 已解决

### [2026-08-02] 二进制构建产物残留
- **现象**: cloud/admin-api_fixed 等二进制文件被提交
- **根因**: .gitignore 缺少 *_fixed 规则
- **解决**: 已更新 .gitignore 并移除追踪
- **状态**: ✅ 已解决

### [2026-08-02] 计划文档过度膨胀
- **现象**: docs/superpowers/plans/ 下有 25 份文档，大量过时
- **根因**: 每次迭代都创建新文档，未归档旧文档
- **解决**: 已归档 19 份过时文档到 docs/archive/
- **状态**: ✅ 已解决

### [2026-08-02] 存在重复开发目录
- **现象**: SP1-auth/ 是 apps/admin-web 的副本
- **根因**: 多环境开发产生的遗留目录
- **解决**: 已添加到 .gitignore
- **状态**: ⚠️ 待清理

### [2026-08-01] admin-web 代码行数异常
- **现象**: Elderly.vue 文件超过 1100 行
- **根因**: 单一组件承担了过多功能
- **建议**: 拆分为多个子组件
- **状态**: 🔄 待优化

### [2026-08-01] family-app 代码行数异常
- **现象**: Flutter 项目显示 346K 行（实际 Dart 代码约 7K 行）
- **根因**: 包含 iOS/Android 原生代码和依赖
- **说明**: 实际业务代码规模合理
- **状态**: ℹ️ 已知

## 常见问题 FAQ

### Q: 如何启动单个服务？
```bash
./scripts/start.sh start <service-name>
```

### Q: 如何查看服务日志？
```bash
./scripts/start.sh logs <service-name>
```

### Q: 端口冲突怎么办？
```bash
./scripts/start.sh ports-check
./scripts/start.sh start --force <service-name>
```

---
最后更新：2026-08-02

### [2026-08-02] 项目根目录垃圾文件清理
- **现象**: SP1-auth/ (重复副本), Users/ (系统目录), admin-api.log (日志) 污染根目录
- **解决**: 已删除 SP1-auth/, Users/, admin-api.log
- **状态**: ✅ 已解决
