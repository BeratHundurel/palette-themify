param(
    [ValidateSet('start', 'stop', 'restart', 'status', 'logs')]
    [string]$Action = 'start'
)

$ErrorActionPreference = 'Stop'
$repositoryRoot = Split-Path -Parent $PSScriptRoot
$composeArguments = @(
    'compose',
    '--env-file', (Join-Path $repositoryRoot 'env.development'),
    '-f', (Join-Path $repositoryRoot 'docker-compose.yml'),
    '-f', (Join-Path $repositoryRoot 'docker-compose.observability.yml')
)

function Invoke-DevelopmentCompose {
    param([string[]]$ComposeCommand)

    & docker @composeArguments @ComposeCommand
    if ($LASTEXITCODE -ne 0) {
        throw "Docker Compose failed with exit code $LASTEXITCODE"
    }
}

function Start-DevelopmentStack {
    Invoke-DevelopmentCompose -ComposeCommand @('up', '--build', '--detach', '--wait', '--wait-timeout', '240')
    Write-Host ''
    Write-Host 'ThemeSmith is running:'
    Write-Host '  App:     http://localhost:5173'
    Write-Host '  Jaeger:  http://localhost:16686'
    Write-Host '  Go API:  http://localhost:8088/health'
    Write-Host '  Zig API: http://localhost:8089/health'
}

switch ($Action) {
    'start' {
        Start-DevelopmentStack
    }
    'stop' {
        Invoke-DevelopmentCompose -ComposeCommand @('down')
        Write-Host 'ThemeSmith development services stopped. Database data was preserved.'
    }
    'restart' {
        Invoke-DevelopmentCompose -ComposeCommand @('down')
        Start-DevelopmentStack
    }
    'status' {
        Invoke-DevelopmentCompose -ComposeCommand @('ps')
    }
    'logs' {
        Invoke-DevelopmentCompose -ComposeCommand @('logs', '--follow', '--tail', '100')
    }
}
