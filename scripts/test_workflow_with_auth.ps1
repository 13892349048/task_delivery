# 带认证的工作流测试脚本
$baseUrl = "http://localhost:8081/api/v1"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "=== 工作流认证测试 ===" -ForegroundColor Green

# 0. 测试基础连接（无需认证）
Write-Host "0. 测试服务连接..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8081/" -Method GET -Headers $headers
    Write-Host "✓ 服务连接成功: $($response.service)" -ForegroundColor Green
} catch {
    Write-Host "✗ 服务连接失败: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 1. 用户登录获取Token
Write-Host "1. 用户登录..." -ForegroundColor Yellow
$loginRequest = @{
    username = "testuser"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginRequest -Headers $headers
    $token = $loginResponse.data.access_token
    Write-Host "✓ 登录成功，获取到Token" -ForegroundColor Green
    
    # 更新headers包含认证信息
    $headers["Authorization"] = "Bearer $token"
} catch {
    Write-Host "✗ 登录失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "请确保数据库中有admin用户，或修改登录凭据" -ForegroundColor Yellow
    exit 1
}

# 2. 测试获取工作流定义列表
Write-Host "2. 获取工作流定义列表..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/workflows/definitions" -Method GET -Headers $headers
    Write-Host "✓ 工作流定义获取成功，共 $($response.data.Count) 个定义" -ForegroundColor Green
    
    # 查找任务分配审批工作流
    $taskAssignmentWorkflow = $response.data | Where-Object { $_.name -eq "task_assignment_approval" }
    if ($taskAssignmentWorkflow) {
        $workflowId = $taskAssignmentWorkflow.id
        Write-Host "  找到任务分配审批工作流，ID: $workflowId" -ForegroundColor Cyan
    } else {
        Write-Host "  未找到任务分配审批工作流定义" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ 工作流定义获取失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 3. 测试启动任务分配审批流程
Write-Host "3. 启动任务分配审批流程..." -ForegroundColor Yellow
$startRequest = @{
    task_id = 1
    assignee_id = 2
    assignment_type = "manual"
    priority = "high"
    requester_id = 1
    reason = "测试任务分配审批"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/workflows/task-assignment/start" -Method POST -Body $startRequest -Headers $headers
    $instanceId = $response.data.id
    Write-Host "✓ 审批流程启动成功: $instanceId" -ForegroundColor Green
    
    # 4. 测试获取待审批任务
    Write-Host "4. 获取待审批任务..." -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/workflows/approvals/pending?user_id=101" -Method GET -Headers $headers
        Write-Host "✓ 待审批任务数量: $($response.data.Count)" -ForegroundColor Green
        
        if ($response.data.Count -gt 0) {
            Write-Host "待审批任务详情:" -ForegroundColor Cyan
            foreach ($approval in $response.data) {
                Write-Host "  - 实例ID: $($approval.instance_id)" -ForegroundColor White
                Write-Host "  - 节点ID: $($approval.node_id)" -ForegroundColor White
                Write-Host "  - 业务类型: $($approval.business_type)" -ForegroundColor White
                Write-Host "  - 分配给: $($approval.assigned_to)" -ForegroundColor White
            }
        }
    } catch {
        Write-Host "✗ 获取待审批任务失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    
    # 5. 测试获取任务分配审批
    Write-Host "5. 获取任务分配审批..." -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/workflows/approvals/task-assignments" -Method GET -Headers $headers
        Write-Host "✓ 任务分配审批数量: $($response.data.Count)" -ForegroundColor Green
    } catch {
        Write-Host "✗ 获取任务分配审批失败: $($_.Exception.Message)" -ForegroundColor Red
    }
    
} catch {
    Write-Host "✗ 启动审批流程失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "错误详情: $($_.ErrorDetails.Message)" -ForegroundColor Red
}

# 6. 测试其他基础API
Write-Host "6. 测试其他基础API..." -ForegroundColor Yellow

# 测试任务列表
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/tasks" -Method GET -Headers $headers
    Write-Host "✓ 任务列表获取成功，共 $($response.data.Count) 个任务" -ForegroundColor Green
} catch {
    Write-Host "✗ 任务列表获取失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 测试员工列表
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/employees" -Method GET -Headers $headers
    Write-Host "✓ 员工列表获取成功，共 $($response.data.Count) 个员工" -ForegroundColor Green
} catch {
    Write-Host "✗ 员工列表获取失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "=== 测试完成 ===" -ForegroundColor Green
