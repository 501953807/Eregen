# ⑥ 品牌官网 — 详细设计文档

> 生成日期：2026-07-17
> 最后更新：2026-08-02 — content/ 目录待填充  
> 对应子系统：⑥ 品牌官网 (Hugo + Tailwind CSS)  
> 框架：Hugo Static Site Generator | CSS：Tailwind 3.4+

---

## 1. 概述

### 1.1 职责

品牌官网是 Eregen 对外的品牌展示窗口，面向消费者、合作伙伴和潜在投资方。提供品牌故事介绍、三档产品线展示、技术白皮书下载、ODM 合作联系入口等功能。作为静态站点托管于 Netlify/Vercel，无需后端服务。

### 1.2 输入输出

| 类型 | 来源/目标 | 说明 |
|------|-----------|------|
| **输入** | 网站访问者浏览 | 静态 HTML/CSS/JS |
| **输出** | 表单提交 | 联系表单 → 邮件通知运营人员 |
| **输出** | 产品购买链接 | 跳转至电商页面或小程序 |

---

## 2. 功能模块

### 2.1 核心页面

| 页面 | Hugo 文件 | 内容 |
|------|----------|------|
| 首页 | `content/_index.md` | 品牌 Slogan、核心价值主张、产品入口 |
| 品牌故事 | `content/about/index.md` | 颐贞品牌理念、团队介绍、专利信息 |
| 产品介绍 | `content/products/index.md` | 三档手环/药盒对比、规格参数 |
| 技术白皮书 | `content/whitepaper/index.md` | 系统架构、通信协议、安全设计 |
| ODM 合作 | `content/partner/index.md` | 合作模式、联系方式、资质要求 |
| 合规声明 | `content/legal/index.md` | 隐私政策、用户协议、广告合规 |
| 博客 | `content/blog/_index.md` | 健康科普、产品更新、行业洞察 |

### 2.2 原型文件

| 文件 | 说明 |
|------|------|
| `apps/ui-prototypes/website-home.html` | 官网首页高保真原型 |

---

## 3. 技术架构

### 3.1 项目结构

```
apps/website/
├── content/                     # Hugo 内容 (Markdown)
│   ├── _index.md                # 首页
│   ├── about/
│   │   ├── index.md
│   │   └── _index.md
│   ├── products/
│   │   ├── index.md
│   │   ├── bracelet.md          # 手环详细介绍
│   │   └── pillbox.md           # 药盒详细介绍
│   ├── whitepaper/
│   │   └── index.md
│   ├── partner/
│   │   └── index.md
│   ├── legal/
│   │   ├── privacy.md
│   │   └── terms.md
│   └── blog/
│       ├── _index.md
│       └── 2026-07-01-eregen-launch.md
├── layouts/                     # Hugo 模板
│   ├── _default/
│   │   ├── baseof.html          # 基础布局
│   │   ├── list.html            # 列表页
│   │   └── single.html          # 单页
│   ├── index/
│   │   └── hero.html            # 首页 Hero 区域
│   └── partials/
│       ├── navbar.html          # 导航栏
│       ├── footer.html          # 页脚
│       └── product-card.html    # 产品卡片组件
├── static/                      # 静态资源
│   ├── images/                  # 产品图/品牌图
│   ├── fonts/                   # 自定义字体
│   └── downloads/               # 白皮书 PDF 下载
├── assets/                      # Tailwind 源文件
│   └── css/
│       └── styles.css           # @tailwind 指令
├── hugo.toml                    # Hugo 配置
├── tailwind.config.js           # Tailwind 配置
└── package.json                 # NPM 依赖
```

### 3.2 技术栈

| 库 | 版本 | 用途 |
|----|------|------|
| Hugo | 0.128+ | 静态站点生成器 |
| Tailwind CSS | 3.4+ | 原子化 CSS 框架 |
| Alpine.js | 3.x | 轻量交互 (导航切换/滚动动画) |
| Google Fonts | — | Noto Sans SC (中文正文字体) |

### 3.3 构建流程

```bash
cd apps/website

# 本地开发 (热更新)
npm run dev
# Hugo server --minify --renderToMemory
# Tailwind watch

# 生产构建
npm run build
# hugo --minify --gc
# Tailwind purge → 最小化 CSS
# 输出: public/
```

---

## 4. 首页布局设计

```
┌─────────────────────────────────────────────────────┐
│  [Eregen Logo]              产品  品牌  技术  合作  │
├─────────────────────────────────────────────────────┤
│                                                     │
│              颐养正道 贞守安康                       │
│           Eregen — Elder Health Ecosystem            │
│                                                     │
│     [查看产品]      [了解技术]                       │
│                                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│   ┌─────────────┐  ┌─────────────┐  ┌─────────────┐│
│   │  手环系列    │  │  药盒系列    │  │  B2B 生态   ││
│   │  心率/SOS/   │  │  自动分药/   │  │  医院/社区  ││
│   │  定位/跌倒   │  │  语音提醒    │  │  保险对接   ││
│   │  [详情→]     │  │  [详情→]     │  │  [详情→]     ││
│   └─────────────┘  └─────────────┘  └─────────────┘│
│                                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│         三档产品线对比                               │
│   ┌────────┬────────┬────────┐                     │
│   │ Starter│  Plus  │  Pro   │                     │
│   │ 80-120 │180-230 │280-400 │                     │
│   │ 心率血氧│跌倒围栏│ECG+AMOLED│                   │
│   └────────┴────────┴────────┘                     │
│                                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│         技术优势                                     │
│   · 全栈自研 · 专利保护 · 开源许可合规              │
│   · Ed25519 设备认证 · AES-256-GCM 加密            │
│   · Go 微服务架构 · NATS 消息总线                   │
│                                                     │
├─────────────────────────────────────────────────────┤
│                                                     │
│  © 2026 Eregen (颐贞) | 隐私政策 | 用户协议        │
│                                                     │
└─────────────────────────────────────────────────────┘
```

---

## 5. 部署

### 5.1 Netlify

```toml
# netlify.toml
[build]
  command = "cd apps/website && npm run build"
  publish = "apps/website/public"

[[redirects]]
  from = "/*"
  to = "/index.html"
  status = 200
```

### 5.2 Vercel

```json
// vercel.json
{
  "buildCommand": "cd apps/website && npm run build",
  "outputDirectory": "apps/website/public",
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ]
}
```

### 5.3 域名配置

```
DNS 记录 (阿里云 DNS):
  CNAME  www → your-site.netlify.app
  A     @ → 104.21.xx.xx (Netlify IP)

SSL: Let's Encrypt 自动续期
```

---

## 6. SEO 优化

| 策略 | 实现 |
|------|------|
| 语义化 HTML | Hugo semantic templates |
| Meta 标签 | 每页独立 title/description/keywords |
| Open Graph | 社交分享卡片图片 |
| Sitemap | `hugo generateSitemap()` 自动生成 |
| robots.txt | 允许搜索引擎抓取 |
| 结构化数据 | JSON-LD Product/Organization schema |

---

© 2026 Eregen (颐贞). All rights reserved.
