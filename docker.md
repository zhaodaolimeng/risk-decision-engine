# Docker 部署指南

## 快速开始

### 方式一：Docker Compose（推荐）

```bash
# 构建并启动服务
docker-compose up -d --build

# 查看日志
docker-compose logs -f risk-engine

# 停止服务
docker-compose down
```

### 方式二：单独使用Docker

```bash
# 构建镜像
docker build -t risk-decision-engine:latest .

# 运行容器
docker run -d \
  --name risk-engine \
  -p 8080:8080 \
  -v risk-records:/app/sandbox/records \
  --restart unless-stopped \
  risk-decision-engine:latest
```

## 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 查看API文档
# 访问 http://localhost:8080
```

## Docker Compose 说明

### 启动核心服务（仅风险引擎）

```bash
docker-compose up -d
```

### 启动包含MySQL的完整环境

```bash
docker-compose --profile mysql up -d
```

### 启动包含Redis的完整环境

```bash
docker-compose --profile redis up -d
```

### 启动所有服务（包含MySQL和Redis）

```bash
docker-compose --profile mysql --profile redis up -d
```

## 数据持久化

- **sandbox-records**: 流量回放记录数据
- **mysql-data**: MySQL数据库数据（可选）
- **redis-data**: Redis缓存数据（可选）

### 备份数据

```bash
# 备份sandbox records
docker run --rm -v risk-engine_sandbox-records:/data -v $(pwd):/backup alpine tar czf /backup/sandbox-backup.tar.gz -C /data .
```

## 配置说明

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| TZ | 时区 | Asia/Shanghai |
| GIN_MODE | Gin模式 | release |

### 挂载配置文件

如果需要使用外部配置文件：

```yaml
volumes:
  - ./your-config-path:/app/config:ro
```

## 常见问题

### 端口被占用

修改 `docker-compose.yml` 中的端口映射：

```yaml
ports:
  - "8081:8080"  # 将主机端口改为8081
```

### 查看日志

```bash
# 查看服务日志
docker-compose logs -f risk-engine

# 查看最近100行
docker logs --tail 100 risk-decision-engine
```

### 进入容器

```bash
docker exec -it risk-decision-engine sh
```
