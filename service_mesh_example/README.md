# Consul service mesh 示例

在本机 Kiri（k3s）上跑 Consul service mesh。Kratos HTTP / gRPC 都由 proto 生成。

| 服务 | 协议 | 集群外 |
| --- | --- | --- |
| biz-a / biz-b / biz-c | HTTP `:8000` | Windows `localhost:18081` / `18082` / `18083`，permissive mTLS |
| svc-a / svc-b / svc-c | gRPC `:9000` | 不暴露。strict mTLS，只有网格内的 biz 和 svc 能调 |

biz 和 svc 都可以按名字调用任意 svc。`depth=1` 时 svc 只返回自己；`depth=2` 时 svc 再去调其他 svc。

## 目录

- `api/` proto 和生成的 pb
- `cmd/biz`、`cmd/svc` 两个进程，用 `SERVICE_NAME` 区分六个部署
- `deploy/Dockerfile` 镜像
- `deploy/consul/values.yaml` Consul Helm 值
- `deploy/k8s/` 业务清单
- `deploy/deploy.ps1` 构建、导入镜像、安装 Consul、发布
- `deploy/generate.ps1` 重新生成 pb

Kiri 的 Docker 和 k3s 的 containerd 不是同一套存储。脚本把镜像 `docker save` 后 `ctr -n k8s.io images import` 进 `kiri-k8s`，清单里写 `imagePullPolicy: Never`（Consul 官方镜像是 `IfNotPresent`，同样先导入）。

## 发布

kubectl 当前 context 需要是 `kiri`。脚本会关闭 kubectl 的 aggregated discovery，否则 k3s 列不全 Consul 2.1 的 CRD。

Consul ACL 在这份 values 里是关的：k3s 的 ServiceAccount 登录会被 auth method 拒绝，sidecar 初始化起不来。透明代理和 mTLS 仍然生效。svc 是 strict mTLS，网格外连不进去。biz 要 permissive，所以 `Mesh` 里打开了 `allowEnablingPermissiveMutualTLS`。改完 Consul 之后需要 `kubectl -n mesh rollout restart deployment`。

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

网格外客户端是 `mesh-client`（没有 sidecar）：

```powershell
kubectl -n mesh exec deploy/mesh-client -- wget -qO- "http://biz-a.mesh.svc.cluster.local:18081/api/v1/ping"
kubectl -n mesh exec deploy/mesh-client -- wget -qO- "http://biz-a.mesh.svc.cluster.local:18081/api/v1/call?depth=2&targets=svc-a,svc-b,svc-c"
```

Kiri 跑在 WSL 里，k3s 又在 `kiri-k8s` 容器里。NodePort 和 ClusterIP 到不了 Windows，只有 LoadBalancer 会被转到 `127.0.0.1`。svc 仍是 ClusterIP。三个 biz 因此是：

```text
http://127.0.0.1:18081/api/v1/ping
http://127.0.0.1:18082/api/v1/ping
http://127.0.0.1:18083/api/v1/ping
```

`depth` 最大 3。不写 `targets` 时默认三个 svc。
