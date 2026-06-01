# 投资助手

实时监控个股与基金的盈亏状态，支持止盈止损预警、大盘行情看板、基金重仓股同步等功能。

---

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22 + Gin + GORM + SQLite |
| 前端 | Vue 3 + Vite + TypeScript + Element Plus + Pinia |
| 定时任务 | robfig/cron/v3 |
| 数据源 | AKShare (Python Bridge) + 新浪行情 API |

---

## 项目结构

```
stock/
├── backend-go/          # Go 后端
│   ├── cmd/server/      # 服务入口
│   ├── internal/
│   │   ├── api/         # HTTP Handler
│   │   ├── service/     # 业务逻辑
│   │   ├── repository/  # 数据访问 (GORM)
│   │   ├── model/       # 数据模型
│   │   ├── schedule/    # 定时任务
│   │   ├── datasource/  # 外部数据获取
│   │   └── strategy/    # 预警策略
│   ├── scripts/         # AKShare Python Bridge
│   └── pkg/logger/      # 日志封装
├── frontend/            # Vue3 前端
│   ├── src/
│   │   ├── views/       # 页面组件
│   │   ├── components/  # 可复用组件
│   │   ├── api/         # API 封装
│   │   ├── stores/      # Pinia 状态管理
│   │   └── utils/       # 工具函数
│   └── dist/            # 构建产物
└── data/                # SQLite 数据库
    └── stock_monitor.db
```

---

## 快速启动

### 环境要求

- Go 1.22+
- Node.js 20+
- Python 3.10+（AKShare 数据桥接依赖）

### 1. 启动后端

```bash
cd backend-go
go run cmd/server/main.go
```

后端默认运行在 `http://localhost:8000`

### 2. 构建并启动前端

```bash
cd frontend
npm install
npm run build
```

前端静态文件由后端自动代理，无需单独启动服务器。直接访问 `http://localhost:8000` 即可。

### 3. 开发模式（前端热更新）

```bash
cd frontend
npm run dev
```

开发服务器运行在 `http://localhost:5173`，代理已配置到 `http://localhost:8000`。

---

## 功能模块

### 大盘看板
- 全球主要市场实时行情（A股、港股、美股、日韩、商品、美债）
- 自动刷新（30秒）
- 交易状态与北京时间显示

### 我的持仓
- 投资组合总览（总成本、总市值、总盈亏、盈亏比例）
- 个股/基金持仓卡片，实时显示价格与盈亏

### 标的配置
- 个股/基金 CRUD
- 新增个股支持远程模糊搜索
- 止盈止损线、监控频率配置

### 基金持仓
- 选择基金查看重仓股明细
- 持仓占比可视化进度条
- 同步最新重仓数据

### 预警记录
- 单日/累计止盈止损触发记录
- 触发值、阈值、当前价、触发时间

---

## 定时任务

| 任务 | 执行时间 | 说明 |
|------|----------|------|
| 个股监控 | 工作日 9:00-15:00 每分钟 | 拉取实时行情、保存每日记录、检查止盈止损 |
| 基金估算 | 工作日 14:50 | 拉取基金估算净值、保存记录、检查止盈止损 |
| 基金同步 | 工作日 15:30 | 同步所有基金重仓股 |
| 收盘更新 | 工作日 15:05 | 更新个股当日收盘价 |

---

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/stocks` | 个股列表（含实时行情） |
| POST | `/api/stocks` | 新增个股 |
| PUT | `/api/stocks/{code}` | 编辑个股 |
| DELETE | `/api/stocks/{code}` | 删除个股 |
| GET | `/api/stocks/search?q=` | 股票模糊搜索 |
| GET | `/api/funds` | 基金列表（含估算净值） |
| POST | `/api/funds` | 新增基金 |
| PUT | `/api/funds/{code}` | 编辑基金 |
| DELETE | `/api/funds/{code}` | 删除基金 |
| GET | `/api/fund-holdings/{code}` | 基金持仓列表 |
| POST | `/api/fund-holdings/sync/{code}` | 同步基金持仓 |
| GET | `/api/alerts` | 预警记录 |
| GET | `/api/daily-records` | 每日记录 |
| GET | `/api/stats` | 投资组合统计 |
| GET | `/api/market-dashboard` | 大盘行情看板 |

---

## 关键设计决策

### 数据获取：AKShare Bridge

AKShare 是 Python 库，Go 通过 `os/exec` 调用 `scripts/akshare_bridge.py` 获取数据。当 AKShare 网络不可用时，自动降级为直接调用新浪行情 API。

### 精度控制

使用 `shopspring/decimal` 处理金额与价格，避免浮点精度问题。JSON 序列化为字符串，前端统一通过 `toNum()` 转换后格式化显示。

### 数据库

SQLite 单文件存储，GORM `AutoMigrate` 自动维护表结构。数据库文件位于 `data/stock_monitor.db`。

---

## License

MIT
