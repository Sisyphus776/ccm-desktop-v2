# CCM Desktop v2.1

Claude Code 桌面配置管理器。可视化管理 Skills、Plugins、Memory、MCP、CLAUDE.md，支持可移植性分析、密钥扫描、备份恢复。

**Electron + Go + React + shadcn/ui**

![Theme](https://img.shields.io/badge/themes-5-blue) ![License](https://img.shields.io/badge/license-MIT-green) ![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-blue)

## 功能

- **Skills 管理** — 列表、详情、验证、导入 GitHub Skill、启用/禁用、批量删除、中文翻译
- **插件管理** — 查看所有插件及 Skill，一键/单个启用禁用
- **Memory 管理** — 创建/查看/删除 Memory，按类型统计，搜索过滤
- **MCP 管理** — 列出 MCP 服务器、状态检查、启用/禁用
- **CLAUDE.md** — 浏览所有层级 CLAUDE.md 内容
- **可移植性分析** — 检测跨电脑迁移问题，一键修复绝对路径
- **密钥扫描** — 扫描配置文件中的 API Key/Token
- **备份恢复** — ZIP 打包备份、一键还原
- **中文翻译** — 内置 41 条种子翻译 + Youdao API，Skill 描述自动中文化
- **5 套主题** — Lacquer 漆器、Alabaster 石膏、Slate 钛空、Photon 光量子、Obsidian 黑曜
- **键盘快捷键** — `Ctrl+1`~`Ctrl+9` 切换页面，`Ctrl+F` 搜索
- **系统托盘** — 最小化到托盘，窗口标题显示状态计数

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | Electron 33 |
| 后端 | Go 1.23 |
| 前端 | React 18 + TypeScript |
| UI 组件 | shadcn/ui + Tailwind CSS |
| 状态管理 | TanStack Query |
| 通信协议 | JSON-RPC 2.0 (stdin/stdout) |
| 翻译 | Youdao API + 内置种子词库 |

## 快速使用（便携版）

1. 从 [Releases](../../releases) 下载 `ccm-desktop-v2.1-portable.zip`
2. 解压到任意目录
3. 首次运行双击 `setup.bat`（安装 Electron 运行时，仅需一次）
4. 之后每次双击 `CCM Desktop v2.1.exe` 启动
5. 也可右键 `CCM Desktop v2.1.exe` → 发送到桌面快捷方式

## 从源码构建

### 前提条件

- [Go](https://go.dev/dl/) 1.23+
- [Node.js](https://nodejs.org/) 18+
- Windows 10/11

### 构建步骤

```bash
# 1. 克隆仓库
git clone https://github.com/Sisyphus776/ccm-desktop-v2.git
cd ccm-desktop-v2

# 2. 安装依赖
npm install

# 3. 编译 Go 后端
cd backend
go build -ldflags="-H windowsgui" -o ../desktop/ccm-backend.exe ./cmd/ccm-backend
cd ..

# 4. 构建前端
npm run build -w frontend

# 5. 编译 Electron
npm run build -w desktop

# 6. 启动
npx electron desktop/main.js
```

## 项目结构

```
ccm-desktop-v2/
├── backend/                 # Go 后端
│   ├── cmd/ccm-backend/     # 入口
│   ├── rpc/                 # JSON-RPC 方法
│   └── internal/            # 核心逻辑（skills, memory, mcp, translate...）
├── desktop/                 # Electron 主进程
│   ├── main.ts              # 窗口管理 + Go 生命周期
│   └── preload.ts           # IPC 桥接
├── frontend/                # React 前端
│   └── src/
│       ├── pages/           # 10 个页面
│       ├── components/      # Sidebar + shadcn/ui
│       ├── lib/             # RPC client, types, theme
│       └── i18n/            # 中英双语
└── package.json             # npm workspaces
```

## 许可证

MIT License
