# Mall Web App (H5)

商城用户端 H5（Vue 3 + TypeScript + Vant）。

## 技术栈

- Vue 3 + Vite + TypeScript
- Vant 4 / Vue Router / Pinia / Axios
- OOP：`AuthSession`、`HttpClient`、各 `*ApiService`、`ApiHub`

## 开发

```bash
npm install
npm run dev
```

默认端口 `5174`，代理到 `http://localhost:8080`。

联调前请确认 server `configs/config.yaml` 中 `env` **不要**为 `test`（test 走 `UserId` Header，非 JWT）。

## 功能

- 注册 / 登录 / 注销（本地清 token）
- 商品浏览、分类、详情
- 购物车、结算（Mock 支付）
- 我的：积分、订单列表

## 脚本

- `npm run dev` 开发
- `npm run build` 类型检查 + 构建
- `npm run preview` 预览构建结果
