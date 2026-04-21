# 仿真测试用例

本目录包含风险决策引擎的仿真测试用例。

## 用例分类

| 分类 | 描述 | 复杂度 |
|------|------|--------|
| [simple](./simple/) | 单一场景、单一规则 | ⭐ |
| [medium](./medium/) | 多个规则、模型调用 | ⭐⭐ |
| [complex](./complex/) | 数据源、多规则、模型集成 | ⭐⭐⭐ |

## 用例结构

每个用例目录包含：

```
case-xxx/
├── case.yaml           # 用例描述
├── config/             # 配置文件
│   ├── rule.yaml      # 规则配置
│   ├── flow.yaml      # 决策流配置
│   ├── strategy.yaml  # 策略配置
│   └── model.yaml     # 模型配置（如需要）
├── input/              # 输入数据
│   └── request.json   # 请求数据
├── output/             # 期望输出
│   └── response.json  # 期望响应
└── mock/               # Mock数据
    ├── datasource/    # 数据源Mock
    └── model/         # 模型服务Mock
```

## 用例运行

待补充...
