# TODO - 开发任务清单

## 当前进度
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

## 待完成任务

### 中优先级
- [x] 错误处理和异常统一响应
- [x] 请求参数校验

### 低优先级
- [ ] Docker容器化
- [ ] MySQL存储支持
- [ ] Redis缓存支持
- [ ] 监控指标
- [ ] API文档 (Swagger)
- [ ] 单元测试覆盖
- [ ] 集成测试框架

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
- `internal/engine/model/client.go` - 模型服务客户端
- `internal/engine/datasource/client.go` - 数据源客户端
- `internal/engine/flow/simple_flow.go` - 中等用例决策流
- `internal/engine/flow/complex_flow.go` - 复杂用例决策流
- `internal/engine/flow/config.go` - 决策流配置文件加载器（YAML）

### API 层
- `internal/api/errors/errors.go` - 错误处理和统一响应
- `internal/api/middleware/error_handler.go` - 错误处理中间件
- `internal/api/validation/validation.go` - 请求参数校验

### 服务入口
- `cmd/server/final_server.go` - 简单用例服务
- `cmd/server/medium_use_case_server.go` - 中等用例服务
- `cmd/server/complex_use_case_server.go` - 复杂用例服务
- `cmd/server/config_rule_server.go` - 配置规则服务（从YAML动态加载）

### 自测文件
- `test/simple_age_rule_test.go` - 简单用例自测
- `test_medium_flow.go` - 中等用例自测
- `test_complex_flow.go` - 复杂用例自测
- `test_config_rule.go` - 规则配置加载自测
- `test_flow_config.go` - 决策流配置加载自测

## 技术债务
- [ ] 规则引擎完整版需要重构 (当前使用简化版)
- [x] 需要从配置文件加载规则而非硬编码 (已实现)
- [ ] extractAge函数需要更通用的数据提取方式
