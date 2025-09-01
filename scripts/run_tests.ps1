# TaskManage API 企业级测试执行脚本
# 执行完整的测试套件：单元测试、集成测试、性能测试

param(
    [string]$TestType = "all",  # all, unit, integration, performance
    [string]$Environment = "test",
    [switch]$Verbose,
    [switch]$Coverage,
    [switch]$Report
)

# 颜色输出函数
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Success { Write-ColorOutput Green $args }
function Write-Warning { Write-ColorOutput Yellow $args }
function Write-Error { Write-ColorOutput Red $args }
function Write-Info { Write-ColorOutput Cyan $args }

# 测试配置
$TestConfig = @{
    ProjectRoot = Split-Path -Parent $PSScriptRoot
    TestTimeout = "10m"
    CoverageOutput = "coverage.out"
    ReportOutput = "test_report.html"
}

Write-Info "=== TaskManage API 企业级测试套件 ==="
Write-Info "测试类型: $TestType"
Write-Info "环境: $Environment"
Write-Info "项目根目录: $($TestConfig.ProjectRoot)"

# 设置环境变量
$env:APP_ENV = $Environment
$env:GO_ENV = $Environment

# 切换到项目根目录
Set-Location $TestConfig.ProjectRoot

# 检查Go环境
Write-Info "检查Go环境..."
try {
    $goVersion = go version
    Write-Success "Go版本: $goVersion"
} catch {
    Write-Error "Go环境未安装或配置错误"
    exit 1
}

# 检查依赖
Write-Info "检查项目依赖..."
go mod tidy
if ($LASTEXITCODE -ne 0) {
    Write-Error "依赖检查失败"
    exit 1
}

# 启动测试数据库（如果需要）
function Start-TestDatabase {
    Write-Info "启动测试数据库..."
    # 这里可以添加启动测试数据库的逻辑
    # 例如：docker-compose -f docker-compose.test.yml up -d mysql redis
}

# 停止测试数据库
function Stop-TestDatabase {
    Write-Info "停止测试数据库..."
    # docker-compose -f docker-compose.test.yml down
}

# 执行单元测试
function Run-UnitTests {
    Write-Info "=== 执行单元测试 ==="
    
    $testArgs = @(
        "test"
        "./internal/..."
        "-v"
        "-timeout=$($TestConfig.TestTimeout)"
    )
    
    if ($Coverage) {
        $testArgs += "-coverprofile=$($TestConfig.CoverageOutput)"
        $testArgs += "-covermode=atomic"
    }
    
    if ($Verbose) {
        $testArgs += "-v"
    }
    
    Write-Info "执行命令: go $($testArgs -join ' ')"
    & go @testArgs
    
    if ($LASTEXITCODE -eq 0) {
        Write-Success "✅ 单元测试通过"
        return $true
    } else {
        Write-Error "❌ 单元测试失败"
        return $false
    }
}

# 执行集成测试
function Run-IntegrationTests {
    Write-Info "=== 执行集成测试 ==="
    
    # 启动测试服务器
    Write-Info "启动测试服务器..."
    $serverProcess = Start-Process -FilePath "go" -ArgumentList @("run", "cmd/taskmanage/main.go", "--env=test") -PassThru -WindowStyle Hidden
    Start-Sleep -Seconds 5  # 等待服务器启动
    
    try {
        $testArgs = @(
            "test"
            "./test/integration/..."
            "-v"
            "-timeout=$($TestConfig.TestTimeout)"
        )
        
        Write-Info "执行命令: go $($testArgs -join ' ')"
        & go @testArgs
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "✅ 集成测试通过"
            return $true
        } else {
            Write-Error "❌ 集成测试失败"
            return $false
        }
    } finally {
        # 停止测试服务器
        if ($serverProcess -and !$serverProcess.HasExited) {
            Write-Info "停止测试服务器..."
            Stop-Process -Id $serverProcess.Id -Force
        }
    }
}

