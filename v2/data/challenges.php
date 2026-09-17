<?php

return [
    'bronze' => [
        ['id'=>'B01','title'=>'盐与摘要','scene'=>'会员登录','summary'=>'把一条普通登录请求还原成服务端认可的摘要请求。','skills'=>'归一化 / UTF-8 / SHA-256','score'=>100,'inputs'=>['username','password']],
        ['id'=>'B02','title'=>'CBC 登录','scene'=>'商户后台','summary'=>'处理动态 IV、块填充和密文封装，完成一次登录。','skills'=>'AES-CBC / IV / Base64','score'=>100,'inputs'=>['account','amount']],
        ['id'=>'B03','title'=>'公钥拼图','scene'=>'支付密码','summary'=>'从页面运行时配置中重组公钥，生成服务端可以解开的密文。','skills'=>'RSA-OAEP / PEM / WebCrypto','score'=>100,'inputs'=>['account','password']],
    ],
    'silver' => [
        ['id'=>'S01','title'=>'规范化签名','scene'=>'订单查询','summary'=>'还原请求规范化规则并构造一条新的签名请求。','skills'=>'Canonical request / HMAC / nonce','score'=>180,'inputs'=>['order','amount']],
        ['id'=>'S02','title'=>'会话密钥','scene'=>'身份预检','summary'=>'完成 bootstrap、派生会话密钥和一次性验证。','skills'=>'HKDF-like derivation / AES-GCM / replay','score'=>180,'inputs'=>['account','code']],
        ['id'=>'S03','title'=>'模块迷宫','scene'=>'报价系统','summary'=>'跨越动态模块边界，找到真正的序列化和签名路径。','skills'=>'ESM / dynamic import / request builder','score'=>180,'inputs'=>['symbol','quantity']],
    ],
    'gold' => [
        ['id'=>'G01','title'=>'扁平迷宫','scene'=>'积分兑换','summary'=>'从诱饵算法和扁平控制流中分离真实请求路径。','skills'=>'control-flow flattening / decoy','score'=>300,'inputs'=>['member','points']],
        ['id'=>'G02','title'=>'镜中调试器','scene'=>'敏感操作','summary'=>'理解多信号运行时检测，同时保持业务链正常完成。','skills'=>'integrity / timing / telemetry','score'=>300,'inputs'=>['account','action']],
        ['id'=>'G03','title'=>'魔改轮函数','scene'=>'设备注册码','summary'=>'通过差分输入输出还原非标准轮函数。','skills'=>'custom rounds / endian / differential','score'=>300,'inputs'=>['device','serial']],
    ],
    'platinum' => [
        ['id'=>'P01','title'=>'Worker 密封舱','scene'=>'交易口令','summary'=>'追踪主线程与 Worker 之间的 envelope 和序号。','skills'=>'Dedicated Worker / transferable / async','score'=>450,'inputs'=>['account','amount']],
        ['id'=>'P02','title'=>'WASM 黑盒','scene'=>'行情订阅','summary'=>'识别 JavaScript wrapper 与线性内存之间的边界。','skills'=>'WASM / linear memory / checksum','score'=>450,'inputs'=>['symbol','tier']],
        ['id'=>'P03','title'=>'网络影子','scene'=>'企业门户','summary'=>'发现 Service Worker 改写后的真实请求。','skills'=>'Service Worker / CacheStorage / fetch','score'=>450,'inputs'=>['account','resource']],
    ],
    'diamond' => [
        ['id'=>'D01','title'=>'设备合约','scene'=>'高价值登录','summary'=>'生成内部一致、可重复验证的设备摘要。','skills'=>'Canvas / WebGL / Audio / consistency','score'=>650,'inputs'=>['account','deviceLabel']],
        ['id'=>'D02','title'=>'轨迹凭证','scene'=>'抢购资格','summary'=>'重建事件顺序、时间分布和轻量 PoW。','skills'=>'behavior telemetry / PoW / scoring','score'=>650,'inputs'=>['account','sku']],
        ['id'=>'D03','title'=>'字节码金库','scene'=>'授权文件','summary'=>'解释自定义栈式字节码并生成兼容结果。','skills'=>'VM / opcode / constant pool','score'=>650,'inputs'=>['account','scope']],
    ],
    'king' => [
        ['id'=>'K01','title'=>'零售银行登录链','scene'=>'金融登录演练','summary'=>'完成 bootstrap、密钥协商、登录和登录后查询四阶段。','skills'=>'ECDH-like handshake / Worker / ratchet','score'=>1000,'inputs'=>['account','password']],
        ['id'=>'K02','title'=>'风控 SDK','scene'=>'支付确认','summary'=>'连续完成设备证明、风控预检和密钥轮转。','skills'=>'risk score / CBOR-like envelope / ratchet','score'=>1000,'inputs'=>['account','order','amount']],
        ['id'=>'K03','title'=>'王者终局：多域状态机','scene'=>'跨域高风险业务','summary'=>'还原多阶段状态机和业务签名，随机载荷下仍需稳定工作。','skills'=>'state machine / VM / multi-stage proof','score'=>1000,'inputs'=>['account','operation']],
    ],
];

