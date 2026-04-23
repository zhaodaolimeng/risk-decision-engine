# Risk Decision Engine - 风险决策引擎

一个用于信贷风控的智能风险决策引擎，支持规则引擎、模型服务调用和数据源接入的完整决策流程。

## 项目简介

本项目是一个从零开始构建的信贷风控风险决策引擎，旨在通过自动化流程和智能算法，帮助金融机构进行信贷风险评估和决策。

## 核心特性

- **规则引擎**：基于 expr 表达式引擎的灵活规则配置
- **规则配置文件**：支持 YAML 格式规则配置，动态加载无需硬编码
- **决策流配置文件**：支持 YAML 格式决策流配置，节点和边灵活定义
- **模型服务集成**：支持 HTTP 调用外部 ML 模型服务
- **数据源接入**：支持预加载多维度外部数据
- **决策流编排**：支持从简单到复杂的决策流程编排
- **HTTP API**：基于 Gin 框架的 RESTful API 服务
- **沙盒流量回放**：记录流量并在不同配置下回放，用于新策略上线验证
- **Docker容器化**：一键部署，容器化运行
- **MySQL存储**：规则配置和决策记录持久化
- **Redis缓存**：热点数据缓存，提升性能
- **监控指标**：内置决策统计和API监控
- **Swagger文档**：自动生成API文档
- **完整测试覆盖**：单元测试 + 集成测试

## 项目结构

```
risk-decision-engine/
├── README.md
├── TODO.md                    # 开发任务清单
├── docker.md                  # Docker部署指南
├── go.mod                     # Go 模块定义
├── go.sum
├── Dockerfile                 # Docker镜像构建文件
├── docker-compose.yml         # Docker Compose配置
├── .dockerignore
├── configs/
│   └── config.example.yaml    # 配置文件示例
├── docs/                      # 文档目录
│   ├── README.md              # 文档导航
│   ├── swagger.md             # Swagger文档说明
│   ├── requirements/          # 需求文档
│   └── design/                # 设计文档
├── cmd/                       # 命令行入口
│   ├── api/
│   └── server/                # 服务器入口
│       ├── sandbox_server.go       # 沙盒完整服务（推荐使用）
│       ├── final_server.go         # 简单用例服务
│       ├── medium_use_case_server.go  # 中等用例服务
│       ├── complex_use_case_server.go # 复杂用例服务
│       └── config_rule_server.go    # 配置规则服务（YAML动态加载）
├── internal/                  # 内部代码
│   ├── engine/                # 决策引擎核心
│   │   ├── rule/              # 规则引擎
│   │   │   ├── simple_rule.go # 规则定义与执行
│   │   │   ├── config.go     # 规则配置加载器
│   │   │   └── rule.go
│   │   ├── model/             # 模型客户端
│   │   ├── datasource/        # 数据源客户端
│   │   ├── flow/              # 决策流引擎
│   │   │   ├── simple_flow.go  # 中等用例决策流
│   │   │   ├── complex_flow.go # 复杂用例决策流
│   │   │   └── config.go      # 决策流配置加载器
│   │   └── expression/        # 表达式引擎
│   ├── sandbox/               # 沙盒功能
│   │   ├── types.go           # 数据结构定义
│   │   ├── recorder.go        # 流量记录器
│   │   ├── replayer.go        # 流量回放器
│   │   ├── diff.go            # 规则配置比对
│   │   └── executor.go        # 可配置决策执行器
│   ├── repository/            # 数据访问层
│   │   ├── model/             # 数据模型
│   │   ├── migration.go       # 数据库迁移
│   │   ├── rule_config_repository.go
│   │   └── decision_record_repository.go
│   ├── metrics/               # 监控指标
│   │   └── metrics.go
│   ├── api/                   # API 层
│   │   ├── handler/           # API处理器
│   │   │   ├── decision.go
│   │   │   ├── sandbox.go
│   │   │   └── metrics.go
│   │   ├── middleware/        # 中间件
│   │   │   ├── error_handler.go
│   │   │   ├── metrics.go
│   │   │   └── validation.go
│   │   ├── errors/            # 错误处理
│   │   ├── validation/        # 参数校验
│   │   └── dto/               # 数据传输对象
│   └── service/               # 业务逻辑层
│       ├── rule/              # 规则服务
│       └── cache_service.go   # 缓存服务
├── pkg/                       # 公共库
│   ├── config/                # 配置管理
│   ├── logger/                # 日志
│   ├── database/              # 数据库
│   └── cache/                 # 缓存
├── test/                      # 测试
│   ├── cases/                 # 测试用例定义
│   │   ├── simple/01-age-rule/
│   │   ├── medium/01-rule-and-model/
│   │   └── complex/01-datasource-rule-model/
│   ├── integration/           # 集成测试
│   │   ├── setup.go
│   │   └── sandbox_test.go
│   ├── sandbox_test.go        # 沙盒测试
│   ├── rule_config_test.go    # 规则配置测试
│   ├── sandbox_demo.go        # 沙盒演示
│   ├── simple_age_rule_test.go
│   ├── test_medium_flow.go
│   ├── test_complex_flow.go
│   └── 其他测试...
└── scripts/                   # 脚本
    └── init.sql               # 数据库初始化SQL
```

