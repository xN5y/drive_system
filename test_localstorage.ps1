$ServerUrl = "http://localhost:8080"
$BearerToken = "(^&%sdfuyigsdfiuhgy(^&*"
$TestId = "local-test-$(Get-Date -Format 'yyyyMMddHHmmss')"

Write-Host "=========================================="
Write-Host "LOCAL STORAGE TEST"
Write-Host "=========================================="
Write-Host ""

Write-Host "Cleaning up old processes..."
Get-Process -Name go -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue | ForEach-Object { 
    Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue 
}
Start-Sleep -Seconds 2

Write-Host "[0/4] Starting server with LOCAL storage backend..."
Write-Host "Opening server in new terminal window..."
$serverProcess = Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD'; `$env:STORAGE_BACKEND='local'; `$env:BEARER_TOKEN='$BearerToken'; `$env:LOCAL_STORAGE_PATH='./storage_data'; `$env:DATABASE_URL='./metadata.db'; `$env:SERVER_PORT='8080'; Write-Host 'LOCAL STORAGE SERVER' -ForegroundColor Green; go run main.go" -PassThru

Write-Host "Waiting for server to start..."
Start-Sleep -Seconds 5

try {
    $testConnection = Invoke-WebRequest -Uri "$ServerUrl/v1/blobs/test" -Method Get -Headers @{ "Authorization" = "Bearer $BearerToken" } -ErrorAction SilentlyContinue
} catch {
    if ($_.Exception.Response.StatusCode -eq 404 -or $_.Exception.Response.StatusCode -eq 401) {
        Write-Host "SUCCESS: Server started (PID: $($serverProcess.Id))"
    } else {
        Write-Host "ERROR: Server may not be running properly"
        Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
        exit 1
    }
}
Write-Host ""

Write-Host "[1/4] Uploading blob..."
$testData = "Hello from Local Storage - Test at $(Get-Date)"
$base64Data = [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($testData))
$requestBody = @{
    id = $TestId
    data = $base64Data
} | ConvertTo-Json

try {
    $uploadResult = Invoke-RestMethod `
        -Uri "$ServerUrl/v1/blobs" `
        -Method Post `
        -Headers @{ "Authorization" = "Bearer $BearerToken" } `
        -Body $requestBody `
        -ContentType "application/json"
    
    Write-Host "SUCCESS: Blob uploaded"
    Write-Host "  ID: $($uploadResult.id)"
    Write-Host ""
} catch {
    Write-Host "ERROR: Failed to upload blob"
    Write-Host "  $($_.Exception.Message)"
    exit 1
}

Write-Host "[2/4] Retrieving blob..."
try {
    $retrieveResult = Invoke-RestMethod `
        -Uri "$ServerUrl/v1/blobs/$TestId" `
        -Method Get `
        -Headers @{ "Authorization" = "Bearer $BearerToken" }
    
    Write-Host "SUCCESS: Blob retrieved"
    Write-Host "  ID: $($retrieveResult.id)"
    Write-Host "  Size: $($retrieveResult.size) bytes"
    Write-Host "  Created: $($retrieveResult.created_at)"
    Write-Host ""
} catch {
    Write-Host "ERROR: Failed to retrieve blob"
    Write-Host "  $($_.Exception.Message)"
    exit 1
}

Write-Host "[3/4] Verifying data..."
$decodedData = [System.Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($retrieveResult.data))

if ($decodedData -eq $testData) {
    Write-Host "SUCCESS: Data matches!"
    Write-Host "  Original: $testData"
    Write-Host "  Retrieved: $decodedData"
} else {
    Write-Host "ERROR: Data mismatch!"
    Write-Host "  Expected: $testData"
    Write-Host "  Got: $decodedData"
    exit 1
}

Write-Host ""
Write-Host "=========================================="
Write-Host "LOCAL STORAGE TEST PASSED"
Write-Host "=========================================="
Write-Host ""
Write-Host "Files saved in: ./storage_data/"
Write-Host "Metadata saved in: ./metadata.db"
Write-Host ""

if (Test-Path "./storage_data/$TestId") {
    Write-Host "File exists on disk: ./storage_data/$TestId"
    $fileSize = (Get-Item "./storage_data/$TestId").Length
    Write-Host "File size: $fileSize bytes"
}

Write-Host ""
Write-Host "[4/4] Stopping server..."
Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
Write-Host "Server stopped"

