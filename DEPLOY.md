# 生产一体部署（单进程三端口）

开发时仍分别启动 `server` / `admin` / `web_app`。生产可用 `cmd/allinone` 把 API 与两个前端静态站打进一个 Go 二进制。

## 端口

| 服务 | 默认地址 | 说明 |
|------|----------|------|
| API | `:8080` | 来自配置 `server.http_addr` |
| Admin | `:8081` | `-admin-addr` 可改 |
| H5 (web_app) | `:8082` | `-web-addr` 可改 |

Admin / H5 端口会反代 API 路径（与各自 Vite proxy 一致），前端继续用相对路径，无需改 `baseURL`。

## 构建

在仓库根目录：

```bash
# Git Bash / Linux / macOS — 当前平台
sh scripts/build_prod.sh

# 交叉编译 Linux（默认 amd64；可选 arm64）
sh scripts/build_prod.sh linux
sh scripts/build_prod.sh linux arm64
```

```powershell
# PowerShell — 当前平台
.\scripts\build_prod.ps1

# 交叉编译 Linux（默认 amd64；可选 arm64）
.\scripts\build_prod.ps1 -Target linux
.\scripts\build_prod.ps1 -Target linux -Arch arm64
```

产出命名：`bin/mall-<os>-<arch>`（Windows 带 `.exe`）。  
当前平台额外会复制一份短名 `bin/mall` / `bin/mall.exe` 方便本地跑。

已用 `CGO_ENABLED=0`，可在 Windows/macOS 上直接交叉出 Linux 二进制（sqlite 走纯 Go 驱动）。

## 运行

建议在 `server/` 下启动（配置、SQLite、`data/images` 等路径相对 cwd）：

```bash
cd server
../bin/mall -conf configs/config_prod.yaml
# 可选：
# ../bin/mall -conf configs/config_prod.yaml -admin-addr :8081 -web-addr :8082
```

上线前请把 `configs/config_prod.yaml` 里的 `storage.public_base` 改成对外可访问的 **API** 地址（例如 `https://api.example.com`），否则图片绝对 URL 会指错主机。
