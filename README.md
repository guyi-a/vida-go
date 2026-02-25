# Vida-Go 🚀

**Vida-Go** 是 [vida](https://github.com/guyi-a/vida) 项目的 Go + Python 混合架构版本，一个现代化的智能视频应用平台。

## ✨ 项目特色

### 🏗️ 混合架构设计
- **Go服务**：处理高并发HTTP请求、数据库操作、业务逻辑
- **Python Agent服务**：专注AI功能（LangGraph Agent）
- **微服务架构**：服务解耦，独立部署和扩展

### 🎯 为什么选择混合架构？
- ✅ **性能提升**：Go的高并发处理能力，适合处理大量HTTP请求
- ✅ **资源优化**：Go的内存占用更小，启动速度更快
- ✅ **类型安全**：Go的静态类型系统，减少运行时错误
- ✅ **保留优势**：Python在AI/ML领域的生态优势（LangChain、LangGraph）

## 🎥 核心功能

### 视频管理
- 视频上传（支持大文件分片）
- 视频转码（基于FFmpeg）
- 视频管理（CRUD操作）
- 视频播放（CDN加速）

### 用户系统
- 用户认证（JWT Token）
- 用户管理（注册、登录、资料管理）
- 用户关系（关注/粉丝系统）

### 社交功能
- 评论系统（多级评论）
- 收藏功能
- 互动统计（播放量、点赞数等）

### 智能搜索
- 全文搜索（Elasticsearch + IK分词）
- AI Agent（LangGraph智能搜索）
- 多维度搜索和排序

### AI Agent服务
- LangGraph Agent工作流
- 工具调用（搜索工具等）
- 流式响应
- 上下文记忆

## 🛠️ 技术栈

### Go服务
- **Web框架**：Gin
- **ORM**：GORM
- **数据验证**：go-playground/validator
- **JWT**：golang-jwt/jwt
- **配置管理**：viper
- **日志**：zap
- **Redis**：go-redis
- **MinIO**：minio-go
- **Kafka**：segmentio/kafka-go
- **Elasticsearch**：elastic/go-elasticsearch

### Python Agent服务
- **Web框架**：FastAPI
- **Agent框架**：LangGraph
- **LLM框架**：LangChain
- **数据验证**：Pydantic

### 基础设施
- **数据库**：PostgreSQL（支持pgvector）
- **缓存**：Redis
- **对象存储**：MinIO
- **消息队列**：Kafka
- **搜索引擎**：Elasticsearch
- **视频处理**：FFmpeg

## 📁 项目结构

```
vida-go/
├── cmd/                    # Go应用入口
│   └── api/               # API服务
│       └── main.go
├── internal/              # Go私有代码
│   ├── api/              # API路由层
│   │   ├── handler/      # HTTP处理器
│   │   └── middleware/   # 中间件
│   ├── service/          # 业务逻辑层
│   ├── repository/       # 数据访问层
│   ├── model/            # 数据库模型
│   ├── schema/           # 请求/响应结构
│   ├── client/           # 外部服务客户端
│   │   └── agent_client.go  # Python Agent客户端
│   ├── infra/            # 基础设施
│   │   ├── database/     # 数据库连接
│   │   ├── redis/        # Redis客户端
│   │   ├── minio/        # MinIO客户端
│   │   ├── kafka/        # Kafka客户端
│   │   └── elasticsearch/ # ES客户端
│   ├── task/             # 异步任务
│   └── config/           # 配置管理
├── pkg/                   # 公共库
│   └── utils/            # 工具函数
├── agent-service/         # Python Agent服务
│   ├── app/
│   │   ├── agent/        # Agent核心逻辑
│   │   │   ├── service/
│   │   │   ├── tools/
│   │   │   └── prompts/
│   │   ├── api/          # FastAPI路由
│   │   └── schemas/      # Pydantic模型
│   ├── main.py           # Python服务入口
│   ├── requirements.txt
│   └── Dockerfile
├── configs/               # 配置文件
├── scripts/               # 脚本文件
├── docs/                  # 文档
│   └── migration-plan.md # 改造计划文档
├── api/                   # API文档
│   └── openapi/
├── docker-compose.yml     # Docker编排
├── Dockerfile             # Go服务镜像
├── Makefile              # 构建命令
└── go.mod                # Go模块定义
```

## 🚀 快速开始

### 环境要求

- **Go** 1.23+
- **Python** 3.10+
- **Docker** & **Docker Compose**
- **FFmpeg**（如果本地运行）

### 1. 克隆项目

```bash
git clone https://github.com/guyi-a/vida-go.git
cd vida-go
```

### 2. 环境配置

创建 `.env` 文件：

```env
# 数据库配置
DATABASE_URL=postgresql://guyi:guyi123@postgres:5432/guyi-vida

# Redis配置
REDIS_URL=redis://redis:6379/0

# MinIO配置
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin

# Kafka配置
KAFKA_BOOTSTRAP_SERVERS=kafka:29092

# Elasticsearch配置
ELASTICSEARCH_HOSTS=elasticsearch:9200

# Agent服务配置
AGENT_SERVICE_URL=http://agent-service:8001

# LLM配置（AI Agent）
DASHSCOPE_API_KEY=your_api_key_here
LLM_MODEL=qwen-max

# JWT配置
SECRET_KEY=your_secret_key_here
```

### 3. 使用Docker Compose启动（推荐）

```bash
docker-compose up -d
```

这将启动以下服务：
- **Go API服务**（端口：8000）
- **Python Agent服务**（端口：8001）
- **PostgreSQL**（端口：5432）
- **Redis**（端口：6379）
- **MinIO**（API：9000，Console：9001）
- **Kafka**（端口：9092）
- **Elasticsearch**（端口：9200）

### 4. 访问服务

- **API文档**：http://localhost:8000/docs
- **健康检查**：http://localhost:8000/healthz
- **MinIO控制台**：http://localhost:9001（minioadmin/minioadmin）

## 💻 本地开发

### Go服务开发

```bash
# 1. 安装依赖
go mod download

# 2. 启动基础设施（数据库等）
docker-compose up -d postgres redis minio kafka elasticsearch

# 3. 运行Go服务
go run cmd/api/main.go

# 或使用Makefile
make run
```

### Python Agent服务开发

```bash
# 1. 进入agent-service目录
cd agent-service

# 2. 创建虚拟环境
python -m venv venv

# 3. 激活虚拟环境（Windows）
.\venv\Scripts\Activate.ps1

# 4. 安装依赖
pip install -r requirements.txt

# 5. 运行服务
uvicorn main:app --reload --host 0.0.0.0 --port 8001
```

## 📚 开发指南

### Makefile命令

```bash
make help          # 查看所有命令
make build         # 编译Go程序
make run           # 运行Go服务
make test          # 运行测试
make lint          # 代码检查
make docker-build  # 构建Docker镜像
make docker-up     # 启动所有服务
make docker-down   # 停止所有服务
```

### 代码规范

- **Go代码**：遵循Go官方规范，使用`gofmt`格式化
- **Python代码**：遵循PEP 8规范，使用`black`格式化
- **提交信息**：使用Conventional Commits规范

## 📖 文档

- [改造计划文档](docs/migration-plan.md) - 详细的Python到Go迁移计划
- [API文档](http://localhost:8000/docs) - Swagger API文档
- [架构设计](docs/architecture.md) - 系统架构设计（待补充）

## 🔄 项目状态

当前项目处于**初始化阶段**，正在按照改造计划逐步迁移功能。

### 已完成
- ✅ 项目结构搭建
- ✅ 改造计划文档

### 进行中
- 🚧 基础设施搭建
- 🚧 数据模型定义

### 待开始
- ⏳ 认证模块
- ⏳ 用户模块
- ⏳ 视频模块
- ⏳ 其他模块...

详细进度请查看 [改造计划文档](docs/migration-plan.md)。

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork本项目
2. 创建功能分支（`git checkout -b feature/AmazingFeature`）
3. 提交更改（`git commit -m 'Add some AmazingFeature'`）
4. 推送到分支（`git push origin feature/AmazingFeature`）
5. 开启Pull Request

## 📄 许可证

本项目采用MIT许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🔗 相关项目

- [vida](https://github.com/guyi-a/vida) - 原Python版本

## 👨‍💻 作者

**guyi-a** - [GitHub](https://github.com/guyi-a)

## 📮 联系方式

如有问题或建议，欢迎提Issue或PR！
