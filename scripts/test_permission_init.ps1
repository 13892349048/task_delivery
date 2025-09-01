# 测试权限初始化脚本
Write-Host "=== 测试权限初始化 ===" -ForegroundColor Green

# 启动应用程序（后台运行）
Write-Host "启动应用程序..." -ForegroundColor Yellow
$process = Start-Process -FilePath ".\taskmanage.exe" -PassThru -WindowStyle Hidden

# 等待应用程序启动
Start-Sleep -Seconds 3

try {
    # 测试超级管理员登录
    Write-Host "测试超级管理员登录..." -ForegroundColor Yellow
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -ContentType "application/json" -Body (@{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json)

    if ($loginResponse.access_token) {
        Write-Host "✓ 超级管理员登录成功" -ForegroundColor Green
        $token = $loginResponse.access_token
        
        # 获取用户信息
        Write-Host "获取用户信息..." -ForegroundColor Yellow
        $headers = @{ "Authorization" = "Bearer $token" }
        $userInfo = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users/me" -Method GET -Headers $headers
        
        Write-Host "用户信息:" -ForegroundColor Cyan
        Write-Host "  用户名: $($userInfo.username)" -ForegroundColor White
        Write-Host "  角色: $($userInfo.role)" -ForegroundColor White
        Write-Host "  状态: $($userInfo.status)" -ForegroundColor White
        
        # 测试权限 - 尝试创建用户（需要user:create权限）
        Write-Host "测试用户创建权限..." -ForegroundColor Yellow
        try {
            $createUserResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method POST -Headers $headers -ContentType "application/json" -Body (@{
                username = "testuser"
                email = "test@example.com"
                password = "password123"
                real_name = "测试用户"
            } | ConvertTo-Json)
            
            Write-Host "✓ 用户创建权限测试通过" -ForegroundColor Green
        } catch {
            Write-Host "✗ 用户创建权限测试失败: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        # 测试权限 - 尝试获取用户列表（需要user:read权限）
        Write-Host "测试用户查看权限..." -ForegroundColor Yellow
        try {
            $usersResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/users" -Method GET -Headers $headers
            Write-Host "✓ 用户查看权限测试通过，找到 $($usersResponse.pagination.total) 个用户" -ForegroundColor Green
        } catch {
            Write-Host "✗ 用户查看权限测试失败: $($_.Exception.Message)" -ForegroundColor Red
        }
        
    } else {
        Write-Host "✗ 超级管理员登录失败" -ForegroundColor Red
    }
    
} catch {
    Write-Host "✗ 测试过程中出现错误: $($_.Exception.Message)" -ForegroundColor Red
} finally {
    # 停止应用程序
    Write-Host "停止应用程序..." -ForegroundColor Yellow
    Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue
}

Write-Host "=== 测试完成 ===" -ForegroundColor Green
