const nativeToString = Function.prototype.toString;
const encoder = new TextEncoder();

function isNativeFunction(value) {
  try {
    return /\[native code\]/.test(nativeToString.call(value));
  } catch {
    return false;
  }
}

async function assetHash() {
  const response = await fetch(new URL('../app.js', import.meta.url), {cache: 'no-store'});
  const bytes = await response.arrayBuffer();
  return Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256', bytes)), (x) => x.toString(16).padStart(2, '0')).join('');
}

export function guardMessage(signals) {
  return [
    Number(signals.probeVersion || 0),
    signals.probeToken || '',
    signals.sourceHash || '',
    Number(signals.score || 0),
    Number(signals.debuggerDelta || 0),
    Number(signals.timingVariance || 0),
    Number(signals.viewportGap || 0),
    signals.visibility || '',
    signals.consoleGuard ? '1' : '0',
    signals.toStringGuard ? '1' : '0',
    signals.stackGuard ? '1' : '0',
  ].join('|');
}

export async function collectRuntimeSignals(nonce = '') {
  const probeVersion = 2;
  const probeToken = Array.from(new Uint8Array(await crypto.subtle.digest('SHA-256', encoder.encode(`${nonce}|guard/2`))), (x) => x.toString(16).padStart(2, '0')).join('').slice(0, 16);
  const deltas = [];
  const rounds = 3 + (parseInt(nonce.slice(0, 2) || '0', 16) % 3);
  for (let index = 0; index < rounds; index += 1) {
    const started = performance.now();
    debugger;
    deltas.push(Math.round(performance.now() - started));
  }
  const debuggerDelta = Math.max(...deltas);
  const timingVariance = Math.max(...deltas) - Math.min(...deltas);
  const viewportGap = Math.max(Math.abs((window.outerWidth || 0) - window.innerWidth), Math.abs((window.outerHeight || 0) - window.innerHeight));
  const consoleGuard = ['log', 'warn', 'error'].every((name) => isNativeFunction(console[name]));
  const toStringGuard = nativeToString === Function.prototype.toString && isNativeFunction(crypto.subtle.digest) && isNativeFunction(performance.now);
  const stackGuard = !String(new Error().stack || '').includes('eval');
  const visibility = document.visibilityState;
  const sourceHash = await assetHash();
  let score = 100;
  if (debuggerDelta > 120) score -= 35;
  if (timingVariance > 60) score -= 15;
  if (!consoleGuard) score -= 20;
  if (!toStringGuard) score -= 20;
  if (!stackGuard) score -= 10;
  if (viewportGap > 160) score -= 20;
  if (visibility !== 'visible') score -= 20;
  if (!document.hasFocus()) score -= 10;
  return {
    probeVersion,
    probeToken,
    score: Math.max(0, score),
    debuggerDelta,
    timingVariance,
    viewportGap,
    visibility,
    consoleGuard,
    toStringGuard,
    stackGuard,
    sourceHash,
  };
}
