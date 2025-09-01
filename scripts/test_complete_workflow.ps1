# 完整工作流实例生命周期测试脚本
# 测试从任务分配申请到审批完成的完整流程

param(
    [string]$ServerPort = "8081",
    [string]$TestUser = "admin",
    [string]$TestPass = "admin123"
)

$BaseURL = "http://localhost:$ServerPort"
$ApiBase = "$BaseURL/api/v1"

Write-Host "=== 完整工作流实例生命周期测试 ===" -ForegroundColor Cyan
Write-Host "服务器地址: $BaseURL" -ForegroundColor Green

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
            TimeoutSec = 15
        }
        
        if ($Body) {
            $params.Body = ($Body | ConvertTo-Json -Depth 10)
            Write-Host "请求体: $($params.Body)" -ForegroundColor Gray
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
    Write-Host "检查服务器状态..." -ForegroundColor Yellow
    $maxAttempts = 10
    $attempt = 0
    
    do {
        try {
            $response = Invoke-WebRequest -Uri "$BaseURL/health" -TimeoutSec 3 -UseBasicParsing
            if ($response.StatusCode -eq 200) {
                Write-Host "✅ 服务器运行正常" -ForegroundColor Green
                return $true
            }
        } catch {
            Start-Sleep -Seconds 1
            $attempt++
            Write-Host "." -NoNewline -ForegroundColor Gray
        }
    } while ($attempt -lt $maxAttempts)
    
    Write-Host "`n❌ 无法连接到服务器" -ForegroundColor Red
    return $false
}

