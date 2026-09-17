const nativeToString = Function.prototype.toString;

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
    signals.sourceHash || '',
    Number(signals.score || 0),
    Number(signals.debuggerDelta || 0),
    Number(signals.viewportGap || 0),
    signals.visibility || '',
    signals.consoleGuard ? '1' : '0',
  ].join('|');
}

export async function collectRuntimeSignals() {
  const started = performance.now();
  debugger;
  const debuggerDelta = Math.round(performance.now() - started);
  const viewportGap = Math.max(Math.abs((window.outerWidth || 0) - window.innerWidth), Math.abs((window.outerHeight || 0) - window.innerHeight));
  const consoleGuard = ['log', 'warn', 'error'].every((name) => isNativeFunction(console[name]));
  const visibility = document.visibilityState;
  const sourceHash = await assetHash();
  let score = 100;
  if (debuggerDelta > 120) score -= 35;
  if (!consoleGuard) score -= 30;
  if (viewportGap > 160) score -= 20;
  if (visibility !== 'visible') score -= 20;
  if (!document.hasFocus()) score -= 10;
  return {
    score: Math.max(0, score),
    debuggerDelta,
    viewportGap,
    visibility,
    consoleGuard,
    sourceHash,
  };
}
