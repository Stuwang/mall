# 构建 admin / web_app，拷贝 dist 到 allinone embed 目录，再编译单进程二进制。
#
# 用法:
#   .\scripts\build_prod.ps1
#   .\scripts\build_prod.ps1 -Target linux
#   .\scripts\build_prod.ps1 -Target linux -Arch arm64
#   .\scripts\build_prod.ps1 -Target windows -Arch amd64
param(
    [string]$Target = "",
    [string]$Arch = ""
)

$ErrorActionPreference = "Stop"

$Root = Resolve-Path (Join-Path $PSScriptRoot "..")
$AdminDist = Join-Path $Root "admin\dist"
$WebDist = Join-Path $Root "web_app\dist"
$EmbedAdmin = Join-Path $Root "server\cmd\allinone\static\admin"
$EmbedWeb = Join-Path $Root "server\cmd\allinone\static\web_app"
$OutDir = Join-Path $Root "bin"

$hostOS = (go env GOOS).Trim()
$hostArch = (go env GOARCH).Trim()
if ([string]::IsNullOrWhiteSpace($Target)) {
    $Target = $hostOS
}
if ([string]::IsNullOrWhiteSpace($Arch)) {
    if ($Target -eq $hostOS) {
        $Arch = $hostArch
    } else {
        $Arch = "amd64"
    }
}

$OutName = "mall-$Target-$Arch"
if ($Target -eq "windows") {
    $OutName = "$OutName.exe"
}

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

Write-Host "==> go build allinone (GOOS=$Target GOARCH=$Arch)"
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null
$OutPath = Join-Path $OutDir $OutName
Push-Location (Join-Path $Root "server")
try {
    $env:CGO_ENABLED = "0"
    $env:GOOS = $Target
    $env:GOARCH = $Arch
    go build -o $OutPath .\cmd\allinone
} finally {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Pop-Location
}

if ($Target -eq $hostOS -and $Arch -eq $hostArch) {
    $ShortName = if ($Target -eq "windows") { "mall.exe" } else { "mall" }
    $ShortPath = Join-Path $OutDir $ShortName
    Copy-Item -Force $OutPath $ShortPath
    Write-Host "==> also: $ShortPath"
}

Write-Host "==> done: $OutPath"
if ($Target -eq "linux") {
    Write-Host "Deploy example:"
    Write-Host "  scp bin\$OutName user@host:/opt/mall/"
    Write-Host "  cd /opt/mall && ./$OutName -conf configs/config_prod.yaml"
} else {
    Write-Host "Run from server/ (or adjust config paths), e.g.:"
    Write-Host "  cd `"$Root\server`"; ..\bin\$OutName -conf configs\config_prod.yaml"
}
