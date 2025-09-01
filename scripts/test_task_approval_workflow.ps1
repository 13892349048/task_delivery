# 任务分配审批工作流完整测试脚本
# 测试从任务分配请求到审批完成的完整流程

Write-Host "=== 任务分配审批工作流完整测试 ===" -ForegroundColor Cyan

$BaseUrl = "http://localhost:8080/api/v1"
$Headers = @{
    "Content-Type" = "application/json"
}

# 测试用户凭据
$TestUsers = @{
    "requester" = @{
        "username" = "test_user_314"
        "password" = "password123"
        "token" = ""
    }
    "approver" = @{
        "username" = "admin"
        "password" = "admin123"
        "token" = ""
    }
}

# 全局变量
$TaskID = 0
$EmployeeID = 1
$WorkflowInstanceID = ""

function Test-APIEndpoint {
    param(
        [string]$Method,
        [string]$Url,
        [hashtable]$Headers,
        [string]$Body = $null
    )
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            Headers = $Headers
        }
        
        if ($Body) {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params
        return @{
            Success = $true
            Data = $response
            StatusCode = 200
        }
    }
    catch {
        return @{
            Success = $false
            Error = $_.Exception.Message
            StatusCode = $_.Exception.Response.StatusCode.value__
        }
    }
}

function Login-User {
    param(
        [string]$Username,
        [string]$Password
    )
    
    Write-Host "登录用户: $Username" -ForegroundColor Yellow
    
    $loginData = @{
        username = $Username
        password = $Password
    } | ConvertTo-Json
    
    $result = Test-APIEndpoint -Method "POST" -Url "$BaseUrl/auth/login" -Headers $Headers -Body $loginData
    
    if ($result.Success) {
        $token = $result.Data.data.token
        Write-Host "✅ 登录成功" -ForegroundColor Green
        return $token
    } else {
        Write-Host "❌ 登录失败: $($result.Error)" -ForegroundColor Red
        return $null
    }
}

function Create-TestTask {
    param([string]$Token)
    
    Write-Host "`n1. 创建测试任务..." -ForegroundColor Yellow
    
    $taskData = @{
        title = "审批流程测试任务 $(Get-Date -Format 'yyyyMMdd-HHmmss')"
        description = "用于测试任务分配审批工作流的测试任务"
        priority = "high"
        due_date = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
    } | ConvertTo-Json
    
    $taskHeaders = $Headers.Clone()
    $taskHeaders["Authorization"] = "Bearer $Token"
    
    $result = Test-APIEndpoint -Method "POST" -Url "$BaseUrl/tasks" -Headers $taskHeaders -Body $taskData
    
    if ($result.Success) {
        $script:TaskID = $result.Data.data.id
        Write-Host "✅ 任务创建成功 - TaskID: $script:TaskID" -ForegroundColor Green
        return $true
    } else {
        Write-Host "❌ 任务创建失败: $($result.Error)" -ForegroundColor Red
        return $false
    }
}

function Start-TaskAssignmentApproval {
    param([string]$Token)
    
    Write-Host "`n2. 启动任务分配审批流程..." -ForegroundColor Yellow
    
    $approvalData = @{
        task_id = $script:TaskID
        assignee_id = $script:EmployeeID
        assignment_type = "manual"
        priority = "high"
        reason = "测试任务分配审批流程"
    } | ConvertTo-Json
    
    $approvalHeaders = $Headers.Clone()
    $approvalHeaders["Authorization"] = "Bearer $Token"
    
    $result = Test-APIEndpoint -Method "POST" -Url "$BaseUrl/tasks/assignment-approval/start" -Headers $approvalHeaders -Body $approvalData
    
    if ($result.Success) {
        $script:WorkflowInstanceID = $result.Data.data.workflow_instance_id
        Write-Host "✅ 审批流程启动成功 - WorkflowInstanceID: $script:WorkflowInstanceID" -ForegroundColor Green
        Write-Host "   状态: $($result.Data.data.status)" -ForegroundColor Cyan
        return $true
    } else {
        Write-Host "❌ 审批流程启动失败: $($result.Error)" -ForegroundColor Red
        return $false
    }
}

