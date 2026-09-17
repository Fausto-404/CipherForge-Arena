# CipherForge Arena

密钥工场 · 前端协议对抗靶场

CipherForge Arena 是一个面向 JavaScript 逆向、前端加密分析和浏览器运行时协议研究的本地训练靶场。

项目不把难度建立在单纯的长代码和乱码上，而是逐步引入：

- 标准密码学与自定义编码；
- 请求规范化、HMAC、Session、nonce 和防重放；
- ESM 动态模块、控制流混淆和运行时完整性检测；
- Dedicated Worker、WASM、Service Worker 和 CacheStorage；
- 设备摘要、行为遥测、PoW、字节码 VM；
- 多阶段密钥状态机、密钥轮转和综合业务协议。

## 当前版本

主线版本：`2.3.0-dev`

共 6 个段位、18 道主线题：

| 段位 | 题目 |
| --- | --- |
| 青铜级 | B01 盐与摘要、B02 CBC 登录、B03 公钥拼图 |
| 白银级 | S01 规范化签名、S02 会话密钥、S03 模块迷宫 |
| 黄金级 | G01 扁平迷宫、G02 镜中调试器、G03 魔改轮函数 |
| 铂金级 | P01 Worker 密封舱、P02 WASM 黑盒、P03 网络影子 |
| 钻石级 | D01 设备合约、D02 轨迹凭证、D03 字节码金库 |
| 王者级 | K01 零售银行登录链、K02 风控 SDK、K03 王者终局 |

段位只用于导航和难度标识，所有段位默认可以直接进入。完成状态用于记录进度和评分，不作为访问门槛。

项目只保留 CipherForge Arena 新主线，不再维护旧版独立 AES/DES/RSA 入口；标准算法已经按难度重新编排进 18 道主线题。

## 本地启动

环境要求：PHP 8.2+，启用 `PDO_SQLite`、`OpenSSL` 和 `Session`。

```bash
php -S 127.0.0.1:8123 -t .
```

打开：<http://127.0.0.1:8123/v2/>

首次访问时会在 `v2/storage/` 创建本地 SQLite 数据库和服务端 RSA 密钥。该目录已被 `.gitignore` 排除。

## Docker

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

品牌资源位于 `v2/assets/brand/`：`cipherforge-arena-icon.svg` 为透明 Icon，`cipherforge-arena-background.svg` 为页面背景。

Docker 部署由 PHP Web 和 Go Crypto Runtime 两个服务组成。PHP 负责页面与题目编排，铂金/王者级的关键验证在独立 Go 服务中完成。

## 设计原则

- 服务端权威判定，不能通过修改前端状态直接通关。
- 每道主线题使用随机 Session 上下文和随机业务载荷。
- nonce、时间窗、AAD、请求体语义和状态顺序由服务端校验。
- 反调试只作为分析面的一部分，不制造无法恢复的浏览器死循环。
- 题目可以通过动态分析，也可以通过理解协议后独立复现。
- 标准密码学题作为新主线的一部分提供，不保留平行 Legacy 入口。

## 研究与使用边界

本项目只用于本地学习、授权测试和前端协议研究。王者级题目模拟大型业务常见的协议复杂度，不使用真实银行、支付平台或其他机构的私有代码、密钥和数据。

## License

Apache-2.0，详见 [LICENSE](LICENSE)。
