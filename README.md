# CipherForge Arena

**密钥工场 · 前端协议对抗靶场**

CipherForge Arena 是一个用于学习 JavaScript 逆向、前端密码分析与浏览器运行时协议研究的本地训练靶场。项目提供 18 道由浅入深的挑战，覆盖从摘要与签名，到浏览器 API、运行时检测和多阶段加密协议的分析练习。

**适合练习：** Web Crypto 调用分析、请求签名还原、浏览器反调试、Worker 与 Service Worker 追踪、自定义密文分析，以及多阶段业务协议的状态还原。
<img width="2998" height="1640" alt="image" src="https://github.com/user-attachments/assets/9501f991-fa95-4482-b57b-f4878620d653" />


## 加密与签名

以下机制均可在当前代码和题目请求中观察到：

| 技术 | 具体练习内容 | 对应题目 |
| --- | --- | --- |
| SHA-256 与数据归一化 | 追踪用户名修剪、大小写转换、UTF-8 编码及摘要输入的拼接顺序；比较原始字段与请求中的摘要值。 | B01 |
| AES-128-CBC | 还原从上下文派生的密钥、随机 16 字节 IV、块填充和 `IV + 密文` 的 Base64 封装。 | B02 |
| RSA-OAEP | 从运行时上下文读取 PEM 公钥，导入 Web Crypto，分析 RSA-OAEP 加密及服务端私钥解密；当前题目使用 SHA-1 作为 OAEP 哈希。 | B03 |
| HMAC-SHA-256 | 还原字段顺序、请求体摘要、时间戳、nonce 和 Canonical Request；处理 Base64URL 与十六进制两种签名表示。 | S01、S03、P01、P02、P03、K01、K02 |
| AES-128-GCM | 分析 12 字节 IV、附加认证数据（AAD）、密文与 16 字节认证标签的分离传输，以及一次性 nonce 校验。 | S02 |
| 自定义密文 `CFX2` | 从模块中还原轮密钥、非线性表、轮函数、分组链接、填充、长度帧和 12 字节 SHA-256 标签；服务端会独立解封并校验明文。 | G03、K03 |

题目中的 `seed`、`nonce` 由当前 Session 的上下文生成。复制一条旧请求通常无法直接用于新的上下文；分析时需要同时记录密钥材料、字段顺序和编码方式。

## 反调试与完整性检测

黄金级部分题目和王者级题目会收集浏览器运行时信号，并把信号与当前 nonce 一起提交校验。可以练习定位检测入口、理解每个信号的来源，以及分析检测结果如何参与签名和服务端判定。

| 检测手法 | 实际检查内容 | 分析切入点 |
| --- | --- | --- |
| `debugger` 时间探针 | 多次执行 `debugger`，用 `performance.now()` 统计暂停耗时及波动。 | 跟踪断点前后的时间差与阈值。 |
| 原生函数完整性 | 检查 `console.log/warn/error`、`crypto.subtle.digest`、`performance.now` 和 `Function.prototype.toString` 的原生特征。 | 分析 Hook、代理或覆写会怎样改变检测结果。 |
| 调用栈检查 | 检查 `Error().stack` 中是否出现 `eval`。 | 观察动态执行环境对堆栈的影响。 |
| 窗口与页面状态 | 比较窗口内外尺寸，读取 `document.visibilityState` 和焦点状态。 | 观察 DevTools 布局和页面失焦对评分的影响。 |
| 脚本哈希校验 | 在浏览器端计算 `app.js` 的 SHA-256，服务端对照磁盘文件哈希。 | 区分请求字段被改写与脚本内容被改写。 |
| 上下文绑定证明 | 基于 nonce 生成探针令牌，再对运行时信号计算 HMAC。 | 还原信号序列化顺序、会话绑定和校验链。 |

对应题目：**G01、G02、G03、K01、K02、K03**。检测逻辑位于 [`runtime-guard.js`](v2/assets/challenge-modules/runtime-guard.js)，PHP 与 Go 校验分别位于 [`bootstrap.php`](v2/bootstrap.php) 和 [`main.go`](services/crypto-runtime/main.go)。

## 对抗与浏览器机制

