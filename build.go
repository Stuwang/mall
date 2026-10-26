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
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
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

  proto                 用 bin/ 里的固定版本工具生成 pb（缺少时自动下载）
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
	binDir, err := ensureProtoTools(root)
	if err != nil {
		return err
	}
	if err := generateKitProto(root, binDir); err != nil {
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

	if err := runProtoc(serverDir, binDir, append([]string{
		"--proto_path=./api",
		"--proto_path=./third_party",
		"--go_out=paths=source_relative:./api",
		"--go-http_out=paths=source_relative:./api",
		"--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:./api",
	}, protos...)...); err != nil {
		return err
	}

	userOut := filepath.ToSlash(filepath.Join("..", "modules", "user", "api"))
	if err := runProtoc(serverDir, binDir,
		"--proto_path="+userOut,
		"--proto_path=./third_party",
		"--go_out=paths=source_relative:"+userOut,
		"--go-http_out=paths=source_relative:"+userOut,
		"--go-grpc_out=require_unimplemented_servers=false,paths=source_relative:"+userOut,
		"user/user.proto",
	); err != nil {
		return err
	}

	if err := generateImProto(serverDir, binDir); err != nil {
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
	return runProtoc(serverDir, binDir, openapiArgs...)
}

// generateImProto 生成 im/api。不写入商城的 openapi.yaml。
func generateImProto(serverDir, binDir string) error {
	imOut := filepath.ToSlash(filepath.Join("..", "im", "api"))
	userOut := filepath.ToSlash(filepath.Join("..", "modules", "user", "api"))
	return runProtoc(serverDir, binDir,
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
func generateKitProto(root, binDir string) error {
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
	return runProtoc(root, binDir, args...)
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

// protoc 与插件版本。改这里后下次 proto 会重新安装。
// protoc 用最新稳定版，跳过 rc。protoc-gen-go-http 用 Kratos v2 插件的最新提交；
// v3 插件会生成 kratos/v3 的 import，当前仓库还是 kratos v2。
const protocRelease = "36.2"

const protoToolsStampName = ".proto-tools.stamp"

type protoPlugin struct {
	name    string
	pkg     string
	version string
}

func protoPlugins() []protoPlugin {
	return []protoPlugin{
		{
			name:    "protoc-gen-go",
			pkg:     "google.golang.org/protobuf/cmd/protoc-gen-go",
			version: "v1.36.12",
		},
		{
			name:    "protoc-gen-go-grpc",
			pkg:     "google.golang.org/grpc/cmd/protoc-gen-go-grpc",
			version: "v1.6.2",
		},
		// 该模块没有发行 tag。这个伪版本是默认分支最新提交，二进制打印 v2.9.2。
		{
			name:    "protoc-gen-go-http",
			pkg:     "github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2",
			version: "v2.0.0-20260404020628-f149714c1d54",
		},
		{
			name:    "protoc-gen-openapi",
			pkg:     "github.com/kollalabs/protoc-gen-openapi",
			version: "v0.0.28",
		},
	}
}

// ensureProtoTools 把 protoc 和插件装到仓库 bin/，并返回该目录。
// 已安装且版本戳一致时直接复用。
func ensureProtoTools(root string) (string, error) {
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return "", err
	}
	stampPath := filepath.Join(binDir, protoToolsStampName)
	want := protoToolsStamp()
	if protoToolsReady(binDir, stampPath, want) {
		return binDir, nil
	}
	fmt.Println("安装 proto 工具到 bin/ ...")
	if err := installProtoc(binDir); err != nil {
		return "", err
	}
	for _, p := range protoPlugins() {
		if err := installGoPlugin(binDir, p); err != nil {
			return "", err
		}
	}
	if err := os.WriteFile(stampPath, []byte(want), 0o644); err != nil {
		return "", err
	}
	return binDir, nil
}

func protoToolsStamp() string {
	var b strings.Builder
	fmt.Fprintf(&b, "protoc %s\n", protocRelease)
	for _, p := range protoPlugins() {
		fmt.Fprintf(&b, "%s %s\n", p.pkg, p.version)
	}
	fmt.Fprintf(&b, "os %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return b.String()
}

func protoToolsReady(binDir, stampPath, want string) bool {
	got, err := os.ReadFile(stampPath)
	if err != nil || string(got) != want {
		return false
	}
	names := []string{"protoc"}
	for _, p := range protoPlugins() {
		names = append(names, p.name)
	}
	for _, name := range names {
		if _, err := os.Stat(filepath.Join(binDir, binName(name))); err != nil {
			return false
		}
	}
	return true
}

func installProtoc(binDir string) error {
	asset, err := protocZipAsset()
	if err != nil {
		return err
	}
	url := fmt.Sprintf(
		"https://github.com/protocolbuffers/protobuf/releases/download/v%s/protoc-%s-%s.zip",
		protocRelease, protocRelease, asset,
	)
	fmt.Printf("+ download %s\n", url)
	zipPath, err := downloadToTemp(url)
	if err != nil {
		return err
	}
	defer os.Remove(zipPath)
	return extractProtoc(zipPath, binDir)
}

func protocZipAsset() (string, error) {
	switch runtime.GOOS {
	case "windows":
		switch runtime.GOARCH {
		case "amd64":
			return "win64", nil
		case "386":
			return "win32", nil
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "linux-x86_64", nil
		case "arm64":
			return "linux-aarch_64", nil
		case "386":
			return "linux-x86_32", nil
		}
	case "darwin":
		switch runtime.GOARCH {
		case "amd64":
			return "osx-x86_64", nil
		case "arm64":
			return "osx-aarch_64", nil
		}
	}
	return "", fmt.Errorf("没有 protoc %s 的预编译包: %s/%s", protocRelease, runtime.GOOS, runtime.GOARCH)
}

func downloadToTemp(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "mall-build")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("下载 protoc: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载 protoc: %s", resp.Status)
	}

	tmp, err := os.CreateTemp("", "protoc-*.zip")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()
	const maxZip = 64 << 20
	n, err := io.Copy(tmp, io.LimitReader(resp.Body, maxZip+1))
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		os.Remove(tmpName)
		return "", fmt.Errorf("保存 protoc 压缩包: %w", err)
	}
	if n > maxZip {
		os.Remove(tmpName)
		return "", fmt.Errorf("protoc 压缩包超过 %d 字节", maxZip)
	}
	return tmpName, nil
}

func extractProtoc(zipPath, binDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("打开 protoc 压缩包: %w", err)
	}
	defer r.Close()

	want := "bin/" + binName("protoc")
	for _, f := range r.File {
		if f.FileInfo().IsDir() || filepath.ToSlash(f.Name) != want {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = writeExecutable(filepath.Join(binDir, binName("protoc")), rc)
		rc.Close()
		return err
	}
	return fmt.Errorf("protoc 压缩包里没有 %s", want)
}

func installGoPlugin(binDir string, p protoPlugin) error {
	spec := p.pkg + "@" + p.version
	fmt.Printf("+ go install %s\n", spec)
	cmd := exec.Command("go", "install", spec)
	cmd.Dir = binDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = replaceEnv(os.Environ(), "GOBIN="+binDir, "GOWORK=off")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("安装 %s: %w", p.name, err)
	}
	if _, err := os.Stat(filepath.Join(binDir, binName(p.name))); err != nil {
		return fmt.Errorf("安装 %s 后没有找到 %s", p.name, binName(p.name))
	}
	return nil
}