function Get-PendingApprovals {
    param([string]$Token)
    
    Write-Host "`n3. 获取待审批任务列表..." -ForegroundColor Yellow
    
    $approvalHeaders = $Headers.Clone()
    $approvalHeaders["Authorization"] = "Bearer $Token"
    
    $result = Test-APIEndpoint -Method "GET" -Url "$BaseUrl/workflows/approvals/pending" -Headers $approvalHeaders
    
    if ($result.Success) {
        $approvals = $result.Data.data
        Write-Host "✅ 获取待审批任务成功 - 数量: $($approvals.Count)" -ForegroundColor Green
        
        foreach ($approval in $approvals) {
            Write-Host "   - 实例ID: $($approval.instance_id)" -ForegroundColor Cyan
            Write-Host "   - 工作流: $($approval.workflow_name)" -ForegroundColor Cyan
            Write-Host "   - 节点: $($approval.node_name)" -ForegroundColor Cyan
            Write-Host "   - 业务ID: $($approval.business_id)" -ForegroundColor Cyan
        }
        return $approvals
    } else {
        Write-Host "❌ 获取待审批任务失败: $($result.Error)" -ForegroundColor Red
        return @()
    }
}

function Get-TaskAssignmentApprovals {
    param([string]$Token)
    
    Write-Host "`n4. 获取任务分配待审批列表..." -ForegroundColor Yellow
    
    $approvalHeaders = $Headers.Clone()
    $approvalHeaders["Authorization"] = "Bearer $Token"
    
    $result = Test-APIEndpoint -Method "GET" -Url "$BaseUrl/workflows/approvals/task-assignments" -Headers $approvalHeaders
    
    if ($result.Success) {
        $approvals = $result.Data.data
        Write-Host "✅ 获取任务分配待审批成功 - 数量: $($approvals.Count)" -ForegroundColor Green
        
        foreach ($approval in $approvals) {
            Write-Host "   - 任务ID: $($approval.task_id)" -ForegroundColor Cyan
            Write-Host "   - 实例ID: $($approval.instance_id)" -ForegroundColor Cyan
            Write-Host "   - 优先级: $($approval.priority)" -ForegroundColor Cyan
        }
        return $approvals
    } else {
        Write-Host "❌ 获取任务分配待审批失败: $($result.Error)" -ForegroundColor Red
        return @()
    }
}

function Process-Approval {
    param(
        [string]$Token,
        [string]$InstanceID,
        [string]$NodeID,
        [string]$Action,
        [string]$Comment
    )
    
    Write-Host "`n5. 处理审批决策 ($Action)..." -ForegroundColor Yellow
    
    $processData = @{
        instance_id = $InstanceID
        node_id = $NodeID
        action = $Action
        comment = $Comment
        variables = @{}
    } | ConvertTo-Json
    
    $processHeaders = $Headers.Clone()
    $processHeaders["Authorization"] = "Bearer $Token"
    
    $result = Test-APIEndpoint -Method "POST" -Url "$BaseUrl/workflows/approvals/process" -Headers $processHeaders -Body $processData
    
    if ($result.Success) {
        Write-Host "✅ 审批处理成功" -ForegroundColor Green
        Write-Host "   结果: $($result.Data.data.result)" -ForegroundColor Cyan
        Write-Host "   状态: $($result.Data.data.status)" -ForegroundColor Cyan
        return $true
    } else {
        Write-Host "❌ 审批处理失败: $($result.Error)" -ForegroundColor Red
        return $false
    }
}

function Verify-TaskStatus {
    param([string]$Token)
    
    Write-Host "`n6. 验证任务状态..." -ForegroundColor Yellow
    
    $taskHeaders = $Headers.Clone()
    $taskHeaders["Authorization"] = "Bearer $Token"
    
    $result = Test-APIEndpoint -Method "GET" -Url "$BaseUrl/tasks/$script:TaskID" -Headers $taskHeaders
    
    if ($result.Success) {
        $task = $result.Data.data
        Write-Host "✅ 任务状态验证成功" -ForegroundColor Green
        Write-Host "   任务ID: $($task.id)" -ForegroundColor Cyan
        Write-Host "   状态: $($task.status)" -ForegroundColor Cyan
        Write-Host "   分配人ID: $($task.assignee_id)" -ForegroundColor Cyan
        Write-Host "   标题: $($task.title)" -ForegroundColor Cyan
        return $task
    } else {
        Write-Host "❌ 任务状态验证失败: $($result.Error)" -ForegroundColor Red
        return $null
    }
}

