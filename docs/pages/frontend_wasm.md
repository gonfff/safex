# Frontend & WebAssembly Documentation

The Safex frontend combines modern web technologies with WebAssembly for secure client-side cryptography.

## Technology Stack

### Core Technologies

- **HTMX**: Dynamic HTML over the wire
- **TailwindCSS + DaisyUI**: Utility-first CSS framework
- **WebAssembly (WASM)**: Client-side cryptographic operations
- **Go Templates**: Server-side HTML templating

### Build Pipeline

```mermaid
graph LR
    TW[TailwindCSS Source] --> CSS[Compiled CSS]
    HTMX[HTMX Library] --> JS[Static Assets]
    Rust[Rust Crypto] --> WASM[WASM Module]

    CSS --> Bundle[Asset Bundle]
    JS --> Bundle
    WASM --> Bundle

    Bundle --> Server[Go Web Server]
```

## Frontend Architecture

### Page Flow

```mermaid
sequenceDiagram
    participant U as User
    participant B as Browser
    participant S as Server
    participant W as WASM Module

    Note over U,W: Secret Creation Flow
    U->>B: Navigate to /
    B->>S: GET /
    S->>B: Return home.html
    U->>B: Fill form + file/text
    B->>W: encrypt_secret(data, pin)
    W->>B: {encrypted_data, salt}
    B->>W: opaque_register_start(pin)
    W->>B: registration_request
    B->>S: POST /opaque/register/start
    S->>B: registration_response
    B->>W: opaque_register_finish(response)
    W->>B: opaque_record
    B->>S: POST /secrets (encrypted + opaque_record)
    S->>B: Return createResult.html with link

    Note over U,W: Secret Retrieval Flow
    U->>B: Navigate to /secrets/:id
    B->>S: GET /secrets/:id
    S->>B: Return retrieve.html
    U->>B: Enter PIN
    B->>W: opaque_login_start(pin)
    W->>B: credential_request
    B->>S: POST /opaque/login/start
    S->>B: {session_id, credential_response}
    B->>W: opaque_login_finish(response)
    W->>B: finalization
    B->>S: POST /secrets/reveal
    S->>B: Return revealResult.html with encrypted data
    B->>W: decrypt_secret(encrypted, pin, salt)
    W->>B: decrypted_data
    B->>U: Display secret content
```
