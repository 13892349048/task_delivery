# 快速API测试脚本 - 企业级测试流程
# 用于快速验证当前实现的API功能

param(
    [string]$ServerPort = "8081",
    [string]$TestUser = "admin",
    [string]$TestPass = "admin123"
)

$BaseURL = "http://localhost:$ServerPort"
$ApiBase = "$BaseURL/api/v1"

Write-Host "=== TaskManage API 快速测试 ===" -ForegroundColor Cyan
Write-Host "服务器地址: $BaseURL" -ForegroundColor Green
Write-Host "开始执行企业级API测试流程..." -ForegroundColor Yellow

# 测试结果统计
$TestResults = @{
    Total = 0
    Passed = 0
    Failed = 0
    Errors = @()
}

# HTTP请求函数
function Invoke-APIRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Body = $null,
        [string]$Token = $null,
        [string]$Description
    )
    
    $TestResults.Total++
    Write-Host "`n[$($TestResults.Total)] 测试: $Description" -ForegroundColor White
    
    try {
        $headers = @{
            'Content-Type' = 'application/json'
        }
        
        if ($Token) {
            $headers['Authorization'] = "Bearer $Token"
        }
        
        $params = @{
            Uri = "$ApiBase$Endpoint"
            Method = $Method
            Headers = $headers
            TimeoutSec = 10
        }
        
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 10)
        }
        
        Write-Host "请求: $Method $Endpoint" -ForegroundColor Gray
        
        $response = Invoke-RestMethod @params
        
        if ($response.success) {
            Write-Host "✅ 通过" -ForegroundColor Green
            $TestResults.Passed++
            return $response
        } else {
            Write-Host "❌ 失败: $($response.message)" -ForegroundColor Red
            $TestResults.Failed++
            $TestResults.Errors += "$Description - API返回失败: $($response.message)"
            return $null
        }
        
    } catch {
        Write-Host "❌ 错误: $($_.Exception.Message)" -ForegroundColor Red
        $TestResults.Failed++
        $TestResults.Errors += "$Description - 请求异常: $($_.Exception.Message)"
        return $null
    }
}

# 等待服务器启动
function Wait-ForServer {
    Write-Host "等待服务器启动..." -ForegroundColor Yellow
    $maxAttempts = 30
    $attempt = 0
    
    do {
        try {
            $response = Invoke-WebRequest -Uri "$BaseURL/health" -TimeoutSec 2 -UseBasicParsing
            if ($response.StatusCode -eq 200) {
                Write-Host "✅ 服务器已启动" -ForegroundColor Green
                return $true
            }
        } catch {
            Start-Sleep -Seconds 1
            $attempt++
            Write-Host "." -NoNewline -ForegroundColor Gray
        }
    } while ($attempt -lt $maxAttempts)
    
    Write-Host "`n❌ 服务器启动超时" -ForegroundColor Red
    return $false
}

