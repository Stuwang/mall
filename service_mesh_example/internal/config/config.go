// Package config 从环境变量读取进程身份和网关地址。
package config

import (
	"fmt"
	"os"
	"strings"
)

const defaultGateway = "mesh-eastwest.mesh.svc.cluster.local:9090"

// Config 是一个进程的运行参数。同一个镜像靠 SERVICE_NAME 区分 biz/svc 实例。
type Config struct {
	Name        string
	HTTPAddr    string
	GRPCAddr    string
	HealthAddr  string
	GatewayAddr string
}

// FromEnv 读取配置。SERVICE_NAME 必须是 biz-* 或 svc-*。
func FromEnv() (Config, error) {
	name := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if name == "" {
		return Config{}, fmt.Errorf("SERVICE_NAME is empty")
	}
	return Config{
		Name:        name,
		HTTPAddr:    getenv("HTTP_ADDR", ":8000"),
		GRPCAddr:    getenv("GRPC_ADDR", ":9000"),
		HealthAddr:  getenv("HEALTH_ADDR", ":8081"),
		GatewayAddr: getenv("GATEWAY_ADDR", defaultGateway),
	}, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
