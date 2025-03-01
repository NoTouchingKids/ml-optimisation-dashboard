# Start-MLServices.ps1
# This script starts all the ML Optimizer services in separate Windows Terminal tabs

# Verify Windows Terminal is installed
if (-not (Get-Command wt -ErrorAction SilentlyContinue)) {
    Write-Error "Windows Terminal is not installed. Please install it from the Microsoft Store."
    exit 1
}

# Define the base paths for each service
$projectRoot = $PSScriptRoot
if (-not $projectRoot -or $projectRoot -eq "") {
    # If running directly, use the current directory
    $projectRoot = Get-Location
}

$backendPath = Join-Path $projectRoot "backend"
$frontendPath = Join-Path $projectRoot "frontend"
$pythonServicePath = Join-Path $projectRoot "python_service"

# Check if directories exist
if (-not (Test-Path $backendPath)) {
    Write-Error "Backend directory not found at: $backendPath"
    exit 1
}

if (-not (Test-Path $frontendPath)) {
    Write-Error "Frontend directory not found at: $frontendPath"
    exit 1
}

if (-not (Test-Path $pythonServicePath)) {
    Write-Error "Python service directory not found at: $pythonServicePath"
    exit 1
}

# # Start Docker services first (if needed)
# Write-Host "Starting Docker services..."
# $dockerComposePath = Join-Path $projectRoot "docker"
# if (Test-Path $dockerComposePath) {
#     wt -w 0 new-tab --title "Docker Services" -d $dockerComposePath PowerShell -NoExit -Command "docker-compose up"
# } else {
#     Write-Warning "Docker compose directory not found, skipping Docker services"
# }

# # Wait for Docker services to initialize
# Start-Sleep -Seconds 5

# Start Frontend Service
Write-Host "Starting Frontend service..."
wt -w 0 new-tab --title "Frontend Service" -d $frontendPath PowerShell -NoExit -Command "bun run dev"

# Start Python Service
Write-Host "Starting Python service..."
wt -w 0 new-tab --title "Python Service" -d $pythonServicePath PowerShell -NoExit -Command "& {
    # Activate virtual environment
    # if (Test-Path .\.venv\Scripts\activate.ps1) {
    #     .\.venv\Scripts\activate.ps1
    # } elseif (Test-Path .\env\Scripts\activate.ps1) {
    #     .\env\Scripts\activate.ps1
    # } else {
    #     Write-Warning 'Virtual environment not found. Running without activation.'
    # }
    .\.venv\Scripts\activate.ps1
    # Run the Python service
    py main.py
}"

Start-Sleep -Seconds 5

# Start Backend Service
Write-Host "Starting Backend service..."
wt -w 0 new-tab --title "Backend Service" -d $backendPath PowerShell -NoExit -Command "go run .\cmd\server\main.go"


Write-Host "All services started in separate tabs!"