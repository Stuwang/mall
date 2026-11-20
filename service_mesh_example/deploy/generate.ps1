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
    $env:PATH = $Bin + [IO.Path]::PathSeparator + $env:PATH
    & (Join-Path $Bin "protoc.exe") `
        -I api -I (Join-Path $Repo "server\third_party") `
        --go_out=. --go_opt=module=mall/service_mesh_example `
        --go-grpc_out=. --go-grpc_opt=module=mall/service_mesh_example `
        --go-http_out=. --go-http_opt=module=mall/service_mesh_example `
        api/biz/v1/biz.proto api/svc/v1/svc.proto
    Assert-Ok "protoc"
} finally {
    Pop-Location
}
