# Hash 模块

密码哈希，两个函数搞定，算法固定为 argon2id（PHC 密码哈希竞赛冠军，当前业界首选）。
哈希结果为 PHC 自描述格式（内含算法、参数、盐），原样存库即可。

# 使用方法

## 注册：哈希后存库

```go
encoded, err := service.HashPassword("原始密码") // 返回编码串，直接存 user.Password
```

## 登录：校验

```go
ok := service.VerifyPassword("用户输入的密码", user.Password)
```

# 函数

HashPassword(password) 哈希密码，argon2id 参数 64MB / t=1 / p=4（RFC 9106 推荐档）
VerifyPassword(password, encoded) 校验密码，恒定时间比较，格式错误一律返回 false

# 编码串格式

$argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>

# 注意事项

- 编码串原样存库即可，盐已内含，不要自己再拼盐
- 用户输入的密码进 VerifyPassword 前不需要预处理（模块不做 trim/大小写转换）
