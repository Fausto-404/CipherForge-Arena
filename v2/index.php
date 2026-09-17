<?php
declare(strict_types=1);
require __DIR__ . '/bootstrap.php';
?>
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="color-scheme" content="light">
  <title>CipherForge Arena / advanced protocol lab</title>
  <link rel="stylesheet" href="assets/style.css?v=<?= rawurlencode(ARENA_VERSION) ?>">
</head>
<body>
  <header class="topbar">
    <div class="brand"><span class="brand-mark">CF</span><span>CipherForge Arena</span><small>advanced protocol lab</small></div>
    <div class="top-meta"><span id="session-label">SESSION —</span><span class="status-dot"></span><span>本地训练环境</span></div>
  </header>
  <main class="layout">
    <aside class="rail" id="rank-rail"></aside>
    <section class="workspace">
      <div class="intro-row">
        <div>
          <p class="eyebrow">FRONTEND SECURITY LAB / 2.3 DEV</p>
          <h1>逆向训练档案</h1>
          <p class="lede">从可见的密码学到多阶段浏览器协议。每个任务都由真实请求、运行时状态和服务端校验共同组成。</p>
        </div>
        <div class="progress-box"><span>已完成</span><strong id="progress-count">0 / 18</strong><div class="progress"><i id="progress-bar"></i></div></div>
      </div>
      <div id="challenge-board" class="challenge-board"></div>
    </section>
  </main>
  <div id="drawer" class="drawer" aria-hidden="true">
    <div class="drawer-inner">
      <button id="drawer-close" class="drawer-close" type="button">关闭</button>
      <div id="drawer-content"></div>
    </div>
  </div>
  <script>window.ARENA_API = 'api.php';</script>
  <script type="module" src="assets/app.js?v=<?= rawurlencode(ARENA_VERSION) ?>"></script>
</body>
</html>