| 场景 | 可以练习的分析动作 | 对应题目 |
| --- | --- | --- |
| 动态模块 | 跟踪 `import()` 加载的 ESM，定位签名函数与主线程请求构造之间的数据流。 | S03 |
| Dedicated Worker | 跟踪运行时创建的 Blob Worker、`postMessage`、异步返回值及 Worker 内的 HMAC 运算。 | P01 |
| Service Worker | 观察注册、接管和 `/v2/api.php` 请求拦截；对比页面发起的请求与 Worker 中的 `fetch`。 | P03 |
| 设备摘要 | 记录 UA、语言、时区、屏幕尺寸、Canvas 输出等字段，分析字段顺序与 SHA-256 摘要的一致性。 | D01 |
| 行为与计算证明 | 分析事件数组、时间戳、HMAC 证明及 SHA-256 前缀条件的轻量 PoW。 | D02 |
| 简化字节码 VM | 阅读 `push`、`xor`、`add` 操作码及栈行为，复现服务端期望的结果。 | D03 |
| 多阶段状态 | 还原 `bootstrap → handshake → login → query` 或 `arena → auth → risk → business` 的状态字段、签名串与密文。 | K01、K03 |

## 挑战路线

| 段位 | 题目 | 主要练习点 |
| --- | --- | --- |
| 青铜级 | B01 盐与摘要 · B02 CBC 登录 · B03 公钥拼图 | SHA-256、AES-CBC、RSA-OAEP |
| 白银级 | S01 规范化签名 · S02 会话密钥 · S03 模块迷宫 | HMAC、AES-GCM、nonce、防重放、动态模块 |
| 黄金级 | G01 扁平迷宫 · G02 镜中调试器 · G03 魔改轮函数 | 运行时检测、脚本完整性、自定义分组密文 |
| 铂金级 | P01 Worker 密封舱 · P02 WASM 黑盒 · P03 网络影子 | Worker 消息链、签名封装、Service Worker 请求拦截 |
| 钻石级 | D01 设备合约 · D02 轨迹凭证 · D03 字节码金库 | 设备字段一致性、行为证明、轻量 PoW、栈式 VM |
| 王者级 | K01 零售银行登录链 · K02 风控 SDK · K03 王者终局 | 多阶段签名、风险评分、状态链、自定义密文解封 |

## 快速开始

### Docker Compose（推荐）

准备 Docker Engine 或 Docker Desktop 与 Docker Compose v2。克隆或下载项目后，在项目根目录执行：

```bash
docker compose up --build -d
```

浏览器打开 **<http://localhost:8080/v2/>**。首次构建会下载 PHP 和 Go 基础镜像。Compose 启动 PHP Web 与 Go Crypto Runtime；后者负责 P01–P03 和 K01–K03 的独立校验。

```bash
docker compose ps          # 查看两个服务的状态
docker compose logs -f     # 查看日志
docker compose down        # 停止服务
```

当前 Compose 配置将 Web 服务发布到宿主机的 `8080` 端口。若只允许本机访问，可将 `docker-compose.yml` 中的端口映射改为 `127.0.0.1:8080:8080`。

### PHP 本地开发

需要 PHP 8.2+，并启用 `PDO_SQLite`、`OpenSSL` 和 `Session`：

```bash
php -S 127.0.0.1:8123 -t .
```

浏览器打开 **<http://127.0.0.1:8123/v2/>**。未设置 `ARENA_RUNTIME_URL` 时，相关题目使用 PHP 本地验证器。

## 练习方式

1. 在题目界面执行业务请求，先在 **Network** 查看请求体、响应和失败原因。
2. 在 **Sources** 中定位请求生成函数，沿着模块、Worker 或 Service Worker 边界记录输入与中间值。
3. 独立复现编码、签名或密文生成过程，再用当前 Session 的上下文提交请求。
4. 对反调试题记录运行时信号与服务端返回结果，逐项分析校验条件。

页面会展示它生成的请求样本，适合先观察协议，再逐步减少对页面实现的依赖。


## 项目结构

```text
.
├── index.php                       # 根路径跳转到 /v2/
├── v2/
│   ├── api.php                     # 题目列表、上下文、提交接口
│   ├── bootstrap.php               # Session、SQLite、PHP 验证逻辑
│   ├── data/challenges.php         # 18 道题的定义
│   ├── assets/                     # 页面、脚本和挑战模块
│   └── storage/                    # 本地数据库与 RSA 私钥（不提交）
├── services/crypto-runtime/        # Go 独立验证服务
├── ADVANCED_LAB_BLUEPRINT.md       # 后续设计蓝图
├── Dockerfile
└── docker-compose.yml
```

运行数据保存在 `v2/storage/`，包含 SQLite 数据库和服务端 RSA 私钥。数据库、私钥和日志文件已由 `.gitignore` 排除；请勿在分享仓库或提交 Issue 时附带这些文件。

## 贡献与许可

欢迎通过 GitHub Issues 反馈题目行为、文档错误或兼容性问题，并通过 Pull Request 提交改进。描述问题时请注明题目编号、运行方式与复现步骤。

