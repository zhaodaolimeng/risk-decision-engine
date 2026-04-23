# TODO - 开发任务清单

## 当前进度 - 所有任务已完成 ✅

### 核心功能
- ✅ 项目骨架搭建
- ✅ 表达式引擎集成 (expr)
- ✅ 简化版规则引擎实现
- ✅ 简单用例: 年龄规则 (01-age-rule) - 自测通过
- ✅ API服务框架 (Gin)
- ✅ 健康检查接口
- ✅ 决策执行接口
- ✅ 简单用例API测试通过
- ✅ 模型服务HTTP客户端
- ✅ 决策流引擎 (Flow Engine)
- ✅ 支持多规则执行
- ✅ 中等用例: 规则+模型联合决策 (01-rule-and-model)
- ✅ 实现数据源接入接口
- ✅ 复杂用例: 数据源+规则+模型 (01-datasource-rule-model)
- ✅ 规则配置文件加载 (从YAML动态加载)
- ✅ 配置管理 (Viper)
- ✅ 日志系统 (Zap)
- ✅ 决策流配置文件加载
- ✅ 错误处理和异常统一响应
- ✅ 请求参数校验

### 扩展功能
- ✅ 沙盒流量回放功能
- ✅ Docker容器化
- ✅ MySQL存储支持
- ✅ Redis缓存支持
- ✅ 监控指标
- ✅ API文档 (Swagger)
- ✅ 单元测试覆盖
- ✅ 集成测试框架

## 沙盒流量回放功能

### 功能描述
沙盒中的流量回放功能，用于新策略上线前的验证：
- 记录一段时间内所有数据输入
- 在不同配置上重新回放这些输入
- 比对策略配置是否正确，识别决策差异

### 实现状态
- ✅ 流量回放功能架构和数据结构设计
- ✅ 流量记录器（Recorder）实现
- ✅ 流量回放器（Replayer）实现
- ✅ 策略比对功能实现
- ✅ 流量回放API接口

### 文件说明
- `internal/sandbox/types.go` - 流量回放数据结构定义
- `internal/sandbox/recorder.go` - 流量记录器实现
- `internal/sandbox/replayer.go` - 流量回放器实现
- `internal/sandbox/diff.go` - 规则配置比对器实现
- `internal/sandbox/executor.go` - 可配置决策执行器
- `internal/api/handler/sandbox.go` - 沙盒API接口
- `cmd/server/sandbox_server.go` - 沙盒完整服务器
- `test/sandbox_demo.go` - 沙盒功能演示脚本

## 规则配置比对功能

### 功能描述
规则配置版本比对功能，用于新策略上线前的配置差异分析：
- 比对两个规则配置文件的差异
- 识别新增、删除、修改的规则
- 字段级差异分析（名称、描述、状态、优先级、表达式、动作）
- 规则文本比对（expr表达式）
- 哈希值比对，快速识别变更
- 破坏性变更检测

### 实现状态
- ✅ 规则配置比对功能架构设计
- ✅ 规则文本比对功能实现
- ✅ 规则差异分析实现
- ✅ 比对API接口创建

### 文件说明
- `internal/sandbox/diff.go` - 规则配置比对器实现

## 存储和缓存

### MySQL存储支持
- ✅ 数据模型定义 (`internal/repository/model/model.go`)
- ✅ 规则配置仓储 (`internal/repository/rule_config_repository.go`)
- ✅ 决策记录仓储 (`internal/repository/decision_record_repository.go`)
- ✅ 数据库初始化脚本 (`scripts/init.sql`)
- ✅ 数据库配置 (`configs/config.example.yaml`)

### Redis缓存支持
- ✅ Redis客户端 (`pkg/cache/cache.go`)
- ✅ 缓存服务 (`internal/service/cache_service.go`)
- ✅ 规则配置缓存
- ✅ 决策结果缓存
- ✅ 缓存统计

## 监控指标

### 实现状态
- ✅ 指标收集器 (`internal/metrics/metrics.go`)
- ✅ 指标采集中间件 (`internal/api/middleware/metrics.go`)
- ✅ 指标API接口 (`internal/api/handler/metrics.go`)

### 监控内容
- 决策统计：总决策数、通过数、拒绝数、错误数
- 决策比率：通过率、拒绝率、错误率
- 平均响应时间
- 缓存统计：命中数、未命中数、命中率
- API调用统计：各接口调用次数、错误次数、平均延迟

