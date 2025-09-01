# 认证工作流测试脚本
$baseUrl = "http://localhost:8081/api/v1"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "=== 认证工作流测试 ===" -ForegroundColor Green

# 1. 用户登录获取Token
Write-Host "1. 用户登录..." -ForegroundColor Yellow
$loginRequest = @{
    username = "testuser"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginRequest -Headers $headers
    $token = $loginResponse.data.access_token
    $userId = $loginResponse.data.user.id
    Write-Host "✓ 登录成功，用户ID: $userId" -ForegroundColor Green
    
    # 更新headers包含认证信息
    $headers["Authorization"] = "Bearer $token"
    
    # 2. 启动任务分配审批流程
    Write-Host "2. 启动任务分配审批流程..." -ForegroundColor Yellow
    $startRequest = @{
        task_id = 1
        assignee_id = 2
        assignment_type = "manual"
        priority = "high"
        reason = "测试任务分配审批"
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/workflows/task-assignment/start" -Method POST -Body $startRequest -Headers $headers
        $instanceId = $response.data.id
        Write-Host "✓ 审批流程启动成功: $instanceId" -ForegroundColor Green
        
        # 等待系统处理
        Start-Sleep -Seconds 2
        
        # 3. 检查待审批任务 - 检查管理员用户ID+100 (模拟直属上级)
        $managerUserId = $userId + 100
        Write-Host "3. 检查管理员用户 $managerUserId 的待审批任务..." -ForegroundColor Yellow
        try {
            $response = Invoke-RestMethod -Uri "$baseUrl/workflows/approvals/pending?user_id=$managerUserId" -Method GET -Headers $headers
            Write-Host "✓ 待审批任务数量: $($response.data.Count)" -ForegroundColor Green
            
            if ($response.data -and $response.data.Count -gt 0) {
                Write-Host "待审批任务详情:" -ForegroundColor Cyan
                foreach ($approval in $response.data) {
                    Write-Host "  - 实例ID: $($approval.instance_id)" -ForegroundColor White
                    Write-Host "  - 节点ID: $($approval.node_id)" -ForegroundColor White
                    Write-Host "  - 业务类型: $($approval.business_type)" -ForegroundColor White
                    Write-Host "  - 分配给: $($approval.assigned_to)" -ForegroundColor White
                    Write-Host "  - 工作流名称: $($approval.workflow_name)" -ForegroundColor White
                }
            } else {
                Write-Host "没有找到待审批任务，可能的原因:" -ForegroundColor Yellow
                Write-Host "  1. 工作流定义未正确初始化" -ForegroundColor Yellow
                Write-Host "  2. 审批节点执行器未正确创建待审批记录" -ForegroundColor Yellow
                Write-Host "  3. 用户ID计算错误 (当前用户: $userId, 预期管理员: $managerUserId)" -ForegroundColor Yellow
            }
        } catch {
            Write-Host "✗ 获取待审批任务失败: $($_.Exception.Message)" -ForegroundColor Red
        }
        
    } catch {
        Write-Host "✗ 启动审批流程失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails.Message) {
            Write-Host "错误详情: $($_.ErrorDetails.Message)" -ForegroundColor Red
        }
    }
    
} catch {
    Write-Host "✗ 登录失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "请确保:" -ForegroundColor Yellow
    Write-Host "  1. 服务器正在运行在 http://localhost:8081" -ForegroundColor Yellow
    Write-Host "  2. 数据库中有admin用户，密码为admin123" -ForegroundColor Yellow
    Write-Host "  3. 或修改登录凭据" -ForegroundColor Yellow
}

Write-Host "=== 测试完成 ===" -ForegroundColor Green
