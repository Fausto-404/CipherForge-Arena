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

`v2/storage/` 只保存本地 SQLite、RSA 私钥和运行日志，已加入 `.gitignore`。旧版 8 题仍然从根目录入口提供；新主线从 `v2/` 进入。

