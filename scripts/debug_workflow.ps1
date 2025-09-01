# 调试工作流脚本
$baseUrl = "http://localhost:8081/api/v1"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "=== 工作流调试 ===" -ForegroundColor Green

# 1. 先启动一个任务分配审批流程
Write-Host "1. 启动任务分配审批流程..." -ForegroundColor Yellow
$startRequest = @{
    task_id = 1
    assignee_id = 2
    assignment_type = "manual"
    priority = "high"
    requester_id = 1
    reason = "调试测试任务分配审批"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/workflows/task-assignment/start" -Method POST -Body $startRequest -Headers $headers
    Write-Host "✓ 审批流程启动成功" -ForegroundColor Green
    Write-Host "响应数据: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Cyan
    
    if ($response.data -and $response.data.id) {
        $instanceId = $response.data.id
        Write-Host "实例ID: $instanceId" -ForegroundColor White
        
        # 等待一下让系统处理
        Start-Sleep -Seconds 2
        
        # 2. 检查待审批任务 - 使用不同的用户ID
        $userIds = @(101, 102, 1, 2, 3)
        foreach ($userId in $userIds) {
            Write-Host "2. 检查用户 $userId 的待审批任务..." -ForegroundColor Yellow
            try {
                $response = Invoke-RestMethod -Uri "$baseUrl/workflows/approvals/pending?user_id=$userId" -Method GET -Headers $headers
                Write-Host "✓ 用户 $userId 待审批任务数量: $($response.data.Count)" -ForegroundColor Green
                
                if ($response.data -and $response.data.Count -gt 0) {
                    Write-Host "找到待审批任务!" -ForegroundColor Green
                    foreach ($approval in $response.data) {
                        Write-Host "  - 实例ID: $($approval.instance_id)" -ForegroundColor White
                        Write-Host "  - 节点ID: $($approval.node_id)" -ForegroundColor White
                        Write-Host "  - 业务类型: $($approval.business_type)" -ForegroundColor White
                        Write-Host "  - 分配给: $($approval.assigned_to)" -ForegroundColor White
                    }
                    break
                }
            } catch {
                Write-Host "✗ 获取用户 $userId 待审批任务失败: $($_.Exception.Message)" -ForegroundColor Red
            }
        }
    }
    
} catch {
    Write-Host "✗ 启动审批流程失败: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "错误详情: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
}

Write-Host "=== 调试完成 ===" -ForegroundColor Green
