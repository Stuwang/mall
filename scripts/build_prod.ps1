# 构建 admin / web_app，拷贝 dist 到 allinone embed 目录，再编译单进程二进制。
$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
$AdminDist = Join-Path $Root "admin\dist"
$WebDist = Join-Path $Root "web_app\dist"
$EmbedAdmin = Join-Path $Root "server\cmd\allinone\static\admin"
$EmbedWeb = Join-Path $Root "server\cmd\allinone\static\web_app"
$OutDir = Join-Path $Root "bin"
$OutName = "mall.exe"

function Invoke-NpmBuild([string]$Dir) {
    Push-Location $Dir
    try {
        if (Test-Path "package-lock.json") {
            npm ci
        } else {
            npm install
        }
        npm run build
    } finally {
        Pop-Location
    }
}

Write-Host "==> build admin"
Invoke-NpmBuild (Join-Path $Root "admin")

Write-Host "==> build web_app"
Invoke-NpmBuild (Join-Path $Root "web_app")

Write-Host "==> copy dist into embed dirs"
if (Test-Path $EmbedAdmin) { Remove-Item -Recurse -Force $EmbedAdmin }
if (Test-Path $EmbedWeb) { Remove-Item -Recurse -Force $EmbedWeb }
New-Item -ItemType Directory -Force -Path $EmbedAdmin | Out-Null
New-Item -ItemType Directory -Force -Path $EmbedWeb | Out-Null
Copy-Item -Path (Join-Path $AdminDist "*") -Destination $EmbedAdmin -Recurse -Force
Copy-Item -Path (Join-Path $WebDist "*") -Destination $EmbedWeb -Recurse -Force
New-Item -ItemType File -Force -Path (Join-Path $EmbedAdmin ".gitkeep") | Out-Null
New-Item -ItemType File -Force -Path (Join-Path $EmbedWeb ".gitkeep") | Out-Null

Write-Host "==> go build allinone"
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
Push-Location (Join-Path $Root "server")
try {
    $env:CGO_ENABLED = "0"
    go build -o (Join-Path $OutDir $OutName) .\cmd\allinone
} finally {
    Pop-Location
}

$OutPath = Join-Path $OutDir $OutName
Write-Host "==> done: $OutPath"
Write-Host "Run from server/ (or adjust config paths), e.g.:"
Write-Host "  cd `"$Root\server`"; ..\bin\$OutName -conf configs\config_prod.yaml"
