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
# Git Bash / Linux / macOS
sh scripts/build_prod.sh
```

```powershell
# PowerShell
.\scripts\build_prod.ps1
```

产出：`bin/mall`（Windows 为 `bin/mall.exe`）。

## 运行

建议在 `server/` 下启动（配置、SQLite、`data/images` 等路径相对 cwd）：

```bash
cd server
../bin/mall -conf configs/config_prod.yaml
# 可选：
# ../bin/mall -conf configs/config_prod.yaml -admin-addr :8081 -web-addr :8082
```

上线前请把 `configs/config_prod.yaml` 里的 `storage.public_base` 改成对外可访问的 **API** 地址（例如 `https://api.example.com`），否则图片绝对 URL 会指错主机。
