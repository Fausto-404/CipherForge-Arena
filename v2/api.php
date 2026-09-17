<?php

declare(strict_types=1);
require __DIR__ . '/bootstrap.php';

$action = $_GET['action'] ?? 'list';

if ($action === 'list') {
    $pdo = arena_db();
    $solved = [];
    $stmt = $pdo->prepare('SELECT challenge_id, score, solved_at FROM solves WHERE sid = ?');
    $stmt->execute([arena_sid()]);
    foreach ($stmt as $row) $solved[$row['challenge_id']] = $row;
    $rankUnlocked = ['bronze' => true, 'silver' => true, 'gold' => true, 'platinum' => true, 'diamond' => true, 'king' => true];
    $items = [];
    foreach (arena_challenges() as $rank => $challenges) {
        $items[$rank] = array_map(static function (array $item) use ($solved): array {
            $item['solved'] = isset($solved[$item['id']]);
            $item['solvedAt'] = $solved[$item['id']]['solved_at'] ?? null;
            return $item;
        }, $challenges);
    }
    arena_json(['version' => ARENA_VERSION, 'sid' => arena_sid(), 'ranks' => $items, 'unlocked' => $rankUnlocked]);
}

if ($action === 'context') {
    $id = strtoupper((string)($_GET['id'] ?? ''));
    if (!arena_find($id)) arena_json(['error' => 'unknown challenge'], 404);
    arena_json(['challenge' => arena_find($id), 'context' => arena_public_context($id)]);
}

if ($action === 'submit') {
    $input = json_decode(file_get_contents('php://input'), true);
    if (!is_array($input)) arena_json(['error' => 'invalid json'], 400);
    $id = strtoupper((string)($input['id'] ?? ''));
    $item = arena_find($id);
    if (!$item) arena_json(['error' => 'unknown challenge'], 404);
    $ctx = arena_context($id);
    [$ok, $detail] = arena_verify($id, is_array($input['payload'] ?? null) ? $input['payload'] : [], $ctx);
    arena_json(['ok' => $ok, ...arena_record($id, $ok, $detail)]);
}

arena_json(['error' => 'unknown action'], 404);
