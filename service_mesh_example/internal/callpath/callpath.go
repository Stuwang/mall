// Package callpath 负责调用目标的整理，不发起网络请求。
package callpath

import "strings"

const (
	minDepth = 1
	maxDepth = 3
)

// NormalizeTargets 把逗号分隔和重复字段收成去重后的服务名，保留第一次出现的顺序。
func NormalizeTargets(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, item := range in {
		for _, part := range strings.Split(item, ",") {
			name := strings.TrimSpace(part)
			if name == "" {
				continue
			}
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}
	return out
}

// AppendTrace 把当前服务名接到已有路径后面。
func AppendTrace(trace, name string) string {
	if trace == "" {
		return name
	}
	return trace + ">" + name
}

// ClampDepth 把空深度当成 1，并限制扇出层数，避免示例请求打成环。
func ClampDepth(depth int32) int32 {
	if depth < minDepth {
		return minDepth
	}
	if depth > maxDepth {
		return maxDepth
	}
	return depth
}

// Downstream 返回除自己以外还要继续调用的服务。
func Downstream(self string, targets []string) []string {
	out := make([]string, 0, len(targets))
	for _, name := range targets {
		if name == self {
			continue
		}
		out = append(out, name)
	}
	return out
}