# 执行性能测试
function Run-PerformanceTests {
    Write-Info "=== 执行性能测试 ==="
    
    # 启动测试服务器
    Write-Info "启动测试服务器..."
    $serverProcess = Start-Process -FilePath "go" -ArgumentList @("run", "cmd/taskmanage/main.go", "--env=test") -PassThru -WindowStyle Hidden
    Start-Sleep -Seconds 5
    
    try {
        $testArgs = @(
            "test"
            "./test/performance/..."
            "-v"
            "-timeout=$($TestConfig.TestTimeout)"
            "-bench=."
            "-benchmem"
        )
        
        Write-Info "执行命令: go $($testArgs -join ' ')"
        & go @testArgs
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "✅ 性能测试完成"
            return $true
        } else {
            Write-Error "❌ 性能测试失败"
            return $false
        }
    } finally {
        if ($serverProcess -and !$serverProcess.HasExited) {
            Stop-Process -Id $serverProcess.Id -Force
        }
    }
}

# 生成测试报告
function Generate-TestReport {
    if (-not $Coverage) {
        Write-Warning "未启用覆盖率统计，跳过报告生成"
        return
    }
    
    Write-Info "=== 生成测试报告 ==="
    
    # 生成HTML覆盖率报告
    if (Test-Path $TestConfig.CoverageOutput) {
        Write-Info "生成HTML覆盖率报告..."
        go tool cover -html=$($TestConfig.CoverageOutput) -o $TestConfig.ReportOutput
        
        if ($LASTEXITCODE -eq 0) {
            Write-Success "✅ 测试报告已生成: $($TestConfig.ReportOutput)"
            
            # 显示覆盖率统计
            Write-Info "覆盖率统计:"
            go tool cover -func=$($TestConfig.CoverageOutput) | Select-Object -Last 1
        } else {
            Write-Error "❌ 报告生成失败"
        }
    }
}

# 执行Postman测试
function Run-PostmanTests {
    Write-Info "=== 执行Postman API测试 ==="
    
    # 检查Newman是否安装
    try {
        $newmanVersion = newman --version
        Write-Success "Newman版本: $newmanVersion"
    } catch {
        Write-Warning "Newman未安装，跳过Postman测试"
        Write-Info "安装命令: npm install -g newman"
        return $false
    }
    
    # 启动测试服务器
    $serverProcess = Start-Process -FilePath "go" -ArgumentList @("run", "cmd/taskmanage/main.go", "--env=test") -PassThru -WindowStyle Hidden
    Start-Sleep -Seconds 5
    
    try {
        $collectionPath = "test/postman/TaskManage_API_Tests.postman_collection.json"
        
        if (Test-Path $collectionPath) {
            Write-Info "执行Postman测试集合..."
            newman run $collectionPath --reporters cli,html --reporter-html-export postman_report.html
            
            if ($LASTEXITCODE -eq 0) {
                Write-Success "✅ Postman测试通过"
                return $true
            } else {
                Write-Error "❌ Postman测试失败"
                return $false
            }
        } else {
            Write-Warning "Postman测试集合不存在: $collectionPath"
            return $false
        }
    } finally {
        if ($serverProcess -and !$serverProcess.HasExited) {
            Stop-Process -Id $serverProcess.Id -Force
        }
    }
}

# 主执行逻辑
$testResults = @{}

try {
    # 启动测试数据库
    Start-TestDatabase
    
    switch ($TestType.ToLower()) {
        "unit" {
            $testResults["unit"] = Run-UnitTests
        }
        "integration" {
            $testResults["integration"] = Run-IntegrationTests
        }
        "performance" {
            $testResults["performance"] = Run-PerformanceTests
        }
        "postman" {
            $testResults["postman"] = Run-PostmanTests
        }
        "all" {
            Write-Info "执行完整测试套件..."
            $testResults["unit"] = Run-UnitTests
            $testResults["integration"] = Run-IntegrationTests
            $testResults["performance"] = Run-PerformanceTests
            $testResults["postman"] = Run-PostmanTests
        }
        default {
            Write-Error "未知的测试类型: $TestType"
            Write-Info "支持的类型: unit, integration, performance, postman, all"
            exit 1
        }
    }
    
    # 生成报告
    if ($Report) {
        Generate-TestReport
    }
    
} finally {
    # 清理资源
    Stop-TestDatabase
}

# 输出测试结果摘要
Write-Info "=== 测试结果摘要 ==="
$allPassed = $true

foreach ($test in $testResults.GetEnumerator()) {
    if ($test.Value) {
        Write-Success "✅ $($test.Key): 通过"
    } else {
        Write-Error "❌ $($test.Key): 失败"
        $allPassed = $false
    }
}

if ($allPassed) {
    Write-Success "🎉 所有测试通过！"
    exit 0
} else {
    Write-Error "💥 部分测试失败！"
    exit 1
}
