<?php

declare(strict_types=1);

const ARENA_VERSION = '2.3.0-dev';
const ARENA_ROOT = __DIR__;

function arena_start(): void
{
    if (session_status() !== PHP_SESSION_ACTIVE) {
        session_name('encrypt_arena');
        session_start();
    }
    $dir = ARENA_ROOT . '/storage';
    if (!is_dir($dir)) {
        mkdir($dir, 0775, true);
    }
    $pdo = arena_db();
    $pdo->exec('CREATE TABLE IF NOT EXISTS attempts (id INTEGER PRIMARY KEY AUTOINCREMENT, sid TEXT, challenge_id TEXT, ok INTEGER, detail TEXT, created_at INTEGER)');
    $pdo->exec('CREATE TABLE IF NOT EXISTS solves (sid TEXT, challenge_id TEXT, score INTEGER, solved_at INTEGER, PRIMARY KEY (sid, challenge_id))');
}

function arena_db(): PDO
{
    static $pdo;
    if ($pdo instanceof PDO) {
        return $pdo;
    }
    $pdo = new PDO('sqlite:' . ARENA_ROOT . '/storage/arena.sqlite');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    $pdo->setAttribute(PDO::ATTR_DEFAULT_FETCH_MODE, PDO::FETCH_ASSOC);
    return $pdo;
}

