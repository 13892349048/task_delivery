# Onboarding Workflow API Test Script
# Test all onboarding workflow related API endpoints

$baseUrl = "http://localhost:8081/api/v1"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "=== Onboarding Workflow API Test ===" -ForegroundColor Green

# 1. User login to get JWT Token
Write-Host "`n1. User login..." -ForegroundColor Yellow
$loginData = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginData -Headers $headers
    $token = $loginResponse.data.access_token
    $headers["Authorization"] = "Bearer $token"
    Write-Host "Login successful, got token" -ForegroundColor Green
} catch {
    Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 2. Create pending employee
Write-Host "`n2. Create pending employee..." -ForegroundColor Yellow
$pendingEmployeeData = @{
    real_name = "Zhang San"
    email = "zhangsan@example.com"
    phone = "13800138001"
    expected_date = "2024-01-15"
    notes = "New employee onboarding, software engineer position"
} | ConvertTo-Json

try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/pending" -Method POST -Body $pendingEmployeeData -Headers $headers
    $employeeId = $createResponse.data.employee_id
    Write-Host "Create pending employee successful, Employee ID: $employeeId" -ForegroundColor Green
    Write-Host "Response: $($createResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
} catch {
    Write-Host "Create pending employee failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error details: $errorBody" -ForegroundColor Red
    }
}

# 3. Confirm onboarding (requires employee ID)
if ($employeeId) {
    Write-Host "`n3. Confirm onboarding..." -ForegroundColor Yellow
    $confirmData = @{
        employee_id = $employeeId
        department_id = 1
        position_id = 1
        hire_date = "2024-01-15"
        probation_end_date = "2024-04-15"
        notes = "Confirm onboarding, assigned to R&D department"
    } | ConvertTo-Json

    try {
        $confirmResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/confirm" -Method POST -Body $confirmData -Headers $headers
        Write-Host "Confirm onboarding successful" -ForegroundColor Green
        Write-Host "Response: $($confirmResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "Confirm onboarding failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error details: $errorBody" -ForegroundColor Red
        }
    }

    # 4. Complete probation period
    Write-Host "`n4. Complete probation period..." -ForegroundColor Yellow
    try {
        $probationResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/$employeeId/probation" -Method POST -Headers $headers
        Write-Host "Complete probation period successful" -ForegroundColor Green
        Write-Host "Response: $($probationResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "Complete probation period failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error details: $errorBody" -ForegroundColor Red
        }
    }

    # 5. Confirm employee (probation to permanent)
    Write-Host "`n5. Confirm employee conversion..." -ForegroundColor Yellow
    $confirmEmployeeData = @{
        employee_id = $employeeId
        confirm_date = "2024-04-15"
        notes = "Good performance during probation, convert to permanent employee"
    } | ConvertTo-Json

    try {
        $confirmEmpResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/confirm-employee" -Method POST -Body $confirmEmployeeData -Headers $headers
        Write-Host "Employee conversion successful" -ForegroundColor Green
        Write-Host "Response: $($confirmEmpResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "Employee conversion failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error details: $errorBody" -ForegroundColor Red
        }
    }

    # 6. Get onboarding history
    Write-Host "`n6. Get onboarding history..." -ForegroundColor Yellow
    try {
        $historyResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/$employeeId/history" -Method GET -Headers $headers
        Write-Host "Get onboarding history successful" -ForegroundColor Green
        Write-Host "Response: $($historyResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "Get onboarding history failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error details: $errorBody" -ForegroundColor Red
        }
    }
}

# 7. Get onboarding workflow list
Write-Host "`n7. Get onboarding workflow list..." -ForegroundColor Yellow
try {
    $workflowsResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/workflows?page=1&page_size=10" -Method GET -Headers $headers
    Write-Host "Get onboarding workflow list successful" -ForegroundColor Green
    Write-Host "Response: $($workflowsResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
} catch {
    Write-Host "Get onboarding workflow list failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error details: $errorBody" -ForegroundColor Red
    }
}

# 8. Test status change
if ($employeeId) {
    Write-Host "`n8. Test employee status change..." -ForegroundColor Yellow
    $statusChangeData = @{
        employee_id = $employeeId
        new_status = "active"
        reason = "Admin manual status adjustment"
        notes = "Test status change functionality"
    } | ConvertTo-Json

    try {
        $statusResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/change-status" -Method POST -Body $statusChangeData -Headers $headers
        Write-Host "Employee status change successful" -ForegroundColor Green
        Write-Host "Response: $($statusResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "Employee status change failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error details: $errorBody" -ForegroundColor Red
        }
    }
}

Write-Host "`n=== Onboarding Workflow API Test Complete ===" -ForegroundColor Green
Write-Host "Please check the test results above to ensure all API endpoints work properly" -ForegroundColor Yellow
