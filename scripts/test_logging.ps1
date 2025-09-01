# 测试日志输出的脚本
Write-Host "开始测试服务器日志输出..." -ForegroundColor Green

# 启动服务器（后台运行）
Write-Host "启动服务器..." -ForegroundColor Yellow
$serverProcess = Start-Process -FilePath ".\taskmanage.exe" -ArgumentList "-env", "development" -PassThru -NoNewWindow

# 等待服务器启动
Start-Sleep -Seconds 3

# 测试几个不同的端点
Write-Host "测试健康检查端点..." -ForegroundColor Yellow
try {
    $response1 = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Host "健康检查成功: $($response1 | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "健康检查失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "测试根路径..." -ForegroundColor Yellow
try {
    $response2 = Invoke-RestMethod -Uri "http://localhost:8080/" -Method GET
    Write-Host "根路径访问成功: $($response2 | ConvertTo-Json)" -ForegroundColor Green
} catch {
    Write-Host "根路径访问失败: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "测试不存在的路径（应该404）..." -ForegroundColor Yellow
try {
    $response3 = Invoke-RestMethod -Uri "http://localhost:8080/nonexistent" -Method GET
    Write-Host "意外成功: $($response3 | ConvertTo-Json)" -ForegroundColor Yellow
} catch {
    Write-Host "正确返回404: $($_.Exception.Message)" -ForegroundColor Green
}

# 等待一下让日志输出
Start-Sleep -Seconds 2

Write-Host "测试完成，请检查控制台日志输出" -ForegroundColor Green
Write-Host "按任意键停止服务器..." -ForegroundColor Yellow
Read-Host

# 停止服务器
if ($serverProcess -and !$serverProcess.HasExited) {
    Write-Host "正在停止服务器..." -ForegroundColor Yellow
    $serverProcess.Kill()
    $serverProcess.WaitForExit()
    Write-Host "服务器已停止" -ForegroundColor Green
}
