# CCM Desktop v2

Claude Code 桌面配置管理器。可视化管理 Skills、Plugins、Memory、MCP、CLAUDE.md，支持可移植性分析、密钥扫描、备份恢复。

**Wails v2 + Go + React + shadcn/ui**

![Theme](https://img.shields.io/badge/themes-5-blue) ![License](https://img.shields.io/badge/license-MIT-green) ![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-blue)

## 功能

- **Skills 管理** — 列表、详情、验证、导入 GitHub Skill、启用/禁用、批量删除、中文翻译
- **插件管理** — 查看所有插件及 Skill，一键/单个启用禁用
- **Memory 管理** — 创建/查看/删除 Memory，按类型统计，搜索过滤
- **MCP 管理** — 列出 MCP 服务器、状态检查、启用/禁用
- **CLAUDE.md** — 浏览、创建、编辑、删除所有层级 CLAUDE.md
- **可移植性分析** — 检测跨电脑迁移问题，一键修复绝对路径
- **密钥扫描** — 扫描配置文件中的 API Key/Token
- **备份恢复** — ZIP 打包备份、一键还原
- **中文翻译** — 百度翻译 API（免费 200 万字符/月），Skill 描述自动中文化
- **5 套主题** — Lacquer 漆器、Alabaster 石膏、Slate 钛空、Photon 光量子、Obsidian 黑曜
- **键盘快捷键** — `Ctrl+1`~`Ctrl+9` 切换页面
- **文件监听** — 外部文件变化自动刷新 UI

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | Wails v2 |
| 后端 | Go 1.23 |
| 前端 | React 18 + TypeScript |
| UI 组件 | shadcn/ui + Tailwind CSS |
| 状态管理 | TanStack Query |
| 通信 | Wails 运行时绑定 |
| 翻译 | 百度翻译 API |
| 打包 | 单 exe (12MB) |

## 快速使用

1. 从 [Releases](../../releases) 下载最新 `CCM Desktop.exe`
2. 双击运行（无需安装，无需 npm，单个文件）

## 从源码构建

### 前提条件

- [Go](https://go.dev/dl/) 1.23+
- [Node.js](https://nodejs.org/) 18+
- [Wails v2](https://wails.io/docs/gettingstarted/installation) CLI
- Windows 10/11

### 构建步骤

```bash
# 1. 克隆仓库
git clone https://github.com/Sisyphus776/ccm-desktop-v2.git
cd ccm-desktop-v2

# 2. 安装依赖
npm install

# 3. 构建前端 (Vite)
npm run build -w frontend

# 4. Wails 构建 (Go + 前端 → 单 exe)
wails build
```

输出文件: `build/bin/CCM Desktop.exe` (~12MB)

## 百度翻译 API（可选）

Settings 页面可配置百度翻译 API 凭证（注册 [fanyi-api.baidu.com](https://fanyi-api.baidu.com)，免费 200 万字符/月）。配置后 Skill 描述自动中文化。

## 项目结构

```
ccm-desktop-v2/
├── main.go                   # Wails 入口
├── app.go                    # Go 后端方法绑定
├── wails.json                # Wails 构建配置
├── frontend/                 # React 前端
│   └── src/
│       ├── pages/            # 10 个页面
│       ├── components/       # Sidebar + shadcn/ui
│       ├── lib/              # RPC client, types, theme, Wails bindings
│       └── i18n/             # 中英双语
├── backend/                  # Go 后端
│   ├── rpc/                  # 业务方法 (skills, plugins, memory, mcp...)
│   └── cmd/ccm-backend/      # CLI 入口 (调试用)
├── internal/                 # Go 核心逻辑
│   ├── skills/               # Skill 发现、验证
│   ├── plugins/              # 插件扫描
│   ├── memory/               # Memory 管理
│   ├── mcp/                  # MCP 配置
│   ├── translate/            # 百度翻译
│   └── ...                   # 其他包
└── package.json              # npm workspaces
```

## 许可证

MIT License
