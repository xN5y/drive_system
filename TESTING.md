# Testing Guide

This document explains how to test each storage backend in Simple Drive.

## Overview

Three independent PowerShell test scripts are provided:
- `test_localstorage.ps1` - Tests local filesystem storage
- `test_s3.ps1` - Tests AWS S3 storage
- `test_database.ps1` - Tests SQLite database storage

Each script:
- Cleans up any existing processes on port 8080
- Starts the server in a new visible terminal window
- Runs tests against the API
- Stops the server automatically

---

## Test 1: Local Storage

### Prerequisites
None. Works out of the box.

### Running the Test

```powershell
.\test_localstorage.ps1
```

### What It Does
1. Cleans up old processes and port 8080
2. Starts server with local storage backend in new terminal
3. Uploads a test blob to `./storage_data/`
4. Retrieves the blob and verifies data integrity
5. Stops the server

### What Gets Created
- `./storage_data/` - Directory containing blob files
- `./metadata.db` - SQLite database with metadata

---

## Test 2: AWS S3 Storage

### Prerequisites

You need an AWS account with:
1. An S3 bucket (must already exist)
2. IAM credentials with these permissions:
   - `AdministratorAccess`
   - `AmazonS3FullAccess`

Edit `test_s3.ps1` and update lines 5-7:

$S3Endpoint = "https://YOUR-BUCKET-NAME.s3.YOUR-REGION.amazonaws.com"
$S3BucketName = "YOUR-BUCKET-NAME"
$S3Region = "YOUR-REGION"


Then update lines 29-30 with your AWS credentials:


$env:S3_ACCESS_KEY = "YOUR_AWS_ACCESS_KEY_ID"
$env:S3_SECRET_KEY = "YOUR_AWS_SECRET_ACCESS_KEY"


### What It Does
1. Cleans up old processes and port 8080
2. Starts server with S3 storage backend in new terminal
3. Uploads a test blob to AWS S3
4. Retrieves the blob and verifies data integrity
5. Stops the server

## Test 3: Database Storage

### Prerequisites
None. SQLite is built into Go.

### Running the Test

```powershell
.\test_database.ps1
```

### What It Does
1. Cleans up old processes and port 8080
2. Starts server with database storage backend in new terminal
3. Uploads a test blob to SQLite database
4. Retrieves the blob and verifies data integrity
5. Stops the server

### What Gets Created
- `./storage.db` - SQLite database with blob data
- `./metadata.db` - SQLite database with metadata

---

## How the Scripts Work

1. **Port Cleanup**:
   - Kills any Go processes
   - Kills any process using port 8080
   - Waits 2 seconds

2. **Server Launch**:
   - Opens a new visible PowerShell window
   - Sets environment variables in that window
   - Runs `go run main.go`
   - Window stays open for debugging

3. **Testing**:
   - Waits 5-6 seconds for server to start
   - Makes HTTP requests to test endpoints
   - Verifies data integrity

4. **Cleanup**:
   - Stops the server process
   - Closes the terminal window

---

## Environment Variables

Each test script sets these internally (no need to set them manually):

### Local Storage
- `STORAGE_BACKEND=local`
- `LOCAL_STORAGE_PATH=./storage_data`
- `BEARER_TOKEN=(^&%sdfuyigsdfiuhgy(^&*`
- `SERVER_PORT=8080`

### S3 Storage
- `STORAGE_BACKEND=s3`
- `S3_ENDPOINT=https://bucket.s3.region.amazonaws.com`
- `S3_ACCESS_KEY=your-key`
- `S3_SECRET_KEY=your-secret`
- `S3_BUCKET_NAME=your-bucket`
- `S3_REGION=your-region`
- `BEARER_TOKEN=(^&%sdfuyigsdfiuhgy(^&*`
- `SERVER_PORT=8080`

### Database Storage
- `STORAGE_BACKEND=database`
- `DATABASE_URL=./storage.db`
- `BEARER_TOKEN=(^&%sdfuyigsdfiuhgy(^&*`
- `SERVER_PORT=8080`

---

