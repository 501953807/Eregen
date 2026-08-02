# Eregen 开发规范

## 分支管理

```
main          # 稳定分支，可部署
develop       # 开发分支
feature/xxx   # 功能开发分支
hotfix/xxx    # 紧急修复分支
```

## 提交消息规范 (Conventional Commits)

```
<type>(<scope>): <subject>

type:
  feat:     新功能
  fix:      修复 bug
  docs:     文档更新
  style:    代码格式（不影响功能）
  refactor: 重构
  test:     测试相关
  chore:    构建/工具相关

示例:
  feat(admin-web): add user management page
  fix(api-server): resolve device connection timeout
  docs: update ARCHITECTURE.md
```

## 代码规范

### Go
- 使用 `gofmt` 格式化代码
- 遵循 Go 官方代码规范
- 错误处理必须显式处理，禁止忽略 error

### TypeScript/Vue
- 使用 ESLint 检查
- 组件命名使用 PascalCase
- 变量命名使用 camelCase
- 常量命名使用 UPPER_SNAKE_CASE

### Flutter/Dart
- 使用 `dart format` 格式化
- 遵循 Effective Dart 指南
- 组件命名使用 PascalCase

## 文档更新要求

每次代码变更后，需同步更新：
1. 相关子项目的 README.md
2. docs/specs/ 中的对应规格文档
3. 根目录 CLAUDE.md 的"当前阶段实现状态"章节

## 数据库变更

1. 修改数据库 schema 时，必须更新 `init-db.sql`
2. 使用迁移脚本管理版本化变更
3. 禁止直接修改已部署的数据库文件

## 禁止提交的内容

- .db 数据库文件
- 二进制构建产物 (*.bin, *_fixed, *.o, *.so 等)
- .env 环境变量文件
- node_modules/, .dart_tool/, vendor/
- IDE 配置 (.vscode/, .idea/)
- 临时日志文件 (*.log)

## 开发流程

1. 从 main 创建 feature/xxx 分支
2. 开发并本地测试
3. 提交前运行格式化工具
4. 推送并创建 Pull Request
5. 代码审查通过后合并到 main

## 常用命令

```bash
# 启动所有服务
./scripts/start.sh start --all

# 查看状态
./scripts/start.sh status --all

# 查看日志
./scripts/start.sh logs <service>

# 清理运行时文件
./scripts/start.sh clean

# 端口冲突检测
./scripts/start.sh ports-check
```

---
最后更新：2026-08-02
