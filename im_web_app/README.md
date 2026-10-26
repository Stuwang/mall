# im_web_app

IM 前端，技术栈与 `web_app` 相同：Vue 3 + Vite + TypeScript + Pinia + Vue Router + Vant + Axios。

对接独立 IM 服务（默认 `http://localhost:8090`）。

## 功能

- 注册 / 登录（IM 进程上的 user 模块）
- 会话列表（本地缓存 + `conv/list` 增量同步）
- 在线用户列表（`user/online`），无好友，点选即可开聊
- 聊天页发送文本；WebSocket `msg.push` 先展示，再用 history 兜底同步

## 开发

```bash
# 终端 1：IM 服务（仓库根目录）
go run build.go build im
./bin/im.exe -conf im/configs/config.yaml

# 终端 2：前端
cd im_web_app
npm install
npm run dev
```

浏览器打开 `http://localhost:5175`。用两个浏览器（或无痕窗口）各注册一个账号，双方都保持页面打开（会连 WebSocket），即可在「在线」里看到对方并聊天。

Vite 已代理：

| 前缀 | 目标 |
|------|------|
| `/user/api` | `localhost:8090` |
| `/im/api` | `localhost:8090` |
| `/im/ws` | `ws://localhost:8090` |

IM 配置需使用 JWT（`env` 不要用 `test`），见 `im/configs/config.yaml`。
