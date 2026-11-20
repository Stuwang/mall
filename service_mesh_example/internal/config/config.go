// Package config 从环境变量读取进程身份和下游地址。
package config

import (
	"fmt"
	"os"
	"strings"
)

const defaultPeers = "svc-a=svc-a.mesh.svc.cluster.local:9000," +
	"svc-b=svc-b.mesh.svc.cluster.local:9000," +
	"svc-c=svc-c.mesh.svc.cluster.local:9000"

// Config 是一个进程的运行参数。同一个镜像靠 SERVICE_NAME 区分 biz/svc 实例。
type Config struct {
	Name       string
	HTTPAddr   string
	GRPCAddr   string
	HealthAddr string
	Peers      map[string]string
}

// FromEnv 读取配置。SERVICE_NAME 必须是 biz-* 或 svc-*。
func FromEnv() (Config, error) {
	name := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if name == "" {
		return Config{}, fmt.Errorf("SERVICE_NAME is empty")
	}
	peers, err := parsePeers(getenv("SVC_ADDRS", defaultPeers))
	if err != nil {
		return Config{}, err
	}
	return Config{
		Name:       name,
		HTTPAddr:   getenv("HTTP_ADDR", ":8000"),
		GRPCAddr:   getenv("GRPC_ADDR", ":9000"),
		HealthAddr: getenv("HEALTH_ADDR", ":8081"),
		Peers:      peers,
	}, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func parsePeers(raw string) (map[string]string, error) {
	peers := map[string]string{}
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		name, addr, ok := strings.Cut(item, "=")
		name = strings.TrimSpace(name)
		addr = strings.TrimSpace(addr)
		if !ok || name == "" || addr == "" {
			return nil, fmt.Errorf("invalid SVC_ADDRS item %q", item)
		}
		if _, exists := peers[name]; exists {
			return nil, fmt.Errorf("duplicate peer %q", name)
		}
		peers[name] = addr
	}
	if len(peers) == 0 {
		return nil, fmt.Errorf("SVC_ADDRS is empty")
	}
	return peers, nil
}