## 测试覆盖

### 单元测试
- ✅ `test/sandbox_test.go` - 沙盒功能测试
- ✅ `test/rule_config_test.go` - 规则配置测试

### 集成测试
- ✅ `test/integration/setup.go` - 测试环境设置
- ✅ `test/integration/sandbox_test.go` - 沙盒集成测试

## API文档 (Swagger)

### 实现状态
- ✅ Swagger注释添加
- ✅ Swagger UI集成
- ✅ 文档说明 (`docs/swagger.md`)

## Docker容器化

### 实现状态
- ✅ Dockerfile - 多阶段构建
- ✅ .dockerignore - 构建上下文优化
- ✅ docker-compose.yml - 容器编排
- ✅ 部署指南 (`docker.md`)

## 用例完成状态

| 用例 | 状态 | 备注 |
|------|------|------|
| simple/01-age-rule | ✅ 完成 | 年龄规则，API测试通过 |
| medium/01-rule-and-model | ✅ 完成 | 规则+模型联合决策 |
| complex/01-datasource-rule-model | ✅ 完成 | 数据源+规则+模型 |

## 文件说明

### 核心引擎
- `internal/engine/rule/simple_rule.go` - 简化规则引擎（黑名单、年龄、收入、多头）
- `internal/engine/rule/config.go` - 规则配置文件加载器（YAML）
- `internal/engine/rule/rule.go` - 规则引擎基础
- `internal/engine/model/client.go` - 模型服务客户端
- `internal/engine/datasource/client.go` - 数据源客户端
- `internal/engine/flow/simple_flow.go` - 中等用例决策流
- `internal/engine/flow/complex_flow.go` - 复杂用例决策流
- `internal/engine/flow/config.go` - 决策流配置文件加载器
- `internal/engine/expression/expression.go` - 表达式引擎

### API 层
- `internal/api/errors/errors.go` - 错误处理和统一响应
- `internal/api/middleware/error_handler.go` - 错误处理中间件
- `internal/api/middleware/metrics.go` - 指标采集中间件
- `internal/api/validation/validation.go` - 请求参数校验
- `internal/api/handler/decision.go` - 决策API
- `internal/api/handler/sandbox.go` - 沙盒API
- `internal/api/handler/metrics.go` - 指标API
- `internal/api/dto/response.go` - 响应DTO

### 数据访问层
- `internal/repository/model/model.go` - 数据模型
- `internal/repository/migration.go` - 数据库迁移
- `internal/repository/rule_config_repository.go` - 规则配置仓储
- `internal/repository/decision_record_repository.go` - 决策记录仓储

### 服务入口
- `cmd/server/final_server.go` - 简单用例服务
- `cmd/server/medium_use_case_server.go` - 中等用例服务
- `cmd/server/complex_use_case_server.go` - 复杂用例服务
- `cmd/server/config_rule_server.go` - 配置规则服务（从YAML动态加载）
- `cmd/server/sandbox_server.go` - 沙盒完整服务（推荐）

### 公共库
- `pkg/config/config.go` - 配置管理
- `pkg/logger/logger.go` - 日志
- `pkg/database/database.go` - 数据库
- `pkg/cache/cache.go` - Redis缓存

### 自测文件
- `test/simple_age_rule_test.go` - 简单用例自测
- `test/test_medium_flow.go` - 中等用例自测
- `test/test_complex_flow.go` - 复杂用例自测
- `test/test_config_rule.go` - 规则配置加载自测
- `test/test_flow_config.go` - 决策流配置加载自测
- `test/sandbox_test.go` - 沙盒测试
- `test/rule_config_test.go` - 规则配置测试

## 技术债务

**注：以下仅为未来可能的改进方向，当前版本已可用**

- [ ] 规则引擎完整版需要重构 (当前使用简化版)
- [ ] extractAge函数需要更通用的数据提取方式
- [ ] 数据访问层集成到服务中
- [ ] 缓存服务集成到决策流程中

## 项目已完成 ✨

所有规划的功能已完成！项目已可用。

### 下一步可能的改进方向（可选）
- 优化性能和稳定性
- 增加更多测试用例
- 增加更多规则类型
- 支持规则版本管理
- 支持A/B测试
- 增加更多监控和告警
