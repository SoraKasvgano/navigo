# Navigation Admin / 网址导航后台管理系统

[English](#english) | [中文](#中文)

---

<a name="english"></a>
## English

### Introduction

Navigation Admin is a lightweight backend management system built with Go, designed for managing navigation page data. It uses SQLite as the database and automatically generates static JSON files for optimal frontend performance.

### Features

- **Go + Gin Framework** - High performance, low resource usage
- **SQLite Database** - No external database server required, single file storage
- **Auto JSON Generation** - Database changes automatically regenerate nav.json
- **Embedded Resources** - Templates and static files compiled into single binary
- **RESTful API** - Standard REST API design
- **Session Authentication** - Secure login with bcrypt password hashing
- **File Upload** - Support for logos and downloadable files
- **Data Import/Export** - Full JSON data import and export
- **Transaction Protection** - All write operations protected by database transactions

### Architecture

```
┌────────────────────────────────────────────────────────────┐
│                    Frontend Browser                         │
│                                                            │
│   Admin Panel (/admin)      Navigation Page (/)            │
│        │                           │                       │
│        ▼                           ▼                       │
│   REST API (/api/admin)    Static nav.json                 │
└────────────────────────────────────────────────────────────┘
                    │                    ▲
                    ▼                    │ Auto-generate
┌────────────────────────────────────────────────────────────┐
│                    Go Backend                               │
│                                                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐ │
│  │ Handlers │◄─│Middleware│◄─│  Models  │◄─│  Database  │ │
│  └──────────┘  └──────────┘  └──────────┘  └────────────┘ │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │              utils/navjson.go                         │ │
│  │         (Generate nav.json on data change)            │ │
│  └──────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────┘
```

### Tech Stack

- **Language:** Go 1.21+
- **Web Framework:** Gin
- **Database:** SQLite (modernc.org/sqlite - pure Go, no CGO)
- **Password:** bcrypt
- **Template:** Go embed (compiled into binary)

### Quick Start

#### Method 1: Docker Deployment (Recommended)

```bash
# Enter project directory
cd goversion

# Create data directories
sudo mkdir -p /home/docker/navigo/{data,uploads,static}
sudo chown -R 1000:1000 /home/docker/navigo

# Start with Docker Compose
docker-compose up -d --build

# View logs
docker-compose logs -f navigo
```

Access at: http://localhost:8787

**For detailed Docker deployment guide, see [DOCKER.md](DOCKER.md)**

#### Method 2: Direct Binary

```bash
# Enter project directory
cd goversion

# Download dependencies
go mod tidy

# Run (development)
go run main.go

# Build (production)
go build -o nav-admin.exe   # Windows
go build -o nav-admin       # Linux/Mac

# Run compiled binary
./nav-admin
```

### Access URLs

| Deployment | URL | Description |
|------------|-----|-------------|
| Docker | http://localhost:8787/ | Navigation page |
| Docker | http://localhost:8787/login | Login page |
| Docker | http://localhost:8787/admin | Admin panel |
| Direct Binary | http://localhost:8080/ | Navigation page |
| Direct Binary | http://localhost:8080/login | Login page |
| Direct Binary | http://localhost:8080/admin | Admin panel |

**Default credentials:** `admin` / `admin`

### Configuration

Environment variables (optional):

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | `8080` | Server port |
| `SERVER_MODE` | `release` | Gin mode (debug/release) |
| `DB_PATH` | `./data/admin.db` | SQLite database path |
| `UPLOAD_PATH` | `./uploads` | Upload directory |
| `NAV_JSON_PATH` | `../static/nav.json` | nav.json output path |
| `SESSION_SECRET` | (built-in) | Session encryption key |

### Project Structure

```
goversion/
├── main.go                 # Entry point, routes
├── go.mod                  # Dependencies
├── config/
│   └── config.go          # Configuration management
├── models/
│   ├── user.go            # User model
│   ├── category.go        # Category model
│   ├── site.go            # Site model
│   ├── announcement.go    # Announcement model
│   └── page_config.go     # Page config model
├── handlers/
│   ├── auth.go            # Login/logout
│   ├── category.go        # Category CRUD
│   ├── site.go            # Site CRUD
│   ├── announcement.go    # Announcement CRUD
│   ├── upload.go          # File upload
│   └── nav.go             # Navigation data & import/export
├── middleware/
│   └── auth.go            # Authentication middleware
├── utils/
│   ├── database.go        # Database initialization
│   ├── response.go        # Unified response format
│   └── navjson.go         # JSON file generator
├── templates/
│   ├── index.html         # Navigation page
│   ├── login.html         # Login page
│   └── admin.html         # Admin panel
└── static/                 # All resources embedded into binary
    ├── style.css          # Main stylesheet
    ├── themify-icons.css  # Icon font stylesheet
    ├── jquery.js          # jQuery library
    ├── nav-go.js          # Navigation page JavaScript
    ├── logo.png           # Logo image
    ├── logo.svg           # Logo vector
    ├── nav.json           # Auto-generated navigation data
    ├── fonts/             # Themify icon fonts
    │   ├── themify.eot
    │   ├── themify.svg
    │   ├── themify.ttf
    │   └── themify.woff
    └── lunar/
        └── lunar.js       # Chinese lunar calendar library
```

### API Reference

#### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/login` | Login |
| GET | `/api/check-auth` | Check login status |
| POST | `/api/admin/logout` | Logout |

#### Categories
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/categories` | Get all categories |
| GET | `/api/admin/categories/:id` | Get single category |
| POST | `/api/admin/categories` | Create category |
| PUT | `/api/admin/categories/:id` | Update category |
| DELETE | `/api/admin/categories/:id` | Delete category |
| PUT | `/api/admin/categories/sort` | Update sort order |

#### Sites
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/categories/:id/sites` | Get sites by category |
| GET | `/api/admin/sites/:id` | Get single site |
| POST | `/api/admin/sites` | Create site |
| PUT | `/api/admin/sites/:id` | Update site |
| DELETE | `/api/admin/sites/:id` | Delete site |
| PUT | `/api/admin/sites/sort` | Update sort order |

#### Announcements
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/announcements` | Get all announcements |
| POST | `/api/admin/announcements` | Create announcement |
| PUT | `/api/admin/announcements/:id` | Update announcement |
| DELETE | `/api/admin/announcements/:id` | Delete announcement |
| GET | `/api/admin/announcement-config` | Get config |
| PUT | `/api/admin/announcement-config` | Update config |

#### Page Config
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/page-config` | Get page config |
| PUT | `/api/admin/page-config` | Update page config |

#### File Management
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/upload` | Upload file |
| DELETE | `/api/admin/upload` | Delete file |
| GET | `/api/admin/files` | List files |

#### Data
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/export` | Export all data |
| POST | `/api/admin/import` | Import data |

### Response Format

```json
{
  "code": 0,
  "message": "success",
  "data": { }
}
```

- `code: 0` = Success
- `code: 400` = Bad request
- `code: 401` = Unauthorized
- `code: 404` = Not found
- `code: 500` = Server error

### Design Principles

1. **Backend as Source of Truth** - Database is the single source, frontend only displays
2. **Read/Write Separation** - Frontend reads static JSON, backend writes to DB and generates JSON
3. **Static-First** - Navigation pages are read-heavy, static files provide best performance
4. **Single Binary** - All resources embedded, one file deployment

### License

MIT License

---

<a name="中文"></a>
## 中文

### 项目简介

网址导航后台管理系统是一个使用Go语言构建的轻量级后台管理系统，用于管理导航页面数据。使用SQLite作为数据库，自动生成静态JSON文件以获得最佳前端性能。

### 功能特性

- **Go + Gin框架** - 高性能，低资源占用
- **SQLite数据库** - 无需外部数据库服务器，单文件存储
- **自动JSON生成** - 数据变更自动重新生成nav.json
- **资源嵌入** - 模板和静态文件编译到单个二进制文件
- **RESTful API** - 标准REST API设计
- **Session认证** - 安全登录，bcrypt密码加密
- **文件上传** - 支持Logo和可下载文件上传
- **数据导入导出** - 完整的JSON数据导入导出
- **事务保护** - 所有写操作使用数据库事务保护

### 系统架构

```
┌────────────────────────────────────────────────────────────┐
│                      前端浏览器                              │
│                                                            │
│   管理后台 (/admin)           导航页面 (/)                  │
│        │                           │                       │
│        ▼                           ▼                       │
│   REST API (/api/admin)     静态 nav.json                  │
└────────────────────────────────────────────────────────────┘
                    │                    ▲
                    ▼                    │ 自动生成
┌────────────────────────────────────────────────────────────┐
│                      Go 后端                                │
│                                                            │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐ │
│  │ Handlers │◄─│Middleware│◄─│  Models  │◄─│  Database  │ │
│  └──────────┘  └──────────┘  └──────────┘  └────────────┘ │
│                                                            │
│  ┌──────────────────────────────────────────────────────┐ │
│  │              utils/navjson.go                         │ │
│  │           (数据变更时生成nav.json)                     │ │
│  └──────────────────────────────────────────────────────┘ │
└────────────────────────────────────────────────────────────┘
```

### 技术栈

- **语言：** Go 1.21+
- **Web框架：** Gin
- **数据库：** SQLite (modernc.org/sqlite - 纯Go实现，无CGO)
- **密码加密：** bcrypt
- **模板：** Go embed（编译到二进制文件）

### 快速开始

#### 方式一：Docker部署（推荐）

```bash
# 进入项目目录
cd goversion

# 创建数据目录
sudo mkdir -p /home/docker/navigo/{data,uploads,static}
sudo chown -R 1000:1000 /home/docker/navigo

# 使用 Docker Compose 启动
docker-compose up -d --build

# 查看日志
docker-compose logs -f navigo
```

访问地址：http://localhost:8787

**详细的Docker部署指南请查看 [DOCKER.md](DOCKER.md)**

#### 方式二：直接运行二进制文件

```bash
# 进入项目目录
cd goversion

# 下载依赖
go mod tidy

# 运行（开发模式）
go run main.go

# 编译（生产部署）
go build -o nav-admin.exe   # Windows
go build -o nav-admin       # Linux/Mac

# 运行编译后的程序
./nav-admin
```

### 访问地址

| 部署方式 | 地址 | 说明 |
|---------|-----|------|
| Docker | http://localhost:8787/ | 导航页面 |
| Docker | http://localhost:8787/login | 登录页面 |
| Docker | http://localhost:8787/admin | 管理后台 |
| 直接运行 | http://localhost:8080/ | 导航页面 |
| 直接运行 | http://localhost:8080/login | 登录页面 |
| 直接运行 | http://localhost:8080/admin | 管理后台 |

**默认账号：** `admin` / `admin`

### 配置说明

环境变量（可选）：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVER_PORT` | `8080` | 服务器端口 |
| `SERVER_MODE` | `release` | Gin模式（debug/release）|
| `DB_PATH` | `./data/admin.db` | SQLite数据库路径 |
| `UPLOAD_PATH` | `./uploads` | 上传文件目录 |
| `NAV_JSON_PATH` | `../static/nav.json` | nav.json输出路径 |
| `SESSION_SECRET` | (内置默认) | Session加密密钥 |

### 项目结构

```
goversion/
├── main.go                 # 入口文件，路由配置
├── go.mod                  # 依赖管理
├── config/
│   └── config.go          # 配置管理
├── models/
│   ├── user.go            # 用户模型
│   ├── category.go        # 分类模型
│   ├── site.go            # 站点模型
│   ├── announcement.go    # 公告模型
│   └── page_config.go     # 页面配置模型
├── handlers/
│   ├── auth.go            # 登录/登出
│   ├── category.go        # 分类增删改查
│   ├── site.go            # 站点增删改查
│   ├── announcement.go    # 公告增删改查
│   ├── upload.go          # 文件上传
│   └── nav.go             # 导航数据、导入导出
├── middleware/
│   └── auth.go            # 认证中间件
├── utils/
│   ├── database.go        # 数据库初始化
│   ├── response.go        # 统一响应格式
│   └── navjson.go         # JSON文件生成器
├── templates/
│   ├── index.html         # 导航页面
│   ├── login.html         # 登录页面
│   └── admin.html         # 管理后台
└── static/                 # 所有资源嵌入到二进制文件
    ├── style.css          # 主样式表
    ├── themify-icons.css  # 图标字体样式
    ├── jquery.js          # jQuery库
    ├── nav-go.js          # 导航页面JavaScript
    ├── logo.png           # Logo图片
    ├── logo.svg           # Logo矢量图
    ├── nav.json           # 自动生成的导航数据
    ├── fonts/             # Themify图标字体
    │   ├── themify.eot
    │   ├── themify.svg
    │   ├── themify.ttf
    │   └── themify.woff
    └── lunar/
        └── lunar.js       # 中国农历库
```

### API接口文档

#### 认证相关
| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/login` | 登录 |
| GET | `/api/check-auth` | 检查登录状态 |
| POST | `/api/admin/logout` | 登出 |

#### 分类管理
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/admin/categories` | 获取所有分类 |
| GET | `/api/admin/categories/:id` | 获取单个分类 |
| POST | `/api/admin/categories` | 创建分类 |
| PUT | `/api/admin/categories/:id` | 更新分类 |
| DELETE | `/api/admin/categories/:id` | 删除分类 |
| PUT | `/api/admin/categories/sort` | 更新排序 |

#### 站点管理
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/admin/categories/:id/sites` | 获取分类下的站点 |
| GET | `/api/admin/sites/:id` | 获取单个站点 |
| POST | `/api/admin/sites` | 创建站点 |
| PUT | `/api/admin/sites/:id` | 更新站点 |
| DELETE | `/api/admin/sites/:id` | 删除站点 |
| PUT | `/api/admin/sites/sort` | 更新排序 |

#### 公告管理
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/admin/announcements` | 获取所有公告 |
| POST | `/api/admin/announcements` | 创建公告 |
| PUT | `/api/admin/announcements/:id` | 更新公告 |
| DELETE | `/api/admin/announcements/:id` | 删除公告 |
| GET | `/api/admin/announcement-config` | 获取公告配置 |
| PUT | `/api/admin/announcement-config` | 更新公告配置 |

#### 页面配置
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/admin/page-config` | 获取页面配置 |
| PUT | `/api/admin/page-config` | 更新页面配置 |

#### 文件管理
| 方法 | 端点 | 说明 |
|------|------|------|
| POST | `/api/admin/upload` | 上传文件 |
| DELETE | `/api/admin/upload` | 删除文件 |
| GET | `/api/admin/files` | 获取文件列表 |

#### 数据管理
| 方法 | 端点 | 说明 |
|------|------|------|
| GET | `/api/admin/export` | 导出所有数据 |
| POST | `/api/admin/import` | 导入数据 |

### 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": { }
}
```

- `code: 0` = 成功
- `code: 400` = 请求错误
- `code: 401` = 未授权
- `code: 404` = 未找到
- `code: 500` = 服务器错误

### 设计理念

1. **后端为数据源** - 数据库是唯一数据源，前端只负责展示
2. **读写分离** - 前端读取静态JSON，后端写入数据库并生成JSON
3. **静态化优先** - 导航页读多写少，静态文件性能最优
4. **单文件部署** - 所有资源嵌入，单个文件即可运行

### 数据库表结构

```sql
-- 用户表
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);

