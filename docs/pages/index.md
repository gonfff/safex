# Safex - Secure Secret Exchange

Safex is a security‑first secret sharing service that keeps sensitive data out of common communication channels. It's run for real on [Koyeb](https://safex.koyeb.app), so you can try it right now. You can also run your own Docker image from [GHCR](https://github.com/gonfff/safex/pkgs/container/safex), or build it yourself because the entire project is open source.

## Core concepts

- Secrets are <strong>encrypted and decrypted locally in the browser</strong> with WebAssembly (WASM) before they ever touch the server, so backend compromises never expose cleartext data.
- The recipient <strong>must prove knowledge of the PIN</strong> via the OPAQUE protocol before any download, blocking offline brute-force attacks and network PIN transmission.
- Expiration policies and view limits ensure every shared secret has a defined lifetime and data will be <strong>destroyed automatically after reading or expiry</strong>.

## How safety is achieved

- Safex does not store the original secret or its PIN.
- No sensitive information is written to any logs.
- Messages are permanently deleted as soon as they are read or expire.
- An attacker needs both the unique link and the PIN to intercept a message.
- The code is open source and can be audited by anyone.

### Warning

Redis and S3 backends not tested yet!!!
