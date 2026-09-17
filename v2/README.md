# CipherForge Arena

本地启动：

```bash
php -S 127.0.0.1:8123 -t .
```

打开：<http://127.0.0.1:8123/v2/>

Docker 启动：

```bash
docker compose up --build
```

打开：<http://127.0.0.1:8080/v2/>

`v2/storage/` 只保存本地 SQLite、RSA 私钥和运行日志，已加入 `.gitignore`。当前仓库唯一入口是 CipherForge Arena 主线。

2.4 的高难链路以 JavaScript 加密逆向为核心：G03/K03 的密文由动态轮密钥、非线性表、CBC 链接、填充和 tag 共同生成；K03 提交后由 Go Runtime 独立解封，不能只伪造业务字段。G02 及王者级题目还会校验脚本完整性和运行时调试信号。

Docker 模式包含两个服务：PHP Web 负责页面、Session 和题目编排；Go Crypto Runtime 负责铂金/王者级的独立验证。未使用 Docker 时，如果没有设置 `ARENA_RUNTIME_URL`，高难题会回退到本地 PHP 验证器，便于开发调试。
