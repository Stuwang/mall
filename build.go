//go:build ignore

// 仓库根目录的开发入口。
//
//	go run build.go proto
//	go run build.go build server
//	go run build.go build allinone
//	go run build.go build kit
//	go run build.go build all
package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "proto":
		if err := generateProto(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "build":
		if err := build(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `用法: go run build.go <命令>

  proto                 生成 pb（server/api、modules/user/api、modules/kit、im/api）
  build server          构建商城 API，输出 bin/server
  build allinone        构建一体包，输出 bin/allinone
  build kit|user|file   编译对应共享模块
  build im              构建 IM，输出 bin/im
  build all             编译共享模块，并构建 server、allinone 与 im
`)
}

type target struct {
	name string
	pkg  string
	bin  string // 为空表示只编译库
}

func targets() []target {
	return []target{
		{name: "kit", pkg: "./modules/kit/..."},
		{name: "user", pkg: "./modules/user/..."},
		{name: "file", pkg: "./modules/file/..."},
		{name: "server", pkg: "./server/cmd", bin: "bin/server"},
		{name: "allinone", pkg: "./server/cmd/allinone", bin: "bin/allinone"},
		{name: "im", pkg: "./im/cmd", bin: "bin/im"},
	}
}

func build(args []string) error {
	if len(args) != 1 {
		usage()
		return fmt.Errorf("build 需要一个目标")
	}
	name := args[0]
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	var list []target
	if name == "all" {
		list = targets()
	} else {
		for _, t := range targets() {
			if t.name == name {
				list = []target{t}
				break
			}
		}
		if len(list) == 0 {
			usage()
			return fmt.Errorf("未知目标 %q", name)
		}
	}
	for _, t := range list {
		if err := buildOne(root, t); err != nil {
			return fmt.Errorf("build %s: %w", t.name, err)
		}
	}
	return nil
}

func buildOne(root string, t target) error {
	args := []string{"build"}
	if t.bin != "" {
		out := t.bin
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
		if err := os.MkdirAll(filepath.Join(root, "bin"), 0o755); err != nil {
			return err
		}
		args = append(args, "-o", out)
	}
	args = append(args, t.pkg)
	return run(root, "go", args...)
}

func generateProto() error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := generateKitProto(root); err != nil {
		return err
	}

	serverDir := filepath.Join(root, "server")
	protos, err := serverProtos(serverDir)
	if err != nil {
		return err
	}
	if len(protos) == 0 {
		return fmt.Errorf("server/api 下没有 proto")
	}

	if err := run(serverDir, "protoc", append([]string{
		"--proto_path=./api",
		"--proto_path=./third_party",
		"--go_out=paths=source_relative:./api",
		"--go-http_out=paths=source_relative:./api",
		"--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:./api",
	}, protos...)...); err != nil {
		return err
	}

	userOut := filepath.ToSlash(filepath.Join("..", "modules", "user", "api"))
	if err := run(serverDir, "protoc",
		"--proto_path="+userOut,
		"--proto_path=./third_party",
		"--go_out=paths=source_relative:"+userOut,
		"--go-http_out=paths=source_relative:"+userOut,
		"--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:"+userOut,
		"user/user.proto",
	); err != nil {
		return err
	}

	if err := generateImProto(serverDir); err != nil {
		return err
	}

	openapiArgs := []string{
		"--proto_path=./api",
		"--proto_path=./third_party",
		"--proto_path=" + userOut,
		"--openapi_out=fq_schema_naming=true:.",
	}
	openapiArgs = append(openapiArgs, protos...)
	openapiArgs = append(openapiArgs, "user/user.proto")
	return run(serverDir, "protoc", openapiArgs...)
}

// generateImProto 生成 im/api。不写入商城的 openapi.yaml。
func generateImProto(serverDir string) error {
	imOut := filepath.ToSlash(filepath.Join("..", "im", "api"))
	userOut := filepath.ToSlash(filepath.Join("..", "modules", "user", "api"))
	return run(serverDir, "protoc",
		"--proto_path="+imOut,
		"--proto_path="+userOut,
		"--proto_path=./third_party",
		"--go_out=paths=source_relative:"+imOut,
		"--go-http_out=paths=source_relative:"+imOut,
		"--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:"+imOut,
		"common/common.proto",
		"im/im.proto",
	)
}

// generateKitProto 生成 modules/kit 下的消息 proto。这些文件没有 HTTP/gRPC service。
func generateKitProto(root string) error {
	kitDir := filepath.Join(root, "modules", "kit")
	protos, err := protoFiles(kitDir)
	if err != nil {
		return err
	}
	if len(protos) == 0 {
		return nil
	}
	kitDir = filepath.ToSlash(kitDir)
	thirdParty := filepath.ToSlash(filepath.Join(root, "server", "third_party"))
	args := []string{
		"--proto_path=" + kitDir,
		"--proto_path=" + thirdParty,
		"--go_out=paths=source_relative:" + kitDir,
	}
	args = append(args, protos...)
	return run(root, "protoc", args...)
}

// protoFiles 返回相对 dir 的 .proto 路径。
func protoFiles(dir string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".proto") {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(out)
	return out, err
}

// serverProtos 返回相对 server 目录的 proto 路径，跳过路径中含 errors 的文件。
func serverProtos(serverDir string) ([]string, error) {
	apiDir := filepath.Join(serverDir, "api")
	var out []string
	err := filepath.WalkDir(apiDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".proto") {
			return nil
		}
		rel, err := filepath.Rel(serverDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if strings.Contains(rel, "errors") {
			return nil
		}
		out = append(out, rel)
		return nil
	})
	sort.Strings(out)
	return out, err
}

func run(dir, name string, args ...string) error {
	fmt.Printf("+ %s\n", strings.Join(append([]string{name}, args...), " "))
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
