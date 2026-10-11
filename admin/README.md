# Mall Admin

商城后台管理前端（Vue 3 + TypeScript + Element Plus）。

## 技术栈

- Vue 3 + Vite + TypeScript
- Element Plus / Vue Router / Pinia / Axios
- OOP 风格：`AuthSession`、`HttpClient`、各 `*ApiService`、`ApiHub`

## 开发

```bash
npm install
npm run dev
```

默认代理到 `http://localhost:8080`。

联调前请确认：

1. server `configs/config.yaml` 中 `env` **不要**为 `test`（test 走 `UserId` Header，非 JWT）
2. 数据库用户 `is_admin = 1`

## 脚本

- `npm run dev` 开发
- `npm run build` 类型检查 + 构建
- `npm run preview` 预览构建结果
