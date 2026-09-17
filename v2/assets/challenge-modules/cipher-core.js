const decode = (value) => new TextDecoder().decode(new Uint8Array(value.match(/.{2}/g).map((x) => parseInt(x, 16))));
const table = [decode('6366782f31'), decode('43465831'), 16, 8, 12, 255];
const encoder = new TextEncoder();

function rot(value, shift) {
  return ((value << shift) | (value >>> (8 - shift))) & table[5];
}

function xorBlock(left, right) {
  return left.map((value, index) => value ^ right[index]);
}

function makeBox(key) {
  return Uint8Array.from({length: 256}, (_, index) => key[(index * 13) % key.length] ^ ((index * 29 + 17) & table[5]));
}

function blockRound(left, right, key, box, round) {
  const nextRight = new Uint8Array(table[3]);
  for (let i = 0; i < table[3]; i += 1) {
    const a = key[(round * 7 + i) % key.length];
    const b = key[(round * 11 + i * 3) % key.length];
    const tweak = ((round + 1) * (i + 3)) & table[5];
    const nonlinear = box[(right[i] + a + tweak) & table[5]];
    nextRight[i] = left[i] ^ (rot(nonlinear, (round + i) % 7 + 1) ^ b);
  }
  return [right, nextRight];
}

function blockSeal(block, key) {
  const box = makeBox(key);
  let left = block.slice(0, table[3]);
  let right = block.slice(table[3]);
  for (let round = 0; round < table[1].length * 2 + table[2] - 8; round += 1) [left, right] = blockRound(left, right, key, box, round);
  return new Uint8Array([...left, ...right]);
}

function pad(bytes) {
  const amount = table[2] - (bytes.length % table[2]) || table[2];
  return new Uint8Array([...bytes, ...Array.from({length: amount}, () => amount)]);
}

function concat(...parts) {
  const size = parts.reduce((total, part) => total + part.length, 0);
  const output = new Uint8Array(size);
  let offset = 0;
  for (const part of parts) {
    output.set(part, offset);
    offset += part.length;
  }
  return output;
}

function hexBytes(value) {
  return new Uint8Array((value.match(/.{1,2}/g) || []).map((chunk) => parseInt(chunk, 16)));
}

function b64url(bytes) {
  return btoa(String.fromCharCode(...bytes)).replaceAll('+', '-').replaceAll('/', '_').replaceAll('=', '');
}

async function _seal(plainText, seed, nonce) {
  const key = new Uint8Array(await crypto.subtle.digest('SHA-256', encoder.encode(`${seed}|${nonce}|${table[0]}`)));
  const nonceBytes = hexBytes(nonce);
  const plain = encoder.encode(plainText);
  const framed = new Uint8Array([0x43, 0x46, 0x58, 0x31, plain.length >>> 8, plain.length & 255, ...plain]);
  const padded = pad(framed);
  const ciphertext = [];
  let previous = new Uint8Array(table[2]);
  for (let offset = 0; offset < padded.length; offset += table[2]) {
    const sealed = blockSeal(xorBlock(padded.slice(offset, offset + table[2]), previous), key);
    ciphertext.push(sealed);
    previous = sealed;
  }
  const body = concat(new Uint8Array([0x43, 0x46, 0x58, 0x31]), nonceBytes, ...ciphertext);
  const tag = new Uint8Array(await crypto.subtle.digest('SHA-256', concat(key, body)));
  return b64url(concat(body, tag.slice(0, table[4])));
}

export { _seal as q };
