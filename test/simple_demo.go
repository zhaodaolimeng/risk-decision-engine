package main

import (
	"fmt"

	"github.com/expr-lang/expr"
)

func main() {
	fmt.Println("=== 简单表达式测试 ===\n")

	// 测试1: 简单比较
	env1 := map[string]interface{}{
		"age": 20,
	}

	result1, err := expr.Eval("age >= 21 && age <= 60", env1)
	if err != nil {
		fmt.Printf("测试1失败: %v\n", err)
	} else {
		fmt.Printf("测试1 - 年龄20岁: %v (应该为 false)\n", result1)
	}

	// 测试2: 通过
	env2 := map[string]interface{}{
		"age": 25,
	}

	result2, err := expr.Eval("age >= 21 && age <= 60", env2)
	if err != nil {
		fmt.Printf("测试2失败: %v\n", err)
	} else {
		fmt.Printf("测试2 - 年龄25岁: %v (应该为 true)\n", result2)
	}

	// 测试3: 边界
	env3 := map[string]interface{}{
		"age": 60,
	}

	result3, err := expr.Eval("age >= 21 && age <= 60", env3)
	if err != nil {
		fmt.Printf("测试3失败: %v\n", err)
	} else {
		fmt.Printf("测试3 - 年龄60岁: %v (应该为 true)\n", result3)
	}

	fmt.Println("\n=== 测试完成 ===")
}