function Get-WorkflowInstance {
    param(
        [string]$Token,
        [string]$InstanceID
    )
    
    Write-Host "`n7. 获取工作流实例状态..." -ForegroundColor Yellow
    
    $workflowHeaders = $Headers.Clone()
    $workflowHeaders["Authorization"] = "Bearer $Token"
    
    $result = Test-APIEndpoint -Method "GET" -Url "$BaseUrl/workflows/instances/$InstanceID" -Headers $workflowHeaders
    
    if ($result.Success) {
        $instance = $result.Data.data
        Write-Host "✅ 工作流实例获取成功" -ForegroundColor Green
        Write-Host "   实例ID: $($instance.id)" -ForegroundColor Cyan
        Write-Host "   状态: $($instance.status)" -ForegroundColor Cyan
        Write-Host "   业务ID: $($instance.business_id)" -ForegroundColor Cyan
        return $instance
    } else {
        Write-Host "❌ 工作流实例获取失败: $($result.Error)" -ForegroundColor Red
        return $null
    }
}

# 主测试流程
Write-Host "`n开始测试任务分配审批工作流..." -ForegroundColor Green

# 步骤1：登录请求用户
Write-Host "`n=== 步骤1：用户认证 ===" -ForegroundColor Magenta
$requesterToken = Login-User -Username $TestUsers.requester.username -Password $TestUsers.requester.password
if (-not $requesterToken) {
    Write-Host "❌ 请求用户登录失败，测试终止" -ForegroundColor Red
    exit 1
}

$approverToken = Login-User -Username $TestUsers.approver.username -Password $TestUsers.approver.password
if (-not $approverToken) {
    Write-Host "❌ 审批用户登录失败，测试终止" -ForegroundColor Red
    exit 1
}

# 步骤2：创建测试任务
Write-Host "`n=== 步骤2：任务创建 ===" -ForegroundColor Magenta
if (-not (Create-TestTask -Token $requesterToken)) {
    Write-Host "❌ 任务创建失败，测试终止" -ForegroundColor Red
    exit 1
}

# 步骤3：启动审批流程
Write-Host "`n=== 步骤3：启动审批流程 ===" -ForegroundColor Magenta
if (-not (Start-TaskAssignmentApproval -Token $requesterToken)) {
    Write-Host "❌ 审批流程启动失败，测试终止" -ForegroundColor Red
    exit 1
}

# 步骤4：获取待审批任务
Write-Host "`n=== 步骤4：查询待审批任务 ===" -ForegroundColor Magenta
$pendingApprovals = Get-PendingApprovals -Token $approverToken
$taskAssignmentApprovals = Get-TaskAssignmentApprovals -Token $approverToken

# 步骤5：处理审批
Write-Host "`n=== 步骤5：处理审批决策 ===" -ForegroundColor Magenta
if ($pendingApprovals.Count -gt 0) {
    $approval = $pendingApprovals[0]
    $success = Process-Approval -Token $approverToken -InstanceID $approval.instance_id -NodeID $approval.node_id -Action "approve" -Comment "测试审批通过"
    
    if (-not $success) {
        Write-Host "❌ 审批处理失败，测试终止" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "⚠️ 没有找到待审批任务" -ForegroundColor Yellow
}

# 步骤6：验证结果
Write-Host "`n=== 步骤6：验证审批结果 ===" -ForegroundColor Magenta
Start-Sleep -Seconds 2  # 等待工作流处理完成

$finalTask = Verify-TaskStatus -Token $requesterToken
$workflowInstance = Get-WorkflowInstance -Token $approverToken -InstanceID $script:WorkflowInstanceID

# 步骤7：结果总结
Write-Host "`n=== 测试结果总结 ===" -ForegroundColor Magenta

$testResults = @{
    "任务创建" = ($script:TaskID -gt 0)
    "审批流程启动" = ($script:WorkflowInstanceID -ne "")
    "待审批任务查询" = ($pendingApprovals.Count -gt 0)
    "审批处理" = ($finalTask -and $finalTask.status -eq "assigned")
    "工作流完成" = ($workflowInstance -and $workflowInstance.status -eq "completed")
}

$successCount = 0
$totalCount = $testResults.Count

foreach ($test in $testResults.GetEnumerator()) {
    if ($test.Value) {
        Write-Host "✅ $($test.Key): 通过" -ForegroundColor Green
        $successCount++
    } else {
        Write-Host "❌ $($test.Key): 失败" -ForegroundColor Red
    }
}

Write-Host "`n=== 最终结果 ===" -ForegroundColor Cyan
Write-Host "测试通过率: $successCount/$totalCount ($([math]::Round($successCount/$totalCount*100, 2))%)" -ForegroundColor $(if ($successCount -eq $totalCount) { "Green" } else { "Yellow" })

if ($successCount -eq $totalCount) {
    Write-Host "🎉 任务分配审批工作流测试完全通过！" -ForegroundColor Green
    exit 0
} else {
    Write-Host "⚠️ 部分测试未通过，请检查系统配置" -ForegroundColor Yellow
    exit 1
}
