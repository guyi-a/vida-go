# Vida-Go Makefile

# 变量定义
APP_NAME=vida-api
MAIN_PATH=./cmd/api
BUILD_DIR=./bin
DOCKER_IMAGE=vida-go
DOCKER_COMPOSE=docker-compose

# Go相关变量
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# 颜色输出
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: all build clean test run help docker-build docker-up docker-down lint fmt deps

# 默认目标
all: deps build

# 显示帮助信息
help:
	@echo "$(GREEN)Vida-Go Makefile Commands:$(NC)"
	@echo "  $(YELLOW)make build$(NC)         - 编译Go程序"
	@echo "  $(YELLOW)make run$(NC)           - 运行Go服务"
	@echo "  $(YELLOW)make test$(NC)          - 运行测试"
	@echo "  $(YELLOW)make clean$(NC)         - 清理编译文件"
	@echo "  $(YELLOW)make deps$(NC)          - 下载依赖"
	@echo "  $(YELLOW)make fmt$(NC)           - 格式化代码"
	@echo "  $(YELLOW)make lint$(NC)          - 代码检查"
	@echo "  $(YELLOW)make docker-build$(NC)  - 构建Docker镜像"
	@echo "  $(YELLOW)make docker-up$(NC)     - 启动所有服务"
	@echo "  $(YELLOW)make docker-down$(NC)   - 停止所有服务"
	@echo "  $(YELLOW)make docker-logs$(NC)   - 查看服务日志"

# 下载依赖
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy

# 编译
build: deps
	@echo "$(GREEN)Building $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(APP_NAME)$(NC)"

# 编译（带版本信息）
build-release:
	@echo "$(GREEN)Building $(APP_NAME) for release...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "$(GREEN)Release build complete: $(BUILD_DIR)/$(APP_NAME)$(NC)"

# 运行
run:
	@echo "$(GREEN)Running $(APP_NAME)...$(NC)"
	$(GOCMD) run $(MAIN_PATH)/main.go

# 运行（带热重载，需要安装air）
dev:
	@echo "$(GREEN)Running with hot reload...$(NC)"
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "$(RED)Error: air not found. Install it with: go install github.com/air-verse/air@latest$(NC)"; \
	fi

# 测试
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests complete$(NC)"

# 测试覆盖率
test-coverage: test
	@echo "$(GREEN)Generating coverage report...$(NC)"
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

# 清理
clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

# 格式化代码
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GOFMT) -s -w .
	@echo "$(GREEN)Format complete$(NC)"

# 代码检查
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v $(GOLINT) > /dev/null; then \
		$(GOLINT) run ./...; \
	else \
		echo "$(RED)Error: golangci-lint not found. Install it from: https://golangci-lint.run/usage/install/$(NC)"; \
	fi

# Docker相关命令

# 构建Docker镜像
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(DOCKER_IMAGE):latest .
	@echo "$(GREEN)Docker image built: $(DOCKER_IMAGE):latest$(NC)"

# 启动所有服务
docker-up:
	@echo "$(GREEN)Starting all services...$(NC)"
	$(DOCKER_COMPOSE) up -d
	@echo "$(GREEN)Services started$(NC)"
	@echo "$(YELLOW)API: http://localhost:8000$(NC)"
	@echo "$(YELLOW)Agent Service: http://localhost:8001$(NC)"

# 停止所有服务
docker-down:
	@echo "$(YELLOW)Stopping all services...$(NC)"
	$(DOCKER_COMPOSE) down
	@echo "$(GREEN)Services stopped$(NC)"

# 停止并删除数据卷
docker-down-volumes:
	@echo "$(RED)Stopping services and removing volumes...$(NC)"
	$(DOCKER_COMPOSE) down -v
	@echo "$(GREEN)Services stopped and volumes removed$(NC)"

# 查看服务日志
docker-logs:
	$(DOCKER_COMPOSE) logs -f

# 查看特定服务日志
docker-logs-api:
	$(DOCKER_COMPOSE) logs -f api

docker-logs-agent:
	$(DOCKER_COMPOSE) logs -f agent-service

# 重启服务
docker-restart:
	@echo "$(YELLOW)Restarting services...$(NC)"
	$(DOCKER_COMPOSE) restart
	@echo "$(GREEN)Services restarted$(NC)"

# 查看服务状态
docker-ps:
	$(DOCKER_COMPOSE) ps

# 进入容器
docker-exec-api:
	$(DOCKER_COMPOSE) exec api sh

docker-exec-agent:
	$(DOCKER_COMPOSE) exec agent-service bash

# 数据库相关

# 数据库迁移（待实现）
migrate-up:
	@echo "$(GREEN)Running database migrations...$(NC)"
	@echo "$(YELLOW)TODO: Implement migration$(NC)"

migrate-down:
	@echo "$(YELLOW)Rolling back database migrations...$(NC)"
	@echo "$(YELLOW)TODO: Implement migration$(NC)"

# 安装开发工具
install-tools:
	@echo "$(GREEN)Installing development tools...$(NC)"
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "$(GREEN)Tools installed$(NC)"

# 生成API文档（Swagger）
swagger:
	@echo "$(GREEN)Generating Swagger documentation...$(NC)"
	@if command -v swag > /dev/null; then \
		swag init -g $(MAIN_PATH)/main.go -o ./docs/swagger; \
		echo "$(GREEN)Swagger docs generated: ./docs/swagger$(NC)"; \
	else \
		echo "$(RED)Error: swag not found. Install it with: go install github.com/swaggo/swag/cmd/swag@latest$(NC)"; \
	fi
