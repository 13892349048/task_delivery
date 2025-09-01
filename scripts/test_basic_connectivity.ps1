# 基础连接测试脚本
$baseUrl = "http://localhost:8081"
$headers = @{
    "Content-Type" = "application/json"
}

Write-Host "=== 基础连接测试 ===" -ForegroundColor Green

# 测试无需认证的端点
$endpoints = @(
    @{ url = "/"; name = "根路径" },
    @{ url = "/health"; name = "健康检查" },
    @{ url = "/health/ready"; name = "就绪检查" },
    @{ url = "/health/live"; name = "存活检查" }
)

foreach ($endpoint in $endpoints) {
    Write-Host "测试 $($endpoint.name): $($endpoint.url)" -ForegroundColor Yellow
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl$($endpoint.url)" -Method GET -Headers $headers -TimeoutSec 10
        Write-Host "✓ 成功 - 状态: $($response.status)" -ForegroundColor Green
    } catch {
        Write-Host "✗ 失败: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.ErrorDetails.Message) {
            Write-Host "  详情: $($_.ErrorDetails.Message)" -ForegroundColor Red
        }
    }
    Start-Sleep -Milliseconds 500
}

# 测试认证端点
Write-Host "`n测试认证端点..." -ForegroundColor Yellow
$loginRequest = @{
    username = "test"
    password = "test123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/v1/auth/login" -Method POST -Body $loginRequest -Headers $headers -TimeoutSec 10
    Write-Host "✓ 登录端点响应正常" -ForegroundColor Green
} catch {
    Write-Host "✗ 登录端点失败: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.ErrorDetails.Message) {
        Write-Host "  详情: $($_.ErrorDetails.Message)" -ForegroundColor Red
    }
}

Write-Host "`n=== 测试完成 ===" -ForegroundColor Green
