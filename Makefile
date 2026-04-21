.PHONY: help build run test tidy clean

# 显示帮助
help:
	@echo "Usage:"
	@echo "  make build       - 编译项目"
	@echo "  make run         - 运行 API 服务"
	@echo "  make test        - 运行测试"
	@echo "  make tidy        - 整理依赖"
	@echo "  make clean       - 清理构建文件"

# 编译项目
build:
	@echo "Building..."
	go build -o bin/api.exe ./cmd/api
	go build -o bin/admin.exe ./cmd/admin

# 运行 API 服务
run:
	@echo "Running API server..."
	go run ./cmd/api

# 运行测试
test:
	@echo "Running tests..."
	go test -v ./...

# 整理依赖
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# 清理
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf tmp/
