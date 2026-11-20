# Regenerate pb.go from proto. Uses the repo bin/protoc and plugins.
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$Repo = Split-Path -Parent $Root
$Bin = Join-Path $Repo "bin"

function Assert-Ok([string]$What) {
    if ($LASTEXITCODE -ne 0) { throw "$What failed: exit $LASTEXITCODE" }
}

Push-Location $Root
try {
    & (Join-Path $Bin "protoc.exe") `
        -I api -I (Join-Path $Repo "server\third_party") `
        --plugin=("protoc-gen-go=" + (Join-Path $Bin "protoc-gen-go.exe")) `
        --plugin=("protoc-gen-go-grpc=" + (Join-Path $Bin "protoc-gen-go-grpc.exe")) `
        --plugin=("protoc-gen-go-http=" + (Join-Path $Bin "protoc-gen-go-http.exe")) `
        --go_out=. --go_opt=module=mall/service_mesh_example `
        --go-grpc_out=. --go-grpc_opt=module=mall/service_mesh_example `
        --go-http_out=. --go-http_opt=module=mall/service_mesh_example `
        api/biz/v1/biz.proto api/svc/v1/svc.proto
    Assert-Ok "protoc"
} finally {
    Pop-Location
}