function arena_json(mixed $value, int $status = 200): never
{
    http_response_code($status);
    header('Content-Type: application/json; charset=utf-8');
    header('Cache-Control: no-store');
    echo json_encode($value, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
    exit;
}

function arena_sid(): string
{
    return session_id();
}

function arena_challenges(): array
{
    static $items;
    return $items ??= require ARENA_ROOT . '/data/challenges.php';
}

function arena_find(string $id): ?array
{
    foreach (arena_challenges() as $rank => $items) {
        foreach ($items as $item) {
            if ($item['id'] === $id) {
                $item['rank'] = $rank;
                return $item;
            }
        }
    }
    return null;
}

function arena_context(string $id): array
{
    if (!isset($_SESSION['challenge_context'][$id])) {
        $_SESSION['challenge_context'][$id] = [
            'seed' => bin2hex(random_bytes(18)),
            'nonce' => bin2hex(random_bytes(12)),
            'issued' => time(),
            'counter' => 0,
        ];
    }
    return $_SESSION['challenge_context'][$id];
}

function arena_public_context(string $id): array
{
    $ctx = arena_context($id);
    $public = [
        'challenge' => $id,
        'seed' => $ctx['seed'],
        'nonce' => $ctx['nonce'],
        'issued' => $ctx['issued'],
        'serverTime' => time(),
    ];
    if ($id === 'B03') {
        $keyPath = ARENA_ROOT . '/storage/rsa_private.pem';
        if (!is_file($keyPath)) {
            $key = openssl_pkey_new(['private_key_bits' => 2048, 'private_key_type' => OPENSSL_KEYTYPE_RSA]);
            openssl_pkey_export($key, $private);
            file_put_contents($keyPath, $private);
        }
        $details = openssl_pkey_get_details(openssl_pkey_get_private(file_get_contents($keyPath)));
        $public['publicKey'] = $details['key'];
    }
    return $public;
}

function arena_nonce_available(string $id, string $nonce, array $ctx): bool
{
    if ($nonce === '' || !hash_equals((string)$ctx['nonce'], $nonce)) {
        return false;
    }
    return !isset($_SESSION['consumed_nonces'][$id][$nonce]);
}

function arena_consume_nonce(string $id, string $nonce): void
{
    $_SESSION['consumed_nonces'][$id][$nonce] = time();
}

function arena_b64(string $value): string
{
    return base64_encode($value);
}

function arena_unb64(string $value): string|false
{
    return base64_decode($value, true);
}

function arena_hex_hmac(string $algo, string $data, string $key): string
{
    return hash_hmac($algo, $data, $key);
}

function arena_mix(string $text, string $seed): string
{
    $state = array_values(unpack('C*', hash('sha256', $seed, true)));
    $bytes = array_values(unpack('C*', $text));
    foreach ($bytes as $i => $byte) {
        $v = $byte ^ $state[$i % count($state)] ^ (($i * 17 + 29) & 255);
        $bytes[$i] = (($v << 3) | ($v >> 5)) & 255;
    }
    return pack('C*', ...$bytes);
}

function arena_device_hash(array $data): string
{
    $fields = ['ua', 'lang', 'tz', 'width', 'height', 'pixel', 'canvas', 'webgl', 'audio'];
    $parts = [];
    foreach ($fields as $field) {
        $parts[] = $field . '=' . (string)($data[$field] ?? '');
    }
    return hash('sha256', implode('&', $parts));
}

function arena_verify(string $id, array $payload, array $ctx): array
{
    $seed = $ctx['seed'];
    $message = '';
    switch ($id) {
        case 'B01':
            $username = trim((string)($payload['username'] ?? ''));
            $digest = (string)($payload['digest'] ?? '');
            $message = strtolower($username) . ':' . $seed . ':' . (string)($payload['passwordProof'] ?? '');
            $expected = hash('sha256', $message);
            return [hash_equals($expected, $digest), 'sha256 normalized credential proof'];
        case 'B02':
            $raw = arena_unb64((string)($payload['ciphertext'] ?? ''));
            if ($raw === false || strlen($raw) < 17) return [false, 'invalid CBC envelope'];
            $iv = substr($raw, 0, 16);
            $key = substr(hash('sha256', 'B02|' . $seed, true), 0, 16);
            $plain = openssl_decrypt(substr($raw, 16), 'AES-128-CBC', $key, OPENSSL_RAW_DATA, $iv);
            return [$plain !== false && is_string($plain) && str_contains($plain, 'account'), 'aes cbc envelope'];
        case 'B03':
            $cipher = arena_unb64((string)($payload['ciphertext'] ?? ''));
            $privatePath = ARENA_ROOT . '/storage/rsa_private.pem';
            if ($cipher === false || !is_file($privatePath)) return [false, 'invalid RSA envelope'];
            $plain = '';
            $ok = openssl_private_decrypt($cipher, $plain, file_get_contents($privatePath), OPENSSL_PKCS1_OAEP_PADDING);
            return [$ok && str_contains($plain, 'password'), 'rsa oaep envelope'];
        case 'S01':
            $body = json_encode($payload['body'] ?? [], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
            $timestamp = (string)($payload['timestamp'] ?? '');
            $nonce = (string)($payload['nonce'] ?? '');
            $canonical = "POST\n/api/order\n" . hash('sha256', $body ?: '') . "\n" . $timestamp . "\n" . $nonce;
            $expected = arena_hex_hmac('sha256', $canonical, $seed);
            $timeOk = abs(time() - (int)$timestamp) <= 30;
            $signatureOk = hash_equals($expected, (string)($payload['signature'] ?? ''));
            $nonceOk = arena_nonce_available($id, $nonce, $ctx);
            $valid = $timeOk && $signatureOk && $nonceOk;
            if ($valid) arena_consume_nonce($id, $nonce);
            return [$valid, !$timeOk ? 'request expired' : (!$nonceOk ? 'nonce replay' : ($signatureOk ? 'canonical hmac request' : 'signature mismatch'))];
        case 'S02':
            $iv = arena_unb64((string)($payload['iv'] ?? ''));
            $cipher = arena_unb64((string)($payload['ciphertext'] ?? ''));
            $tag = arena_unb64((string)($payload['tag'] ?? ''));
            $key = substr(hash('sha256', 'S02|' . $seed, true), 0, 16);
            $plain = ($iv !== false && $cipher !== false && $tag !== false) ? openssl_decrypt($cipher, 'aes-128-gcm', $key, OPENSSL_RAW_DATA, $iv, $tag, $id . '|' . $ctx['nonce']) : false;
            $valid = $plain !== false && str_contains((string)$plain, 'code') && arena_nonce_available($id, $ctx['nonce'], $ctx);
            if ($valid) arena_consume_nonce($id, $ctx['nonce']);
            return [$valid, $valid ? 'session aes-gcm proof' : ($plain !== false ? 'nonce replay' : 'session envelope invalid')];
        case 'S03':
            return [hash_equals(hash_hmac('sha256', (string)($payload['symbol'] ?? '') . '|' . (string)($payload['quantity'] ?? ''), $seed), (string)($payload['signature'] ?? '')), 'dynamic module signature'];
        case 'G01':
            return [hash_equals(hash_hmac('sha256', (string)($payload['member'] ?? '') . ':' . (string)($payload['points'] ?? ''), $seed), (string)($payload['proof'] ?? '')), 'flattened flow proof'];
        case 'G02':
            $signals = $payload['signals'] ?? [];
            $score = (int)($signals['score'] ?? 0);
            $proof = hash_hmac('sha256', json_encode($signals, JSON_UNESCAPED_SLASHES), $seed);
            return [$score >= 60 && hash_equals($proof, (string)($payload['proof'] ?? '')), 'runtime integrity telemetry'];
        case 'G03':
            $expected = bin2hex(arena_mix((string)($payload['device'] ?? '') . '|' . (string)($payload['serial'] ?? ''), $seed));
            $actual = (string)($payload['token'] ?? '');
            return [hash_equals($expected, $actual), 'custom round function'];
        case 'P01':
            return [hash_equals(hash_hmac('sha256', (string)($payload['account'] ?? '') . '|' . (string)($payload['amount'] ?? ''), $seed), (string)($payload['workerProof'] ?? '')), 'worker envelope'];
        case 'P02':
            $expected = rtrim(strtr(base64_encode(hash_hmac('sha256', (string)($payload['symbol'] ?? '') . '|' . (string)($payload['tier'] ?? ''), $seed, true)), '+/', '-_'), '=');
            return [hash_equals($expected, (string)($payload['token'] ?? '')), 'wasm wrapper token'];
        case 'P03':
            return [$payload['viaWorker'] === true && hash_equals(hash_hmac('sha256', (string)($payload['account'] ?? '') . '|' . (string)($payload['resource'] ?? ''), $seed), (string)($payload['shadowProof'] ?? '')), 'service worker rewritten request'];
        case 'D01':
            $device = $payload['device'] ?? [];
            $hash = arena_device_hash($device);
            return [hash_equals($hash, (string)($payload['deviceHash'] ?? '')) && strlen((string)($device['canvas'] ?? '')) >= 8, 'device consistency contract'];
        case 'D02':
            $events = $payload['events'] ?? [];
            $proof = hash_hmac('sha256', json_encode($events, JSON_UNESCAPED_SLASHES), $seed);
            $pow = hash('sha256', $seed . ':' . (string)($payload['pow'] ?? ''));
            return [count($events) >= 8 && str_starts_with($pow, '000'), hash_equals($proof, (string)($payload['telemetryProof'] ?? '')) ? 'behavior score' : 'telemetry mismatch'];
        case 'D03':
            $program = $payload['program'] ?? [];
            $stack = [];
            foreach ($program as $op) {
                if (($op['op'] ?? '') === 'push') $stack[] = (int)$op['value'];
                if (($op['op'] ?? '') === 'xor' && count($stack) >= 2) { $b = array_pop($stack); $a = array_pop($stack); $stack[] = $a ^ $b; }
                if (($op['op'] ?? '') === 'add' && count($stack) >= 2) { $b = array_pop($stack); $a = array_pop($stack); $stack[] = ($a + $b) & 255; }
            }
            $expected = ((int)($payload['scopeCode'] ?? 0) ^ hexdec(substr($seed, 0, 2))) & 255;
            return [isset($stack[0]) && $stack[0] === $expected, 'bytecode vm result'];
        case 'K01':
            $proof = hash_hmac('sha256', 'bootstrap|' . (string)($payload['account'] ?? '') . '|login|' . (string)($payload['password'] ?? '') . '|query', $seed);
            return [hash_equals($proof, (string)($payload['proof'] ?? '')) && ($payload['stages'] ?? []) === ['bootstrap','handshake','login','query'], 'four-stage login chain'];
        case 'K02':
            $body = (string)($payload['account'] ?? '') . '|' . (string)($payload['order'] ?? '') . '|' . (string)($payload['amount'] ?? '');
            $expected = hash_hmac('sha256', 'device|risk|confirm|' . $body, $seed);
            return [hash_equals($expected, (string)($payload['ratchet'] ?? '')) && (int)($payload['riskScore'] ?? 0) >= 70, 'risk sdk ratchet'];
        case 'K03':
            $body = (string)($payload['account'] ?? '') . '|' . (string)($payload['operation'] ?? '');
            $expected = hash_hmac('sha256', 'arena>auth>risk>' . $body, $seed);
            return [hash_equals($expected, (string)($payload['businessProof'] ?? '')) && ($payload['state'] ?? []) === ['arena','auth','risk','business'], 'multi-domain state machine'];
    }
    return [false, 'unknown challenge'];
}

function arena_record(string $id, bool $ok, string $detail): array
{
    $item = arena_find($id);
    $pdo = arena_db();
    $stmt = $pdo->prepare('INSERT INTO attempts (sid, challenge_id, ok, detail, created_at) VALUES (?, ?, ?, ?, ?)');
    $stmt->execute([arena_sid(), $id, $ok ? 1 : 0, $detail, time()]);
    $solved = false;
    if ($ok && $item) {
        $stmt = $pdo->prepare('INSERT OR IGNORE INTO solves (sid, challenge_id, score, solved_at) VALUES (?, ?, ?, ?)');
        $stmt->execute([arena_sid(), $id, $item['score'], time()]);
        $solved = $stmt->rowCount() > 0;
    }
    return ['solved' => $ok, 'newSolve' => $solved, 'score' => $ok && $item ? $item['score'] : 0, 'detail' => $detail];
}

arena_start();
