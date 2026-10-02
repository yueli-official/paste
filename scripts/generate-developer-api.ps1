param(
    [Parameter(Mandatory = $true)][string]$GeneratorPath,
    [switch]$Check
)

$ErrorActionPreference = 'Stop'
$generator = (Resolve-Path -LiteralPath $GeneratorPath).Path
$root = Split-Path $PSScriptRoot -Parent
$temporary = Join-Path ([IO.Path]::GetTempPath()) ('paste-developer-api-' + [guid]::NewGuid().ToString('N'))
$null = New-Item -ItemType Directory -Path $temporary
$oldOutput = $env:PASTE_OPENAPI_OUTPUT
try {
    Push-Location $root
    try {
        & go run ./cmd/developer-api --output (Join-Path $temporary 'personal-api.json')
        if ($LASTEXITCODE -ne 0) { throw 'PAT declaration export failed' }
        $env:PASTE_OPENAPI_OUTPUT = Join-Path $temporary 'paste.json'
        & go run ./cmd/paste
        if ($LASTEXITCODE -ne 0) { throw 'OpenAPI export failed' }
    } finally {
        Pop-Location
        $env:PASTE_OPENAPI_OUTPUT = $oldOutput
    }
    $arguments = @($generator, '--declaration', (Join-Path $temporary 'personal-api.json'), '--openapi', (Join-Path $temporary 'paste.json'), '--output-dir', (Join-Path $root 'docs/developer-api'))
    if ($Check) { $arguments += '--check' }
    & python @arguments
    if ($LASTEXITCODE -ne 0) { throw 'Developer API generation/check failed' }
    $contracts = @{'personal-api.json' = 'contracts/personal-api.json'; 'paste.json' = 'contracts/openapi/paste.json'}
    foreach ($name in $contracts.Keys) {
        $source = Join-Path $temporary $name
        $target = Join-Path $root $contracts[$name]
        if ($Check) {
            if (!(Test-Path -LiteralPath $target) -or (Get-FileHash -LiteralPath $source).Hash -ne (Get-FileHash -LiteralPath $target).Hash) { throw "Contract differs: $($contracts[$name])" }
        } else {
            Copy-Item -LiteralPath $source -Destination $target
        }
    }
} finally {
    $resolved = [IO.Path]::GetFullPath($temporary)
    $tempRoot = [IO.Path]::GetFullPath([IO.Path]::GetTempPath()).TrimEnd('\', '/') + [IO.Path]::DirectorySeparatorChar
    if (!$resolved.StartsWith($tempRoot, [StringComparison]::OrdinalIgnoreCase)) { throw 'Unexpected temporary directory' }
    Remove-Item -LiteralPath $resolved -Recurse -Force
}
