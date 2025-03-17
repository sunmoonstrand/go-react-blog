# 博客系统

一个使用Go和React构建的现代化博客系统，包括前台展示和后台管理功能。

## 技术栈

### 后端
- Go 1.24
- Gin Web框架
- GORM ORM框架
- PostgreSQL 16数据库
- Redis 7缓存
- JWT认证
- Swagger API文档

### 前端
- React 19
- Ant Design (后台管理UI)
- Chakra UI (前台展示UI)
- TypeScript
- React Router
- React Query
- zustand

## 项目结构

```
/
├── server/             # 后端Go代码
│   ├── api/            # API控制器
│   ├── config/         # 配置文件
│   ├── internal/       # 内部包
│   │   ├── auth/       # 认证相关
│   │   ├── cache/      # 缓存相关
│   │   ├── config/     # 配置加载
│   │   ├── logger/     # 日志工具
│   │   ├── middleware/ # 中间件
│   │   ├── model/      # 数据模型
│   │   ├── service/    # 业务服务
│   │   └── utils/      # 工具函数
│   └── main.go         # 主入口
├── admin/              # 后台管理前端
│   ├── public/         # 静态资源
│   ├── src/            # 源代码
│   │   ├── api/        # API请求
│   │   ├── components/ # 组件
│   │   ├── hooks/      # 自定义钩子
│   │   ├── layouts/    # 布局组件
│   │   ├── pages/      # 页面
│   │   ├── store/      # 状态管理
│   │   ├── types/      # 类型定义
│   │   └── utils/      # 工具函数
│   └── package.json    # 依赖配置
└── web/                # 前台展示前端
    ├── public/         # 静态资源
    ├── src/            # 源代码
    │   ├── api/        # API请求
    │   ├── components/ # 组件
    │   ├── hooks/      # 自定义钩子
    │   ├── layouts/    # 布局组件
    │   ├── pages/      # 页面
    │   ├── store/      # 状态管理
    │   ├── types/      # 类型定义
    │   └── utils/      # 工具函数
    └── package.json    # 依赖配置
```

## 功能特性

### 用户管理
- 用户注册、登录、退出
- 用户角色和权限管理
- 个人资料管理

### 文章管理
- 文章创建、编辑、删除
- 文章分类和标签
- 文章状态管理（草稿、已发布、已归档）
- Markdown编辑器支持

### 分类和标签
- 分类树形结构
- 标签云
- 按分类和标签筛选文章

### 评论系统
- 评论发布和回复
- 评论审核
- 防垃圾评论

### 系统配置
- 站点基本信息配置
- SEO配置
- 社交媒体链接配置

## 安装和运行

### 环境要求
- Go 1.24+
- Node.js 18+
- PostgreSQL 16+
- Redis 7+

### 后端

1. 克隆仓库
```bash
git clone https://github.com/sunmoonstrand/go-react-blog.git
cd blog/server
```

2. 安装依赖
```bash
go mod tidy
```

3. 配置数据库
```bash
# 创建数据库
createdb -U postgres blog

# 导入初始数据
psql -U postgres -d blog -f db/init.sql
```

4. 运行服务
```bash
go run main.go
```

### 后台管理前端

1. 进入目录
```bash
cd ../admin
```

2. 安装依赖
```bash
npm install
```

3. 运行开发服务器
```bash
npm run dev
```

### 前台展示前端

1. 进入目录
```bash
cd ../web
```

2. 安装依赖
```bash
npm install
```

3. 运行开发服务器
```bash
npm run dev
```

## API文档

启动后端服务后，访问 http://localhost:8080/swagger/index.html 查看API文档。

## 部署

### Docker部署

1. 构建镜像
```bash
docker-compose build
```

2. 启动服务
```bash
docker-compose up -d
```

## 贡献指南

1. Fork仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建Pull Request

## 许可证

本项目采用MIT许可证 - 详情请查看 [LICENSE](LICENSE) 文件。