-- 分类表
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    id_str TEXT NOT NULL,
    classify TEXT NOT NULL,
    icon TEXT NOT NULL,
    sort_no INTEGER DEFAULT 0
);

-- 站点表
CREATE TABLE sites (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cat_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    href TEXT NOT NULL,
    description TEXT,
    logo TEXT,
    sort_no INTEGER DEFAULT 0,
    FOREIGN KEY (cat_id) REFERENCES categories(id) ON DELETE CASCADE
);

-- 公告表
CREATE TABLE announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    content TEXT NOT NULL
);

-- 公告配置表
CREATE TABLE announcement_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    interval INTEGER DEFAULT 5000
);

-- 页面配置表
CREATE TABLE page_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    title TEXT,
    subtitle TEXT,
    logo TEXT,
    footer_text TEXT,
    icp TEXT
);
```

### nav.json输出格式

```json
[
  {
    "_id": "announcement_config",
    "type": "announcement_config",
    "interval": 5000,
    "announcements": [
      {"id": 1, "timestamp": "2025-01-01", "content": "公告内容"}
    ]
  },
  {
    "_id": "category_id",
    "classify": "分类名称",
    "icon": "ti-cloud",
    "sites": [
      {
        "name": "站点名称",
        "href": "https://example.com",
        "desc": "站点描述",
        "logo": "static/icons/logo.png"
      }
    ]
  }
]
```

### 安全说明

- 密码使用bcrypt加密存储
- nav.json不包含任何敏感信息（用户、密码等）
- 管理接口需要Session认证
- 文件上传有类型和大小限制
- 所有数据库写操作使用事务保护

### 许可证

GPLv3