## 用例说明

### 1. 简单用例 - 年龄规则
- 文件：`test/cases/simple/01-age-rule/`
- 服务：`cmd/server/final_server.go`
- 功能：基于年龄的单一规则决策（21-60岁通过）

### 2. 中等用例 - 规则+模型联合决策
- 文件：`test/cases/medium/01-rule-and-model/`
- 服务：`cmd/server/medium_use_case_server.go`
- 功能：先规则过滤（年龄、收入），再调用模型，根据模型分决策

### 3. 复杂用例 - 数据源+规则+模型完整流程
- 文件：`test/cases/complex/01-datasource-rule-model/`
- 服务：`cmd/server/complex_use_case_server.go`
- 功能：预加载多数据源（用户、征信、多头）→ 规则过滤 → 模型调用 → 综合决策

### 4. 配置规则服务 - YAML动态加载
- 配置文件：`test/cases/*/config/rule.yaml`
- 服务：`cmd/server/config_rule_server.go`
- 功能：从YAML配置文件动态加载规则，支持运行时重载

### 5. 沙盒服务（推荐使用）
- 服务：`cmd/server/sandbox_server.go`
- 功能：完整服务，包含所有API + 沙盒功能
- 推荐用于生产和测试

## 快速开始

### 方式一：使用 Docker（推荐）

```bash
# 构建并启动
docker-compose up -d --build

# 查看日志
docker-compose logs -f risk-engine

# 停止服务
docker-compose down
```

访问 Swagger 文档：http://localhost:8080/swagger/index.html

### 方式二：直接运行 Go 程序

```bash
# 运行沙盒服务（推荐）
go run cmd/server/sandbox_server.go

# 或运行其他服务
go run cmd/server/final_server.go
go run cmd/server/medium_use_case_server.go
go run cmd/server/complex_use_case_server.go
```

### 运行测试

```bash
# 运行所有测试
go test ./... -v

# 运行集成测试
go test ./test/integration -v

# 运行特定测试
go test ./test/integration -run TestSandboxWorkflow -v
```

## 规则配置格式

规则配置使用YAML格式，示例：

```yaml
rules:
  - ruleId: "R001"
    name: "年龄准入规则"
    priority: 100
    status: "ACTIVE"
    condition:
      operator: "AND"
      expressions:
        - field: "age"
          operator: ">="
          value: 21
        - field: "age"
          operator: "<="
          value: 60
    actions:
      true:
        result: "PASS"
      false:
        result: "REJECT"
        reason: "年龄不符合要求"
```

## API 接口

### 基础接口

#### 健康检查
```
GET /health
```

#### 执行决策
```
POST /api/v1/decision/execute
Content-Type: application/json

{
  "requestId": "req_001",
  "businessId": "biz_001",
  "data": {
    "contractId": "CTR_001",
    "applicant": {
      "age": 30,
      "monthlyIncome": 10000
    }
  }
}
```

#### 规则管理
```
GET /api/v1/rules                    # 获取规则列表
POST /api/v1/rules/reload            # 重载规则
```

### 沙盒接口

#### 流量记录
```
POST /api/v1/sandbox/record/start  # 开始记录
POST /api/v1/sandbox/record/stop   # 停止记录
GET  /api/v1/sandbox/record/sessions  # 列出记录会话
GET  /api/v1/sandbox/record/sessions/:id  # 获取会话详情
GET  /api/v1/sandbox/record/sessions/:id/records  # 获取会话记录
```

#### 流量回放
```
POST /api/v1/sandbox/replay/start  # 开始回放（简单）
POST /api/v1/sandbox/replay/start/options  # 开始回放（带选项）
GET  /api/v1/sandbox/replay/sessions  # 列出回放会话
GET  /api/v1/sandbox/replay/sessions/:id  # 获取回放会话
GET  /api/v1/sandbox/replay/sessions/:id/report  # 获取回放报告
```

#### 规则配置比对
```
POST /api/v1/sandbox/diff/rules  # 比对两个规则配置文件
```

