# CipherForge Arena

密钥工场 · 前端协议对抗靶场

CipherForge Arena 是一个面向 JavaScript 逆向、前端加密分析和浏览器运行时协议研究的本地训练靶场。

项目包含：

- 标准密码学与自定义编码；
- 请求规范化、HMAC、Session、nonce 和防重放；
- ESM 动态模块、控制流混淆和运行时完整性检测；
- Dedicated Worker、WASM、Service Worker 和 CacheStorage；
- 设备摘要、行为遥测、PoW、字节码 VM；
- 多阶段密钥状态机、密钥轮转和自定义密文 envelope。

## 题型概览

共 6 个段位、18 道主线题：

| 段位 | 题目 |
| --- | --- |
| 青铜级 | B01 盐与摘要、B02 CBC 登录、B03 公钥拼图 |
| 白银级 | S01 规范化签名、S02 会话密钥、S03 模块迷宫 |
| 黄金级 | G01 扁平迷宫、G02 镜中调试器、G03 魔改轮函数 |
| 铂金级 | P01 Worker 密封舱、P02 WASM 黑盒、P03 网络影子 |
| 钻石级 | D01 设备合约、D02 轨迹凭证、D03 字节码金库 |
| 王者级 | K01 零售银行登录链、K02 风控 SDK、K03 王者终局 |


## 本地启动

环境要求：PHP 8.2+，启用 `PDO_SQLite`、`OpenSSL` 和 `Session`。

```bash
php -S 127.0.0.1:8123 -t .
```

打开：<http://127.0.0.1:8123/v2/>

首次访问时会在 `v2/storage/` 创建本地 SQLite 数据库和服务端 RSA 密钥。该目录已被 `.gitignore` 排除。

## Docker部署（推荐）

```bash
docker compose up --build
```

打开：<http://127.0.0.1:8080/v2/>

## 目录

```text
.
├── index.php                # 根路径重定向到主线
├── v2/
│   ├── api.php             # 题目列表、上下文、提交验证
│   ├── bootstrap.php       # SQLite、Session、验证器
│   ├── data/challenges.php # 18 道题的公开定义
│   ├── assets/             # 页面、运行时脚本、品牌资源
│   └── storage/            # 本地运行数据，不提交
├── ADVANCED_LAB_BLUEPRINT.md
├── services/crypto-runtime/  # Go 高难协议验证运行时
├── Dockerfile
└── docker-compose.yml
```



## License

Apache-2.0，详见 [LICENSE](LICENSE)。
