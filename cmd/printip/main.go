package main

import (
	"fmt"
	"net"
)

func main() {
	fmt.Println("本机可用IP地址:")
	fmt.Println("==================")

	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("获取网卡失败: %v\n", err)
		return
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip.To4() != nil {
				fmt.Printf("  %s\n", ip.String())
			}
		}
	}

	fmt.Println("\n访问地址:")
	fmt.Println("  - 健康检查: http://<IP>:8080/health")
	fmt.Println("  - 决策接口: http://<IP>:8080/api/v1/decision/execute")
}
