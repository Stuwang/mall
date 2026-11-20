# Consul service mesh 示例

在本机 Kiri（k3s）上跑 Consul service mesh。Kratos HTTP / gRPC 都由 proto 生成。透明代理是关的，Pod 里没有 nftables 劫持。入口只有 API Gateway 的 Envoy，按路径转发。

| 方向 | 地址 | 路由 |
| --- | --- | --- |
| 南北向，HTTP | Service `mesh-ingress`，`127.0.0.1:8080` | `/biz_a`、`/biz_b`、`/biz_c` |
| 东西向，gRPC | Service `mesh-eastwest`，ClusterIP `:9090` | `/mesh.svc.v1.SvcA`、`SvcB`、`SvcC` |

biz 和 svc 的业务端口只听 `127.0.0.1`。调用方不再写每个 svc 的地址，gRPC 只拨网关这一个端口，靠 proto service 的方法路径区分。

`depth=1` 时 svc 只返回自己；`depth=2` 时 svc 再去调其他 svc。

## 目录

- `api/` proto 和生成的 pb
- `cmd/biz`、`cmd/svc` 两个进程，用 `SERVICE_NAME` 区分六个部署
- `deploy/Dockerfile` 镜像
- `deploy/consul/values.yaml` Consul Helm 值
- `deploy/k8s/` 业务清单和 Gateway
- `deploy/deploy.ps1` 构建、导入镜像、安装 Consul、发布
- `deploy/generate.ps1` 重新生成 pb

Kiri 的 Docker 和 k3s 的 containerd 不是同一套存储。脚本把镜像 `docker save` 后 `ctr -n k8s.io images import` 进 `kiri-k8s`，清单里写 `imagePullPolicy: Never`（Consul 官方镜像是 `IfNotPresent`，同样先导入）。

## 发布

kubectl 当前 context 需要是 `kiri`。脚本会关闭 kubectl 的 aggregated discovery，否则 k3s 列不全 Consul 2.1 的 CRD。

Consul ACL 在这份 values 里是关的：k3s 的 ServiceAccount 登录会被 auth method 拒绝，sidecar 初始化起不来。网关到 sidecar 仍然是 mTLS。改完 Consul 之后需要 `kubectl -n mesh rollout restart deployment`。

```powershell
cd service_mesh_example
.\deploy\deploy.ps1
```

Consul UI：

```powershell
kubectl -n consul port-forward svc/consul-ui 8501:443
```

ACL 开启后，UI token 在 secret `consul-bootstrap-acl-token`。

## 调用

网格外客户端是 `mesh-client`（没有 sidecar）。三个 biz 共用网关的 HTTP 端口：

```powershell
kubectl -n mesh exec deploy/mesh-client -- wget -qO- "http://mesh-ingress.mesh.svc.cluster.local:8080/biz_a/ping"
kubectl -n mesh exec deploy/mesh-client -- wget -qO- "http://mesh-ingress.mesh.svc.cluster.local:8080/biz_a/call?depth=2&targets=svc-a,svc-b,svc-c"
```

Kiri 只把 LoadBalancer 转到 Windows。`mesh-ingress` 是 LoadBalancer，`mesh-eastwest` 是 ClusterIP，不会出现在本机端口上。本机访问：

```text
http://127.0.0.1:8080/biz_a/ping
http://127.0.0.1:8080/biz_b/ping
http://127.0.0.1:8080/biz_c/ping
```

服务之间的 gRPC 拨 `mesh-eastwest.mesh.svc.cluster.local:9090`，不经过 `mesh-ingress`。`depth` 最大 3。不写 `targets` 时默认三个 svc。
