import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

import { initSync, start_registration, finish_registration, encrypt, start_login, finish_login, decrypt } from '../web/static/wasm/safex_wasm.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const wasmPath = path.resolve(__dirname, '../web/static/wasm/safex_wasm_bg.wasm');

const wasmBytes = fs.readFileSync(wasmPath);
initSync(wasmBytes);

const origin = process.env.SAFEX_ORIGIN || 'http://127.0.0.1:18080';
const pin = '123456';

const bytesToBase64 = (bytes) => Buffer.from(bytes).toString('base64');
const base64ToBytes = (b64) => new Uint8Array(Buffer.from(b64, 'base64'));

const registrationStart = start_registration(pin);
const registerResp = await fetch(`${origin}/opaque/register/start`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ request: bytesToBase64(registrationStart.message) }),
});
const registerPayload = await registerResp.json();
if (!registerResp.ok) {
  console.error('register failed', registerPayload);
  process.exit(1);
}
const serverResponseBytes = base64ToBytes(registerPayload.response);
const registrationFinish = finish_registration(
  registrationStart.handle,
  pin,
  serverResponseBytes,
);
const secretId = registerPayload.secretId;

const messageBytes = new TextEncoder().encode('hello world');
const encryptedBytes = encrypt(registrationFinish.export_key, messageBytes);
const encryptedBlob = new Blob([encryptedBytes], { type: 'application/octet-stream' });

const formData = new FormData();
formData.set('secret_id', secretId);
formData.set('opaque_upload', bytesToBase64(registrationFinish.upload));
formData.set('ttl_minutes', '5');
formData.set('payload_type', 'text');
formData.set('file', encryptedBlob, 'message.encrypted');

const createResp = await fetch(`${origin}/secrets`, { method: 'POST', body: formData });
const createText = await createResp.text();
console.log('create status', createResp.status);
if (!createResp.ok) {
  console.error('create failed', createText);
  process.exit(1);
}

const loginStart = start_login(pin);
const loginInitResp = await fetch(`${origin}/opaque/login/start`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    secretId,
    request: bytesToBase64(loginStart.message),
  }),
});
const loginPayload = await loginInitResp.json();
console.log('login start status', loginInitResp.status, loginPayload);
if (!loginInitResp.ok) {
  console.error('login start failed');
  process.exit(1);
}

try {
  const serverBytes = base64ToBytes(loginPayload.response);
  const loginFinish = finish_login(loginStart.handle, pin, serverBytes);
  console.log('finish login export key length', loginFinish.export_key.length);
  const form = new FormData();
  form.set('secret_id', secretId);
  form.set('session_id', loginPayload.sessionId);
  form.set('finalization', bytesToBase64(loginFinish.finalization));
  const revealResp = await fetch(`${origin}/secrets/reveal`, { method: 'POST', body: form });
  const revealHtml = await revealResp.text();
  console.log('reveal status', revealResp.status);
  console.log(revealHtml.slice(0, 200));
} catch (error) {
  console.error('finish_login failed', error);
  process.exit(1);
}
