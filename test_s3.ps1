$ServerUrl = "http://localhost:8080"
$BearerToken = "(^&%sdfuyigsdfiuhgy(^&*"
$TestId = "s3-test-$(Get-Date -Format 'yyyyMMddHHmmss')"

$S3Endpoint = "https://fahad-rekaz.s3.ca-central-1.amazonaws.com"
$S3BucketName = "fahad-rekaz"
$S3Region = "ca-central-1"

Write-Host "=========================================="
Write-Host "AWS S3 STORAGE TEST"
Write-Host "=========================================="
Write-Host ""
Write-Host "Configuration:"
Write-Host "  Endpoint: $S3Endpoint"
Write-Host "  Bucket: $S3BucketName"
Write-Host "  Region: $S3Region"
Write-Host ""

Write-Host "Cleaning up old processes..."
Get-Process -Name go -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Get-NetTCPConnection -LocalPort 8080 -ErrorAction SilentlyContinue | ForEach-Object { 
    Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue 
}
Start-Sleep -Seconds 2

Write-Host "[1/5] Starting server with AWS S3 backend..."

$S3AccessKey = "YOUR_AWS_ACCESS_KEY_ID"
$S3SecretKey = "YOUR_AWS_SECRET_ACCESS_KEY"

Write-Host "Opening server in new terminal window..."
$serverProcess = Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd '$PWD'; `$env:STORAGE_BACKEND='s3'; `$env:BEARER_TOKEN='$BearerToken'; `$env:S3_ENDPOINT='$S3Endpoint'; `$env:S3_ACCESS_KEY='$S3AccessKey'; `$env:S3_SECRET_KEY='$S3SecretKey'; `$env:S3_BUCKET_NAME='$S3BucketName'; `$env:S3_REGION='$S3Region'; `$env:DATABASE_URL='./metadata.db'; `$env:SERVER_PORT='8080'; Write-Host 'AWS S3 SERVER' -ForegroundColor Green; go run main.go" -PassThru

Write-Host "Waiting for server to start..."
Start-Sleep -Seconds 6

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

Write-Host "[2/5] Uploading blob to AWS S3..."
$testData = "Hello from S3 Storage - Test at $(Get-Date)"
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
    
    Write-Host "SUCCESS: Blob uploaded to AWS S3"
    Write-Host "  ID: $($uploadResult.id)"
    Write-Host ""
} catch {
    Write-Host "ERROR: Failed to upload blob"
    Write-Host "  $($_.Exception.Message)"
    exit 1
}

Write-Host "[3/5] Retrieving blob from AWS S3..."
try {
    $retrieveResult = Invoke-RestMethod `
        -Uri "$ServerUrl/v1/blobs/$TestId" `
        -Method Get `
        -Headers @{ "Authorization" = "Bearer $BearerToken" }
    
    Write-Host "SUCCESS: Blob retrieved from AWS S3"
    Write-Host "  ID: $($retrieveResult.id)"
    Write-Host "  Size: $($retrieveResult.size) bytes"
    Write-Host "  Created: $($retrieveResult.created_at)"
    Write-Host ""
} catch {
    Write-Host "ERROR: Failed to retrieve blob"
    Write-Host "  $($_.Exception.Message)"
    exit 1
}

Write-Host "[4/5] Verifying data..."
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
Write-Host "AWS S3 STORAGE TEST PASSED"
Write-Host "=========================================="
Write-Host ""
Write-Host "AWS S3 Details:"
Write-Host "  Region: $S3Region"
Write-Host "  Bucket: $S3BucketName"
Write-Host "  Object ID: $TestId"
Write-Host ""
Write-Host "View in AWS Console:"
Write-Host "  https://s3.console.aws.amazon.com/s3/buckets/$S3BucketName"
Write-Host ""
Write-Host "[5/5] Stopping server..."
Stop-Process -Id $serverProcess.Id -Force -ErrorAction SilentlyContinue
Write-Host "Server stopped"

