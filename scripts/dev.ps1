# Nexbic Platform - Development Quick Start
param(
    [switch]$InitDB,
    [switch]$Start,
    [switch]$Stop,
    [switch]$Reset
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot

function Write-Step {
    param([string]$Message)
    Write-Host "==> $Message" -ForegroundColor Cyan
}

function Write-OK {
    param([string]$Message)
    Write-Host "  [OK] $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "  [ERR] $Message" -ForegroundColor Red
}

switch ($true) {
    ($InitDB -or $Reset) {
        Write-Step "Initializing database..."

        $envVars = @{}
        Get-Content "$ProjectRoot\.env" | ForEach-Object {
            if ($_ -match '^([^#=]+)=(.+)$') {
                $envVars[$matches[1]] = $matches[2].Trim('"', "'")
            }
        }

        $DB_USER = $envVars['DB_USER']
        $DB_PASSWORD = $envVars['DB_PASSWORD']
        $DB_NAME = $envVars['DB_NAME']
        $DB_HOST = $envVars['DB_HOST']

        $env:PGPASSWORD = $DB_PASSWORD

        $initFiles = @(
            "001-schema.sql",
            "002-functions.sql",
            "003-roles.sql",
            "004-platform.sql"
        )

        foreach ($file in $initFiles) {
            $path = "$ProjectRoot\postgres\init\$file"
            if (Test-Path $path) {
                Write-Host "  Running $file..."
                psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f $path -q
                if ($LASTEXITCODE -eq 0) {
                    Write-OK "$file applied"
                } else {
                    Write-Error "$file failed"
                }
            }
        }

        Write-OK "Database initialized"
        break
    }

    ($Start) {
        Write-Step "Starting Nexbic Platform..."

        Write-Host "1. Starting Docker services (PostgreSQL, Redis, PostgREST)..."
        docker compose -f "$ProjectRoot\docker-compose.yml" up -d postgres redis postgrest pgbouncer
        Write-OK "Docker services started"

        Write-Host "2. Initializing database..."
        & $PSScriptRoot/dev.ps1 -InitDB

        Write-Host "3. Starting Go services..."
        $jobs = @()

        $jobs += Start-Job -Name "gateway" -ScriptBlock {
            Set-Location $using:ProjectRoot
            go run ./gateway
        }

        $jobs += Start-Job -Name "management-api" -ScriptBlock {
            Set-Location $using:ProjectRoot
            go run ./management-api
        }

        Write-Host "4. Starting dashboard..."
        $jobs += Start-Job -Name "dashboard" -ScriptBlock {
            Set-Location "$using:ProjectRoot\dashboard"
            npm run dev
        }

        Write-Host ""
        Write-Host "=== Nexbic Platform Running ===" -ForegroundColor Green
        Write-Host "Gateway:        http://localhost:8080"
        Write-Host "Management API: http://localhost:8081"
        Write-Host "Dashboard:      http://localhost:5173"
        Write-Host "PostgREST:      http://localhost:3000"
        Write-Host "PostgreSQL:     localhost:5432"
        Write-Host "Redis:          localhost:6379"
        Write-Host ""
        Write-Host "Press Ctrl+C to stop all services..."

        try {
            while ($true) {
                Start-Sleep -Seconds 1
                $running = $jobs | Where-Object { $_.State -eq 'Running' }
                if ($running.Count -eq 0) {
                    Write-Error "All services stopped"
                    break
                }
            }
        }
        finally {
            Write-Host "Stopping all services..."
            $jobs | Stop-Job -PassThru | Remove-Job
            docker compose -f "$ProjectRoot\docker-compose.yml" down
        }
        break
    }

    ($Stop) {
        Write-Step "Stopping all services..."
        Get-Job -Name "gateway", "management-api", "dashboard" -ErrorAction SilentlyContinue |
            Stop-Job -PassThru | Remove-Job
        docker compose -f "$ProjectRoot\docker-compose.yml" down
        Write-OK "All services stopped"
        break
    }

    default {
        Write-Host @"
Nexbic Platform - Development Script

Usage:
  .\scripts\dev.ps1 -InitDB     Initialize database schema
  .\scripts\dev.ps1 -Start      Start all services
  .\scripts\dev.ps1 -Stop       Stop all services
  .\scripts\dev.ps1 -Reset      Reset database and restart

Prerequisites:
  - Go 1.23+
  - Docker Desktop
  - PostgreSQL client (psql)
  - Node.js 22+

"@
    }
}
