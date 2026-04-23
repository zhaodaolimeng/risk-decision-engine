# Swagger API 文档

## 前置要求

安装 swag 工具：

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

## 生成 Swagger 文档

在项目根目录运行：

```bash
swag init -g cmd/server/sandbox_server.go -o docs
```

这会在 `docs` 目录下生成以下文件：
- `docs.go` - Go 文件
- `swagger.json` - JSON 格式的 API 文档
- `swagger.yaml` - YAML 格式的 API 文档

## 访问 Swagger UI

启动服务后，在浏览器访问：

```
http://localhost:8080/swagger/index.html
```

## 重新生成文档

当 API 有变更时，重新运行：

```bash
swag init -g cmd/server/sandbox_server.go -o docs
```

## Swagger 注释规范

### 基础注释

在包级别添加基础信息：

```go
// Risk Decision Engine API
//
// 风险决策引擎 - 沙盒服务
// 包含规则执行、流量记录、流量回放、规则比对等功能
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
//
// swagger:meta
package main
```

### API 接口注释

```go
// ExecuteDecision godoc
// @Summary 执行决策
// @Description 根据输入数据执行规则决策
// @Tags 决策
// @Accept json
// @Produce json
// @Param request body DecisionRequest true "决策请求"
// @Success 200 {object} object{code=string,message=string,data=DecisionResponse}
// @Failure 400 {object} object{code=string,message=string}
// @Failure 500 {object} object{code=string,message=string}
// @Router /api/v1/decision/execute [post]
func executeDecision(c *gin.Context) {
    // ...
}
```

### API 分组

建议按功能分组：
- `系统` - 健康检查等基础接口
- `规则` - 规则管理接口
- `决策` - 决策执行接口
- `沙盒-记录` - 流量记录接口
- `沙盒-回放` - 流量回放接口
- `沙盒-比对` - 规则比对接口

## Docker 中的 Swagger

Dockerfile 中已包含 swagger 路由，无需额外配置。

## 注意事项

1. 修改 API 后记得重新生成 swagger 文档
2. 确保所有请求/响应结构都有对应的类型定义
3. 使用清晰的 @Summary 和 @Description 描述接口功能
4. 为接口指定正确的 @Tags 便于分类
