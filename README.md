# Risk Decision Engine - 风险决策引擎

一个用于信贷风控的智能风险决策引擎，支持规则引擎、模型服务调用和数据源接入的完整决策流程。

## 项目简介

本项目是一个从零开始构建的信贷风控风险决策引擎，旨在通过自动化流程和智能算法，帮助金融机构进行信贷风险评估和决策。

## 核心特性

- 规则引擎：基于 expr 表达式引擎的灵活规则配置
- 模型服务集成：支持 HTTP 调用外部 ML 模型服务
- 数据源接入：支持预加载多维度外部数据
- 决策流编排：支持从简单到复杂的决策流程编排
- HTTP API：基于 Gin 框架的 RESTful API 服务

## 项目结构

```
risk-decision-engine/
├── README.md
├── TODO.md                    # 开发任务清单
├── go.mod                     # Go 模块定义
├── go.sum
├── Makefile
├── configs/
│   └── config.yaml            # 配置文件
├── docs/                      # 文档目录
│   ├── requirements/          # 需求文档
│   ├── design/                # 设计文档
│   └── api/                   # API文档
├── cmd/                       # 命令行入口
│   ├── api/
│   └── server/                # 服务器入口
│       ├── final_server.go        # 简单用例服务
│       ├── medium_use_case_server.go  # 中等用例服务
│       └── complex_use_case_server.go # 复杂用例服务
├── internal/                  # 内部代码
│   ├── engine/                # 决策引擎核心
│   │   ├── rule/              # 规则引擎
│   │   ├── model/             # 模型客户端
│   │   ├── datasource/        # 数据源客户端
│   │   ├── flow/              # 决策流引擎
│   │   └── expression/        # 表达式引擎
│   ├── api/                   # API 层
│   │   ├── handler/
│   │   └── dto/
│   └── service/               # 业务逻辑层
├── pkg/                       # 公共库
│   ├── config/
│   ├── logger/
│   └── database/
├── test/                      # 测试用例
│   └── cases/                 # 测试用例定义
│       ├── simple/01-age-rule/
│       ├── medium/01-rule-and-model/
│       └── complex/01-datasource-rule-model/
├── test_simple_age_rule_test.go    # 简单用例自测
├── test_medium_flow.go             # 中等用例自测
└── test_complex_flow.go            # 复杂用例自测
```

## 用例说明

### 1. 简单用例 - 年龄规则
- 文件：`test/cases/simple/01-age-rule/`
- 服务：`cmd/server/final_server.go`
- 自测：`test_simple_age_rule_test.go`
- 功能：基于年龄的单一规则决策（21-60岁通过）

### 2. 中等用例 - 规则+模型联合决策
- 文件：`test/cases/medium/01-rule-and-model/`
- 服务：`cmd/server/medium_use_case_server.go`
- 自测：`test_medium_flow.go`
- 功能：先规则过滤（年龄、收入），再调用模型，根据模型分决策

### 3. 复杂用例 - 数据源+规则+模型完整流程
- 文件：`test/cases/complex/01-datasource-rule-model/`
- 服务：`cmd/server/complex_use_case_server.go`
- 自测：`test_complex_flow.go`
- 功能：预加载多数据源（用户、征信、多头）→ 规则过滤 → 模型调用 → 综合决策

## 快速开始

### 运行简单用例服务

```bash
go run cmd/server/final_server.go
```

### 运行中等用例服务

```bash
go run cmd/server/medium_use_case_server.go
```

### 运行复杂用例服务

```bash
go run cmd/server/complex_use_case_server.go
```

### 运行自测

```bash
# 简单用例自测
go run test_simple_age_rule_test.go

# 中等用例自测
go run test_medium_flow.go

# 复杂用例自测
go run test_complex_flow.go
```

## API 接口

### 健康检查
```
GET /health
```

### 执行决策
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

## 核心模块

### 规则引擎 (`internal/engine/rule/`)
- `SimpleRule` - 简单规则定义
- 内置规则：年龄规则、收入规则、黑名单规则、多头查询规则
- 支持 expr 表达式语法

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

## 技术栈

- Web 框架：Gin
- 表达式引擎：expr-lang/expr
- 配置管理：Viper
- 日志：Zap
- ORM：GORM

## 开发状态

- ✅ 项目骨架搭建
- ✅ 表达式引擎集成
- ✅ 规则引擎实现
- ✅ 简单用例完成
- ✅ 中等用例完成
- ✅ 复杂用例完成

## 开发计划

详见 [TODO.md](./TODO.md)

## 许可证

待补充...
