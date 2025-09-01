# 入职工作流API测试脚本
# 测试所有入职工作流相关的API端点

$baseUrl = "http://localhost:8081/api/v1"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "=== 入职工作流API测试 ===" -ForegroundColor Green

# 1. 用户登录获取JWT Token
Write-Host "`n1. 用户登录..." -ForegroundColor Yellow
$loginData = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginData -Headers $headers
    $token = $loginResponse.data.token
    $headers["Authorization"] = "Bearer $token"
    Write-Host "登录成功，获取到Token" -ForegroundColor Green
} catch {
    Write-Host "登录失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 2. 创建待入职员工
Write-Host "`n2. 创建待入职员工..." -ForegroundColor Yellow
$pendingEmployeeData = @{
    real_name = "张三"
    email = "zhangsan@example.com"
    phone = "13800138001"
    expected_date = "2024-01-15"
    notes = "新员工入职，软件工程师岗位"
} | ConvertTo-Json

try {
    $createResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/pending" -Method POST -Body $pendingEmployeeData -Headers $headers
    $employeeId = $createResponse.data.employee_id
    Write-Host "创建待入职员工成功，员工ID: $employeeId" -ForegroundColor Green
    Write-Host "响应: $($createResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
} catch {
    Write-Host "创建待入职员工失败: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "错误详情: $errorBody" -ForegroundColor Red
    }
}

# 3. 确认入职（需要员工ID）
if ($employeeId) {
    Write-Host "`n3. 确认入职..." -ForegroundColor Yellow
    $confirmData = @{
        employee_id = $employeeId
        department_id = 1
        position_id = 1
        hire_date = "2024-01-15"
        probation_end_date = "2024-04-15"
        notes = "确认入职，分配到研发部门"
    } | ConvertTo-Json

    try {
        $confirmResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/confirm" -Method POST -Body $confirmData -Headers $headers
        Write-Host "确认入职成功" -ForegroundColor Green
        Write-Host "响应: $($confirmResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "确认入职失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "错误详情: $errorBody" -ForegroundColor Red
        }
    }

    # 4. 完成试用期
    Write-Host "`n4. 完成试用期..." -ForegroundColor Yellow
    try {
        $probationResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/$employeeId/probation" -Method POST -Headers $headers
        Write-Host "完成试用期成功" -ForegroundColor Green
        Write-Host "响应: $($probationResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "完成试用期失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "错误详情: $errorBody" -ForegroundColor Red
        }
    }

    # 5. 确认员工（试用期转正）
    Write-Host "`n5. 确认员工转正..." -ForegroundColor Yellow
    $confirmEmployeeData = @{
        employee_id = $employeeId
        confirm_date = "2024-04-15"
        notes = "试用期表现良好，转为正式员工"
    } | ConvertTo-Json

    try {
        $confirmEmpResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/confirm-employee" -Method POST -Body $confirmEmployeeData -Headers $headers
        Write-Host "员工转正成功" -ForegroundColor Green
        Write-Host "响应: $($confirmEmpResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "员工转正失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "错误详情: $errorBody" -ForegroundColor Red
        }
    }

    # 6. 获取入职历史记录
    Write-Host "`n6. 获取入职历史记录..." -ForegroundColor Yellow
    try {
        $historyResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/$employeeId/history" -Method GET -Headers $headers
        Write-Host "获取入职历史记录成功" -ForegroundColor Green
        Write-Host "响应: $($historyResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "获取入职历史记录失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "错误详情: $errorBody" -ForegroundColor Red
        }
    }
}

# 7. 获取入职工作流列表
Write-Host "`n7. 获取入职工作流列表..." -ForegroundColor Yellow
try {
    $workflowsResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/workflows?page=1&page_size=10" -Method GET -Headers $headers
    Write-Host "获取入职工作流列表成功" -ForegroundColor Green
    Write-Host "响应: $($workflowsResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
} catch {
    Write-Host "获取入职工作流列表失败: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "错误详情: $errorBody" -ForegroundColor Red
    }
}

# 8. 测试状态变更
if ($employeeId) {
    Write-Host "`n8. 测试员工状态变更..." -ForegroundColor Yellow
    $statusChangeData = @{
        employee_id = $employeeId
        new_status = "active"
        reason = "管理员手动调整状态"
        notes = "测试状态变更功能"
    } | ConvertTo-Json

    try {
        $statusResponse = Invoke-RestMethod -Uri "$baseUrl/onboarding/change-status" -Method POST -Body $statusChangeData -Headers $headers
        Write-Host "员工状态变更成功" -ForegroundColor Green
        Write-Host "响应: $($statusResponse | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    } catch {
        Write-Host "员工状态变更失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "错误详情: $errorBody" -ForegroundColor Red
        }
    }
}

Write-Host "`n=== 入职工作流API测试完成 ===" -ForegroundColor Green
Write-Host "请检查上述测试结果，确保所有API端点正常工作" -ForegroundColor Yellow