# 主测试流程
function Start-APITests {
    # 1. 健康检查
    Write-Host "`n=== 基础健康检查 ===" -ForegroundColor Cyan
    
    try {
        $health = Invoke-RestMethod -Uri "$BaseURL/health" -TimeoutSec 5
        Write-Host "✅ 健康检查通过: $($health.status)" -ForegroundColor Green
    } catch {
        Write-Host "❌ 健康检查失败: $($_.Exception.Message)" -ForegroundColor Red
        return
    }
    
    # 2. 认证测试
    Write-Host "`n=== 认证模块测试 ===" -ForegroundColor Cyan
    
    # 用户登录
    $loginData = @{
        username = $TestUser
        password = $TestPass
    }
    
    $loginResponse = Invoke-APIRequest -Method "POST" -Endpoint "/auth/login" -Body $loginData -Description "用户登录"
    
    if (-not $loginResponse) {
        Write-Host "❌ 无法获取认证Token，终止测试" -ForegroundColor Red
        return
    }
    
    $authToken = $loginResponse.data.access_token
    Write-Host "🔑 获取到认证Token" -ForegroundColor Green
    
    # 3. 员工管理模块测试
    Write-Host "`n=== 员工管理模块测试 ===" -ForegroundColor Cyan
    
    # 获取员工列表
    $employees = Invoke-APIRequest -Method "GET" -Endpoint "/employees?page=1&limit=10" -Token $authToken -Description "获取员工列表"
    
    # 创建测试员工
    $employeeData = @{
        user_id = 1
        employee_no = "EMP_TEST_$(Get-Date -Format 'yyyyMMddHHmmss')"
        department = "测试部门"
        position = "测试工程师"
        level = "中级"
        status = "active"
        max_tasks = 5
    }
    
    $newEmployee = Invoke-APIRequest -Method "POST" -Endpoint "/employees" -Body $employeeData -Token $authToken -Description "创建员工"
    
    if ($newEmployee) {
        $employeeId = $newEmployee.data.id
        
        # 更新员工状态
        $statusData = @{ status = "busy" }
        Invoke-APIRequest -Method "PUT" -Endpoint "/employees/$employeeId/status" -Body $statusData -Token $authToken -Description "更新员工状态"
        
        # 按状态查询员工
        Invoke-APIRequest -Method "GET" -Endpoint "/employees/status?status=busy" -Token $authToken -Description "按状态查询员工"
        
        # 获取员工工作负载
        Invoke-APIRequest -Method "GET" -Endpoint "/employees/$employeeId/workload" -Token $authToken -Description "获取员工工作负载"
    }
    
    # 获取工作负载统计
    Invoke-APIRequest -Method "GET" -Endpoint "/employees/workload/stats" -Token $authToken -Description "获取工作负载统计"
    
    # 获取部门工作负载
    Invoke-APIRequest -Method "GET" -Endpoint "/employees/workload/departments/测试部门" -Token $authToken -Description "获取部门工作负载"
    
    # 4. 技能管理模块测试
    Write-Host "`n=== 技能管理模块测试 ===" -ForegroundColor Cyan
    
    # 获取技能列表
    $skills = Invoke-APIRequest -Method "GET" -Endpoint "/skills" -Token $authToken -Description "获取技能列表"
    
    # 创建测试技能
    $skillData = @{
        name = "测试技能_$(Get-Date -Format 'HHmmss')"
        category = "测试分类"
        description = "这是一个测试技能"
        tags = @("test", "skill", "demo")
    }
    
    $newSkill = Invoke-APIRequest -Method "POST" -Endpoint "/skills" -Body $skillData -Token $authToken -Description "创建技能"
    
    if ($newSkill -and $newEmployee) {
        $skillId = $newSkill.data.id
        
        # 分配技能给员工
        $assignData = @{
            employee_id = $employeeId
            skill_id = $skillId
            level = 3
        }
        
        Invoke-APIRequest -Method "POST" -Endpoint "/skills/assign" -Body $assignData -Token $authToken -Description "分配技能给员工"
        
        # 获取员工技能
        Invoke-APIRequest -Method "GET" -Endpoint "/skills/employees/$employeeId" -Token $authToken -Description "获取员工技能列表"
    }
    
    # 获取技能分类
    Invoke-APIRequest -Method "GET" -Endpoint "/skills/categories" -Token $authToken -Description "获取技能分类"
    
    # 5. 错误处理测试
    Write-Host "`n=== 错误处理测试 ===" -ForegroundColor Cyan
    
    # 无效Token测试
    try {
        Invoke-RestMethod -Uri "$ApiBase/employees" -Headers @{'Authorization' = 'Bearer invalid_token'} -TimeoutSec 5
        Write-Host "❌ 无效Token测试失败 - 应该返回401错误" -ForegroundColor Red
    } catch {
        if ($_.Exception.Response.StatusCode -eq 401) {
            Write-Host "✅ 无效Token正确返回401错误" -ForegroundColor Green
            $TestResults.Passed++
        } else {
            Write-Host "❌ 无效Token返回了错误的状态码: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
            $TestResults.Failed++
        }
        $TestResults.Total++
    }
    
    # 不存在资源测试
    try {
        Invoke-RestMethod -Uri "$ApiBase/employees/99999" -Headers @{'Authorization' = "Bearer $authToken"} -TimeoutSec 5
        Write-Host "❌ 不存在资源测试失败 - 应该返回404错误" -ForegroundColor Red
    } catch {
        if ($_.Exception.Response.StatusCode -eq 404) {
            Write-Host "✅ 不存在资源正确返回404错误" -ForegroundColor Green
            $TestResults.Passed++
        } else {
            Write-Host "❌ 不存在资源返回了错误的状态码: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
            $TestResults.Failed++
        }
        $TestResults.Total++
    }
}

# 输出测试结果
function Show-TestResults {
    Write-Host "`n=== 测试结果摘要 ===" -ForegroundColor Cyan
    Write-Host "总测试数: $($TestResults.Total)" -ForegroundColor White
    Write-Host "通过: $($TestResults.Passed)" -ForegroundColor Green
    Write-Host "失败: $($TestResults.Failed)" -ForegroundColor Red
    
    $successRate = if ($TestResults.Total -gt 0) { 
        [math]::Round(($TestResults.Passed / $TestResults.Total) * 100, 2) 
    } else { 0 }
    
    Write-Host "成功率: $successRate%" -ForegroundColor $(if ($successRate -ge 90) { "Green" } elseif ($successRate -ge 70) { "Yellow" } else { "Red" })
    
    if ($TestResults.Errors.Count -gt 0) {
        Write-Host "`n错误详情:" -ForegroundColor Red
        foreach ($error in $TestResults.Errors) {
            Write-Host "  • $error" -ForegroundColor Red
        }
    }
    
    if ($successRate -ge 90) {
        Write-Host "`n🎉 测试通过！API功能正常" -ForegroundColor Green
    } elseif ($successRate -ge 70) {
        Write-Host "`n⚠️  测试部分通过，需要关注失败项" -ForegroundColor Yellow
    } else {
        Write-Host "`n💥 测试失败较多，需要修复问题" -ForegroundColor Red
    }
}

# 主执行流程
if (-not (Wait-ForServer)) {
    Write-Host "请先启动TaskManage服务器：" -ForegroundColor Yellow
    Write-Host "go run cmd/taskmanage/main.go --env=test" -ForegroundColor White
    exit 1
}

Start-APITests
Show-TestResults

Write-Host "`n测试完成！" -ForegroundColor Cyan
