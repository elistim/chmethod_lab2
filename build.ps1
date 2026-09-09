$ErrorActionPreference = 'Stop'
Push-Location $PSScriptRoot
try {
    New-Item -ItemType Directory -Force dist | Out-Null
    go build -trimpath -ldflags '-s -w' -o dist/chmethod_lab2.exe ./cmd/lab
    if ($LASTEXITCODE -ne 0) { throw 'Go build failed' }
    Write-Host 'Built: dist/chmethod_lab2.exe'
} finally {
    Pop-Location
}
