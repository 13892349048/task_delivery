-- 检查角色权限关系的SQL脚本

-- 1. 查看所有权限
SELECT 'All Permissions:' as section;
SELECT id, name, description, resource, action FROM permissions ORDER BY resource, action;

-- 2. 查看所有角色
SELECT 'All Roles:' as section;
SELECT id, name, description FROM roles ORDER BY name;

-- 3. 查看角色权限关联表
SELECT 'Role Permission Associations:' as section;
SELECT 
    r.name as role_name,
    p.name as permission_name,
    p.resource,
    p.action
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON p.id = rp.permission_id
ORDER BY r.name, p.resource, p.action;

-- 4. 统计每个角色的权限数量
SELECT 'Permission Count by Role:' as section;
SELECT 
    r.name as role_name,
    COUNT(rp.permission_id) as permission_count
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
GROUP BY r.id, r.name
ORDER BY permission_count DESC;

-- 5. 查看超级管理员的具体权限
SELECT 'Super Admin Permissions:' as section;
SELECT 
    p.name as permission_name,
    p.description,
    p.resource,
    p.action
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE r.name = 'super_admin'
ORDER BY p.resource, p.action;

-- 6. 查看用户角色关联
SELECT 'User Role Associations:' as section;
SELECT 
    u.username,
    u.real_name,
    u.role as user_role_field,
    r.name as assigned_role_name
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON r.id = ur.role_id
ORDER BY u.username;
