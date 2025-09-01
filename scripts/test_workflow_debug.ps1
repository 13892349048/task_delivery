# 工作流调试测试脚本
$baseUrl = "http://localhost:8081/api/v1"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "=== 工作流调试测试 ===" -ForegroundColor Green

# 1. 先测试登录
Write-Host "1. 测试登录..." -ForegroundColor Yellow
$loginRequest = @{
    username = "admin"
    password = "admin123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginRequest -Headers $headers
    $token = $loginResponse.data.access_token
    $userId = $loginResponse.data.user.id
    Write-Host "✓ 登录成功，用户ID: $userId, Token前10位: $($token.Substring(0,10))..." -ForegroundColor Green
    
    # 更新headers包含认证信息
    $headers["Authorization"] = "Bearer $token"
    
    # 2. 测试启动工作流
    Write-Host "2. 启动任务分配审批流程..." -ForegroundColor Yellow
    $startRequest = @{
        task_id = 1
        assignee_id = 2
        assignment_type = "manual"
        priority = "high"
        reason = "调试测试"
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/workflows/task-assignment/start" -Method POST -Body $startRequest -Headers $headers
        Write-Host "✓ 工作流启动成功" -ForegroundColor Green
        Write-Host "实例ID: $($response.data.id)" -ForegroundColor Cyan
        
        # 等待处理
        Start-Sleep -Seconds 3
        
        # 3. 检查多个用户的待审批任务
        $testUsers = @($userId, ($userId + 100), 1, 2, 101, 102)
        foreach ($testUserId in $testUsers) {
            Write-Host "3. 检查用户 $testUserId 的待审批任务..." -ForegroundColor Yellow
            try {
                $response = Invoke-RestMethod -Uri "$baseUrl/workflows/approvals/pending?user_id=$testUserId" -Method GET -Headers $headers
                if ($response.data -and $response.data.Count -gt 0) {
                    Write-Host "✓ 找到 $($response.data.Count) 个待审批任务 (用户 $testUserId)" -ForegroundColor Green
                    foreach ($approval in $response.data) {
                        Write-Host "  实例: $($approval.instance_id), 节点: $($approval.node_id), 分配给: $($approval.assigned_to)" -ForegroundColor White
                    }
                    break
                } else {
                    Write-Host "- 用户 $testUserId: 无待审批任务" -ForegroundColor Gray
                }
            } catch {
                Write-Host "✗ 用户 $testUserId 查询失败: $($_.Exception.Message)" -ForegroundColor Red
            }
        }
        
    } catch {
        Write-Host "✗ 启动工作流失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails.Message) {
            Write-Host "详情: $($_.ErrorDetails.Message)" -ForegroundColor Red
        }
    }
    
} catch {
    Write-Host "✗ 登录失败: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "可能的原因:" -ForegroundColor Yellow
    Write-Host "  1. 服务器未运行在 localhost:8081" -ForegroundColor Yellow
    Write-Host "  2. admin用户不存在或密码错误" -ForegroundColor Yellow
    Write-Host "  3. 数据库连接问题" -ForegroundColor Yellow
}

Write-Host "=== 调试完成 ===" -ForegroundColor Green