请求示例:
```json
{
  "name": "规则v1.0 vs v1.1",
  "oldConfigPath": "test/cases/simple/01-age-rule/config/rule.yaml",
  "newConfigPath": "test/cases/simple/01-age-rule/config/rule_v2.yaml"
}
```

### 监控指标接口

```
GET /api/v1/metrics               # 获取完整指标
GET /api/v1/metrics/health        # 带指标的健康检查
POST /api/v1/metrics/reset        # 重置指标
```

### Swagger 文档

服务启动后访问：http://localhost:8080/swagger/index.html

## 沙盒流量回放功能

### 用途
用于新策略上线前的验证，确保策略变更不会导致意外的决策结果变化。

### 工作流程
1. **记录阶段**：在生产环境记录一段时间的真实流量
2. **回放阶段**：在测试环境使用新配置回放已记录的流量
3. **比对阶段**：自动比对新旧配置的决策结果差异
4. **报告阶段**：生成详细的回放报告，列出所有不匹配项

### 规则配置比对功能

#### 用途
用于规则版本变更时的差异分析，识别配置变更的影响范围。

#### 功能特性
- 规则文本比对：比对规则表达式的变更
- 字段级差异分析：识别哪些字段发生了变化
- 新增/删除/修改规则分类
- 哈希值比对：快速识别规则是否变更
- 破坏性变更检测：识别可能影响决策结果的变更
- 详细比对报告：包含所有变更的完整报告

## 核心模块

### 规则引擎 (`internal/engine/rule/`)
- `SimpleRule` - 简单规则定义
- 内置规则：年龄规则、收入规则、黑名单规则、多头查询规则
- 支持 expr 表达式语法
- `config.go` - YAML规则配置文件加载器，支持动态加载和运行时重载

### 模型客户端 (`internal/engine/model/`)
- `ModelService` - 模型服务接口
- `ModelClient` - HTTP 客户端实现
- 支持 mock 模式便于测试

### 数据源客户端 (`internal/engine/datasource/`)
- `DataSourceService` - 数据源服务接口
- `DataSourceClient` - HTTP 客户端实现
- 支持多数据源预加载

### 决策流引擎 (`internal/engine/flow/`)
- `SimpleFlowEngine` - 中等用例决策流
- `ComplexFlowEngine` - 复杂用例决策流
- 支持数据源加载、规则执行、模型调用、综合决策

### 沙盒流量回放 (`internal/sandbox/`)
- 流量记录器：记录生产流量，保存请求和响应
- 流量回放器：在不同配置下回放历史流量
- 策略比对：自动比对决策结果差异，生成回放报告
- 规则配置比对：比对不同版本规则配置的差异
- 支持会话管理和持久化存储

### 数据访问层 (`internal/repository/`)
- 规则配置持久化
- 决策记录持久化
- 基于GORM的数据库访问

### 缓存服务 (`pkg/cache/`, `internal/service/cache_service.go`)
- Redis缓存客户端
- 规则配置缓存
- 决策结果缓存
- 热点数据缓存
- 缓存命中率统计

### 监控指标 (`internal/metrics/`)
- 决策统计（通过/拒绝/错误数量）
- 缓存统计（命中/未命中/命中率）
- API调用统计
- 平均响应时间统计

### 错误处理和中间件 (`internal/api/`)
- 统一错误响应结构
- 错误处理中间件
- Panic 恢复中间件
- 请求参数校验
- 指标采集中间件

## 技术栈

- **Web 框架**：Gin
- **表达式引擎**：expr-lang/expr
- **配置管理**：Viper
- **日志**：Zap
- **ORM**：GORM
- **数据库**：MySQL
- **缓存**：Redis
- **容器化**：Docker / Docker Compose
- **API文档**：Swagger

## 开发状态

- ✅ 项目骨架搭建
- ✅ 表达式引擎集成
- ✅ 规则引擎实现
- ✅ 简单用例完成
- ✅ 中等用例完成
- ✅ 复杂用例完成
- ✅ 规则配置文件加载（YAML动态加载）
- ✅ 配置管理 (Viper)
- ✅ 日志系统 (Zap)
- ✅ 决策流配置文件加载
- ✅ 错误处理和异常统一响应
- ✅ 请求参数校验
- ✅ 沙盒流量回放功能
- ✅ Docker容器化
- ✅ MySQL存储支持
- ✅ Redis缓存支持
- ✅ 监控指标
- ✅ API文档 (Swagger)
- ✅ 单元测试覆盖
- ✅ 集成测试框架

## 详细文档

- [部署指南](./docker.md)
- [需求文档](./docs/requirements/README.md)
- [设计文档](./docs/design/README.md)
- [开发任务清单](./TODO.md)

## 许可证

待补充...
