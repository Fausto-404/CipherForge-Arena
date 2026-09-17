const hex = (buffer) => Array.from(new Uint8Array(buffer), (x) => x.toString(16).padStart(2, '0')).join('');
export async function signQuote(seed, symbol, quantity) {
  const key = await crypto.subtle.importKey('raw', new TextEncoder().encode(seed), {name: 'HMAC', hash: 'SHA-256'}, false, ['sign']);
  const sig = await crypto.subtle.sign('HMAC', key, new TextEncoder().encode(`${symbol}|${quantity}`));
  return hex(sig);
}
