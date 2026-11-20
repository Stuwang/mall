# Build mesh-biz / mesh-svc, import them into Kiri's k3s, install Consul, and apply the mesh.
# Kiri docker and k3s containerd are separate. imagePullPolicy=Never only works after ctr import.
param(
    [switch]$SkipBuild,
    [switch]$SkipConsul,
    [switch]$SkipVerify
)

$ErrorActionPreference = "Stop"
# Consul 2.1 CRDs are too large for k3s aggregated discovery, so kubectl cannot see the kinds.
$env:KUBECTL_DISABLE_AGGREGATED_DISCOVERY = "true"
$Root = Split-Path -Parent $PSScriptRoot
$Tools = Join-Path $Root ".tools"
$K3sContainer = "kiri-k8s"
$ChartVersion = "2.1.0"
$HelmVersion = "v3.18.4"

function Assert-Ok([string]$What) {
    if ($LASTEXITCODE -ne 0) { throw "$What failed: exit $LASTEXITCODE" }
}

function Ensure-Helm {
    New-Item -ItemType Directory -Force -Path $Tools | Out-Null
    $helm = Join-Path $Tools "helm.exe"
    if (Test-Path $helm) { return $helm }
    $zip = Join-Path $Tools "helm.zip"
    Write-Host "==> download helm $HelmVersion"
    curl.exe -fsSL --retry 3 --max-time 180 -o $zip "https://get.helm.sh/helm-$HelmVersion-windows-amd64.zip"
    Assert-Ok "download helm"
    Expand-Archive -Path $zip -DestinationPath (Join-Path $Tools "helm-zip") -Force
    Move-Item -Force (Join-Path $Tools "helm-zip\windows-amd64\helm.exe") $helm
    return $helm
}

function Ensure-Chart {
    $tgz = Join-Path $Tools "consul-$ChartVersion.tgz"
    if (Test-Path $tgz) { return $tgz }
    Write-Host "==> download consul chart $ChartVersion"
    curl.exe -fsSL --retry 3 --max-time 180 -o $tgz "https://helm.releases.hashicorp.com/consul-$ChartVersion.tgz"
    Assert-Ok "download consul chart"
    return $tgz
}

function Ensure-DockerImage([string]$Image) {
    docker image inspect $Image *> $null
    if ($LASTEXITCODE -eq 0) { return }
    Write-Host "==> docker pull $Image"
    docker pull $Image
    Assert-Ok "docker pull $Image"
}

function Import-K3sImage([string]$Image) {
    Write-Host "==> import $Image into $K3sContainer"
    $safe = ($Image -replace '[\\/:]', '_')
    $tar = Join-Path $Tools "$safe.tar"
    docker save $Image -o $tar
    Assert-Ok "docker save $Image"
    docker cp $tar "${K3sContainer}:/tmp/kiri-import.tar"
    Assert-Ok "docker cp $Image"
    docker exec $K3sContainer ctr -n k8s.io images import /tmp/kiri-import.tar
    Assert-Ok "ctr import $Image"
    docker exec $K3sContainer rm -f /tmp/kiri-import.tar | Out-Null
    Remove-Item $tar -Force
}

if ((kubectl config current-context) -ne "kiri") {
    throw "kubectl context is not kiri"
}

$consulImages = @(
    "hashicorp/consul:$ChartVersion",
    "hashicorp/consul-k8s-control-plane:$ChartVersion",
    "hashicorp/consul-dataplane:$ChartVersion"
)

if (-not $SkipConsul) {
    foreach ($image in $consulImages) {
        Ensure-DockerImage $image
        Import-K3sImage $image
    }
    $helm = Ensure-Helm
    $chart = Ensure-Chart
    Write-Host "==> helm upgrade consul"
    & $helm upgrade --install consul $chart `
        --namespace consul --create-namespace `
        --values (Join-Path $PSScriptRoot "consul\values.yaml") `
        --wait --timeout 15m
    Assert-Ok "helm upgrade consul"
}

if (-not $SkipBuild) {
    Write-Host "==> build mesh-biz:local"
    docker build -f (Join-Path $PSScriptRoot "Dockerfile") --build-arg BIN=biz -t mesh-biz:local $Root
    Assert-Ok "docker build biz"
    Write-Host "==> build mesh-svc:local"
    docker build -f (Join-Path $PSScriptRoot "Dockerfile") --build-arg BIN=svc -t mesh-svc:local $Root
    Assert-Ok "docker build svc"
    Ensure-DockerImage "alpine:3.21"
    Import-K3sImage "mesh-biz:local"
    Import-K3sImage "mesh-svc:local"
    Import-K3sImage "alpine:3.21"
}

# k3s ServiceLB image. The node cannot pull Docker Hub, and Kiri only forwards LoadBalancer ports.
Ensure-DockerImage "rancher/klipper-lb:v0.4.17"
Import-K3sImage "rancher/klipper-lb:v0.4.17"

Write-Host "==> kubectl apply"
kubectl apply -f (Join-Path $PSScriptRoot "k8s\namespace.yaml")
Assert-Ok "apply namespace"
kubectl apply -f (Join-Path $PSScriptRoot "k8s\configmap.yaml")
Assert-Ok "apply configmap"

kubectl apply -f (Join-Path $PSScriptRoot "k8s\mesh.yaml")
Assert-Ok "apply mesh config"
kubectl apply -f (Join-Path $PSScriptRoot "k8s\biz.yaml")
Assert-Ok "apply biz"
kubectl apply -f (Join-Path $PSScriptRoot "k8s\svc.yaml")
Assert-Ok "apply svc"
kubectl apply -f (Join-Path $PSScriptRoot "k8s\client.yaml")
Assert-Ok "apply client"

$deploys = @("biz-a", "biz-b", "biz-c", "svc-a", "svc-b", "svc-c", "mesh-client")
foreach ($name in $deploys) {
    kubectl -n mesh rollout status "deployment/$name" --timeout=360s
    Assert-Ok "rollout $name"
}

if ($SkipVerify) { return }

Write-Host "==> ping biz-a from outside the mesh"
$ping = kubectl -n mesh exec deploy/mesh-client -- wget -T 20 -qO- "http://biz-a.mesh.svc.cluster.local:18081/api/v1/ping"
Assert-Ok "ping biz-a"
Write-Host $ping
if ($ping -notmatch "biz-a") { throw "ping response missing biz-a: $ping" }

Write-Host "==> biz-a call svc-a,svc-b,svc-c depth=2"
$call = kubectl -n mesh exec deploy/mesh-client -- wget -T 30 -qO- "http://biz-a.mesh.svc.cluster.local:18081/api/v1/call?depth=2&targets=svc-a,svc-b,svc-c"
Assert-Ok "call biz-a"
Write-Host $call
foreach ($name in @("svc-a", "svc-b", "svc-c")) {
    if ($call -notmatch $name) { throw "call response missing $name" }
}

Write-Host "==> svc-a must reject a client that is not in the mesh"
kubectl -n mesh exec deploy/mesh-client -- wget -T 8 -S -O - "http://svc-a.mesh.svc.cluster.local:9000/"
if ($LASTEXITCODE -eq 0) {
    throw "svc-a accepted a non-mesh HTTP call"
}
Write-Host "svc-a rejected the outside call (exit $LASTEXITCODE)"
