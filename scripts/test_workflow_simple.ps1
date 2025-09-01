# 简单的工作流测试脚本
$baseUrl = "http://localhost:8080/api/v1"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "=== 工作流测试 ===" -ForegroundColor Green

# 1. 测试获取工作流定义
Write-Host "1. 获取工作流定义..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/workflows/definitions/task_assignment_approval" -Method GET -Headers $headers
    Write-Host "✓ 工作流定义获取成功: $($response.data.name)" -ForegroundColor Green
} catch {
    Write-Host "✗ 工作流定义获取失败: $($_.Exception.Message)" -ForegroundColor Red
}

# 2. 测试启动审批流程
Write-Host "2. 启动任务分配审批流程..." -ForegroundColor Yellow
$startRequest = @{
    task_id = 1
    assignee_id = 2
    assignment_type = "manual"
    priority = "high"
    requester_id = 1
    reason = "测试任务分配审批"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/tasks/assignment-approval/start" -Method POST -Body $startRequest -Headers $headers
    $instanceId = $response.data.id
    Write-Host "✓ 审批流程启动成功: $instanceId" -ForegroundColor Green
    
    # 3. 测试获取待审批任务
    Write-Host "3. 获取待审批任务..." -ForegroundColor Yellow
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
    
} catch {
    Write-Host "✗ 启动审批流程失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "=== 测试完成 ===" -ForegroundColor Green
