# Safex

<p align="center">
  <strong>Safe secret exchange</strong><br>
  <a href="https://github.com/gonfff/safex/actions/workflows/ci.yml">
    <img src="https://github.com/gonfff/safex/actions/workflows/ci.yml/badge.svg" alt="Build Status">
  </a>
  <a href="https://img.shields.io/github/license/gonfff/safex">
    <img src="https://img.shields.io/github/license/gonfff/safex" alt="License: MIT">
  </a>
  <a href="https://goreportcard.com/report/github.com/gonfff/safex/app">
    <img src="https://goreportcard.com/badge/github.com/gonfff/safex/app" alt="Go Report Card">
  </a>
  <a href="https://github.com/gonfff/safex/pkgs/container/safex">
    <img src="https://img.shields.io/badge/image_size-39MB-blue?logo=docker&color=blue" alt="Docker image size">
  </a>
</p>

Safex is a security‑first secret sharing service that keeps sensitive data <strong>with zero server trust</strong>. It’s run for real on [Koyeb](https://safex.koyeb.app), so you can try it right now. You can also run your own Docker image from [GHCR](https://github.com/gonfff/safex/pkgs/container/safex), or build it yourself because the entire project is open source.

## Core concepts

- Secrets are <strong>encrypted and decrypted locally in the browser</strong> with WebAssembly (WASM) before they ever touch the server, so backend compromises never expose cleartext data.
- The recipient <strong>must prove knowledge of the PIN</strong> via the OPAQUE protocol before any download, blocking offline brute-force attacks and network PIN transmission.
- Expiration policies and view limits ensure every shared secret has a defined lifetime and data will be <strong>destroyed automatically after reading or expiry</strong>.

## How safety is achieved

- Secrets are encrypted and decrypted locally in the browser with WebAssembly (WASM).
- Secrets are stored encrypted on the server, which has no access to plaintext data.
- Safex does not receive PINs at any point, thanks to the OPAQUE protocol.
- No sensitive information is written to any logs.
- Messages are permanently deleted as soon as they are read or expire.
- An attacker needs both the unique link and the PIN to intercept a message.
- The code is open source and can be audited by anyone.

## Documentation

A full walkthrough is available on the [Docs page](https://gonfff.github.io/safex/).

### Warning

Redis and S3 backends not tested yet!!!

## Usage

1. Create a new secret with text or file, set expiration and choose a PIN.
   ![Create](screenshots/create.png)
2. Share the generated link and PIN with the recipient via separate channels.
   ![Share](screenshots/link.png)
3. The recipient opens the link, enters the PIN, and retrieves the secret.
   ![Retrieve](screenshots/retrieve.png)
4. Read/download/copy the secret before it self-destructs.
   ![Success](screenshots/load.png)