func runProtoc(dir, binDir string, args ...string) error {
	protoc := filepath.Join(binDir, binName("protoc"))
	full := make([]string, 0, len(args)+len(protoPlugins()))
	for _, p := range protoPlugins() {
		plugin := filepath.ToSlash(filepath.Join(binDir, binName(p.name)))
		full = append(full, "--plugin="+p.name+"="+plugin)
	}
	full = append(full, args...)
	return run(dir, protoc, full...)
}

func binName(name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func writeExecutable(dest string, r io.Reader) error {
	tmp := dest + ".download"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(f, r)
	closeErr := f.Close()
	if copyErr != nil {
		os.Remove(tmp)
		return copyErr
	}
	if closeErr != nil {
		os.Remove(tmp)
		return closeErr
	}
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}

// replaceEnv 用 overrides（KEY=VALUE）覆盖同名环境变量。Go 的 Getenv 取第一个，所以不能只追加。
func replaceEnv(base []string, overrides ...string) []string {
	vals := make(map[string]string, len(overrides))
	order := make([]string, 0, len(overrides))
	for _, o := range overrides {
		k, v, ok := strings.Cut(o, "=")
		if !ok {
			continue
		}
		key := envKey(k)
		if _, exists := vals[key]; !exists {
			order = append(order, key)
		}
		vals[key] = k + "=" + v
	}
	out := make([]string, 0, len(base)+len(overrides))
	seen := make(map[string]bool, len(vals))
	for _, e := range base {
		k, _, ok := strings.Cut(e, "=")
		if !ok {
			out = append(out, e)
			continue
		}
		key := envKey(k)
		repl, replace := vals[key]
		if !replace {
			out = append(out, e)
			continue
		}
		if !seen[key] {
			out = append(out, repl)
			seen[key] = true
		}
	}
	for _, key := range order {
		if !seen[key] {
			out = append(out, vals[key])
		}
	}
	return out
}

func envKey(k string) string {
	if runtime.GOOS == "windows" {
		return strings.ToUpper(k)
	}
	return k
}