# 主测试流程
function Start-CompleteWorkflowTests {
    Write-Host "`n=== 阶段1: 系统准备 ===" -ForegroundColor Cyan
    
    # 用户登录
    $loginData = @{
        username = $TestUser
        password = $TestPass
    }
    
    $loginResponse = Invoke-APIRequest -Method "POST" -Endpoint "/auth/login" -Body $loginData -Description "管理员登录"
    
    if (-not $loginResponse) {
        Write-Host "❌ 登录失败，终止测试" -ForegroundColor Red
        return
    }
    
    $authToken = $loginResponse.data.access_token
    Write-Host "🔑 认证Token获取成功" -ForegroundColor Green
    
    # 检查工作流定义
    $workflowDef = Invoke-APIRequest -Method "GET" -Endpoint "/workflows/definitions" -Token $authToken -Description "获取工作流定义列表"
    
    if ($workflowDef -and $workflowDef.data) {
        Write-Host "✅ 工作流定义已加载，共 $($workflowDef.data.Count) 个定义" -ForegroundColor Green
        foreach ($def in $workflowDef.data) {
            Write-Host "  - $($def.id): $($def.name)" -ForegroundColor Gray
        }
    }
    
    Write-Host "`n=== 阶段2: 创建测试数据 ===" -ForegroundColor Cyan
    
    # 创建测试任务
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $taskData = @{
        title = "工作流测试任务-$timestamp"
        description = "用于测试完整工作流生命周期的任务"
        priority = "high"  # 使用高优先级触发更复杂的审批流程
        due_date = (Get-Date).AddDays(5).ToString("yyyy-MM-ddTHH:mm:ssZ")
    }
    
    $newTask = Invoke-APIRequest -Method "POST" -Endpoint "/tasks" -Body $taskData -Token $authToken -Description "创建高优先级测试任务"
    
    if (-not $newTask) {
        Write-Host "❌ 无法创建测试任务，终止测试" -ForegroundColor Red
        return
    }
    
    $taskId = $newTask.data.id
    Write-Host "✅ 测试任务创建成功: ID=$taskId, Priority=$($newTask.data.priority)" -ForegroundColor Green
    
    # 获取可用员工列表
    $employees = Invoke-APIRequest -Method "GET" -Endpoint "/employees?status=active&limit=5" -Token $authToken -Description "获取可用员工列表"
    
    $employeeId = 1  # 默认员工ID
    if ($employees -and $employees.data -and $employees.data.Count -gt 0) {
        $employeeId = $employees.data[0].id
        Write-Host "✅ 选择员工进行分配: ID=$employeeId" -ForegroundColor Green
    } else {
        Write-Host "⚠️  使用默认员工ID: $employeeId" -ForegroundColor Yellow
    }
    
    Write-Host "`n=== 阶段3: 触发工作流审批 ===" -ForegroundColor Cyan
    
    # 申请任务分配（应触发工作流）
    $assignmentData = @{
        assignee_id = $employeeId
        reason = "测试完整工作流实例生命周期 - $timestamp"
        method = "manual"
    }
    
    $assignmentResult = Invoke-APIRequest -Method "POST" -Endpoint "/tasks/$taskId/assign" -Body $assignmentData -Token $authToken -Description "申请任务分配（触发工作流）"
    
    if (-not $assignmentResult) {
        Write-Host "❌ 任务分配申请失败" -ForegroundColor Red
        return
    }
    
    Write-Host "✅ 任务分配申请提交成功" -ForegroundColor Green
    Write-Host "状态: $($assignmentResult.data.status)" -ForegroundColor Yellow
    Write-Host "工作流实例ID: $($assignmentResult.data.workflow_instance_id)" -ForegroundColor Yellow
    Write-Host "备注: $($assignmentResult.data.comment)" -ForegroundColor Gray
    
    $workflowInstanceId = $assignmentResult.data.workflow_instance_id
    
    # 验证工作流是否正确启动
    if ($assignmentResult.data.status -eq "pending_approval" -and $workflowInstanceId) {
        Write-Host "🎉 工作流审批已成功触发！" -ForegroundColor Green
    } else {
        Write-Host "⚠️  工作流可能未正确启动，状态: $($assignmentResult.data.status)" -ForegroundColor Yellow
    }
    
    Write-Host "`n=== 阶段4: 查看工作流实例详情 ===" -ForegroundColor Cyan
    
    if ($workflowInstanceId) {
        # 获取工作流实例详情
        $instance = Invoke-APIRequest -Method "GET" -Endpoint "/workflows/instances/$workflowInstanceId" -Token $authToken -Description "获取工作流实例详情"
        
        if ($instance -and $instance.data) {
            Write-Host "✅ 工作流实例详情:" -ForegroundColor Green
            Write-Host "  实例ID: $($instance.data.id)" -ForegroundColor Gray
            Write-Host "  工作流ID: $($instance.data.workflow_id)" -ForegroundColor Gray
            Write-Host "  业务ID: $($instance.data.business_id)" -ForegroundColor Gray
            Write-Host "  业务类型: $($instance.data.business_type)" -ForegroundColor Gray
            Write-Host "  状态: $($instance.data.status)" -ForegroundColor Gray
            Write-Host "  当前节点: $($instance.data.current_nodes -join ', ')" -ForegroundColor Gray
            Write-Host "  启动时间: $($instance.data.started_at)" -ForegroundColor Gray
        }
    }
    
    Write-Host "`n=== 阶段5: 管理员处理审批 ===" -ForegroundColor Cyan
    
    # 获取待审批列表
    $pendingApprovals = Invoke-APIRequest -Method "GET" -Endpoint "/workflows/approvals/pending" -Token $authToken -Description "获取待审批列表"
    
    if ($pendingApprovals -and $pendingApprovals.data -and $pendingApprovals.data.Count -gt 0) {
        Write-Host "✅ 发现 $($pendingApprovals.data.Count) 个待审批项目" -ForegroundColor Green
        
        # 找到我们刚创建的审批项目
        $targetApproval = $null
        foreach ($approval in $pendingApprovals.data) {
            if ($approval.instance_id -eq $workflowInstanceId) {
                $targetApproval = $approval
                break
            }
        }
        
        if ($targetApproval) {
            Write-Host "✅ 找到目标审批项目:" -ForegroundColor Green
            Write-Host "  实例ID: $($targetApproval.instance_id)" -ForegroundColor Gray
            Write-Host "  节点ID: $($targetApproval.node_id)" -ForegroundColor Gray
            Write-Host "  节点名称: $($targetApproval.node_name)" -ForegroundColor Gray
            Write-Host "  业务ID: $($targetApproval.business_id)" -ForegroundColor Gray
            Write-Host "  优先级: $($targetApproval.priority)" -ForegroundColor Gray
            
            # 处理审批（批准）
            $approvalData = @{
                instance_id = $targetApproval.instance_id
                node_id = $targetApproval.node_id
                action = "approve"
                comment = "测试审批通过 - 自动化测试 $timestamp"
            }
            
            $approvalResult = Invoke-APIRequest -Method "POST" -Endpoint "/workflows/approvals/process" -Body $approvalData -Token $authToken -Description "处理审批（批准）"
            
            if ($approvalResult) {
                Write-Host "✅ 审批处理成功" -ForegroundColor Green
                Write-Host "结果状态: $($approvalResult.data.action)" -ForegroundColor Yellow
                Write-Host "是否完成: $($approvalResult.data.is_completed)" -ForegroundColor Yellow
                Write-Host "消息: $($approvalResult.data.message)" -ForegroundColor Gray
                
                if ($approvalResult.data.is_completed) {
                    Write-Host "🎉 工作流已完成！" -ForegroundColor Green
                } else {
                    Write-Host "⏳ 工作流继续执行中..." -ForegroundColor Yellow
                }
            }
        } else {
            Write-Host "⚠️  未找到对应的审批项目" -ForegroundColor Yellow
        }
    } else {
        Write-Host "⚠️  当前没有待审批项目" -ForegroundColor Yellow
    }
    
    Write-Host "`n=== 阶段6: 验证最终结果 ===" -ForegroundColor Cyan
    
    # 等待一下让系统处理完成
    Start-Sleep -Seconds 3
    
    # 检查任务最终状态
    $finalTask = Invoke-APIRequest -Method "GET" -Endpoint "/tasks/$taskId" -Token $authToken -Description "检查任务最终状态"
    
    if ($finalTask) {
        Write-Host "✅ 任务最终状态: $($finalTask.data.status)" -ForegroundColor Green
        if ($finalTask.data.status -eq "assigned") {
            Write-Host "🎉 任务分配成功完成！" -ForegroundColor Green
        } elseif ($finalTask.data.status -eq "pending_approval") {
            Write-Host "⏳ 任务仍在审批中" -ForegroundColor Yellow
        } else {
            Write-Host "ℹ️  任务状态: $($finalTask.data.status)" -ForegroundColor Blue
        }
    }
    
    # 检查工作流实例最终状态
    if ($workflowInstanceId) {
        $finalInstance = Invoke-APIRequest -Method "GET" -Endpoint "/workflows/instances/$workflowInstanceId" -Token $authToken -Description "检查工作流实例最终状态"
        
        if ($finalInstance) {
            Write-Host "✅ 工作流实例最终状态: $($finalInstance.data.status)" -ForegroundColor Green
            if ($finalInstance.data.completed_at) {
                Write-Host "完成时间: $($finalInstance.data.completed_at)" -ForegroundColor Gray
            }
        }
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
    
    Write-Host "成功率: $successRate%" -ForegroundColor $(if ($successRate -ge 85) { "Green" } elseif ($successRate -ge 70) { "Yellow" } else { "Red" })
    
    if ($TestResults.Errors.Count -gt 0) {
        Write-Host "`n错误详情:" -ForegroundColor Red
        foreach ($error in $TestResults.Errors) {
            Write-Host "  • $error" -ForegroundColor Red
        }
    }
    
    Write-Host "`n=== 测试结论 ===" -ForegroundColor Cyan
    if ($successRate -ge 85) {
        Write-Host "🎉 完整工作流实例生命周期测试通过！" -ForegroundColor Green
        Write-Host "✅ 任务分配审批工作流系统运行正常" -ForegroundColor Green
    } elseif ($successRate -ge 70) {
        Write-Host "⚠️  工作流系统基本正常，但有部分功能需要关注" -ForegroundColor Yellow
    } else {
        Write-Host "💥 工作流系统存在问题，需要检查和修复" -ForegroundColor Red
    }
}

# 主执行流程
Write-Host "开始完整工作流实例生命周期测试..." -ForegroundColor Yellow
Write-Host "请确保TaskManage服务器正在运行" -ForegroundColor Yellow
Write-Host ""

if (-not (Wait-ForServer)) {
    Write-Host "请启动服务器: go run cmd/taskmanage/main.go" -ForegroundColor Red
    exit 1
}

Start-CompleteWorkflowTests
Show-TestResults

Write-Host "`n完整工作流测试完成！" -ForegroundColor Cyan
