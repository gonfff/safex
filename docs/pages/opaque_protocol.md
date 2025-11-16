# OPAQUE Protocol in Safex

OPAQUE (Oblivious Pseudorandom Functions with Applications to Key Exchange) is a modern password-authenticated key exchange (PAKE) protocol that provides advanced security properties for PIN-based authentication in Safex.

## Why OPAQUE?

### Traditional PIN Problems

In typical systems, PIN verification has several vulnerabilities:

```mermaid
sequenceDiagram
    participant C as Client
    participant S as Server
    participant A as Attacker

    Note over C,A: Traditional PIN Authentication (Vulnerable)
    C->>S: PIN in plaintext/hash
    S->>S: Store PIN hash
    A->>S: Compromise server
    A->>A: Offline brute force PIN hash
    A->>A: Recover original PIN
```

**Security Issues**:

- Server sees the PIN (even if hashed)
- Server compromise exposes PIN hashes
- Offline brute-force attacks possible
- Network eavesdropping can reveal PINs

### OPAQUE Solution

OPAQUE eliminates these vulnerabilities through cryptographic innovation:

```mermaid
sequenceDiagram
    participant C as Client (Browser)
    participant S as Server (Go Backend)
    participant R as Rust OPAQUE

    Note over C,R: OPAQUE Authentication (Secure)
    C->>C: Generate OPAQUE client state
    C->>S: Send blinded PIN (no PIN info leaked)
    S->>R: Process with OPAQUE server
    R->>S: Generate response (no PIN knowledge)
    S->>C: Return cryptographic challenge
    C->>C: Prove PIN knowledge cryptographically
    C->>S: Send proof (PIN never transmitted)
    S->>R: Verify proof
    R->>S: Authentication result (success/failure)
```

**Security Advantages**:

- ✅ Server never sees the PIN
- ✅ No offline attacks possible
- ✅ Network traffic reveals nothing about PIN
- ✅ Quantum-resistant cryptography

## How OPAQUE Works in Safex

### Registration Flow (Secret Creation)

When a user creates a secret with a PIN:

```mermaid
sequenceDiagram
    participant U as User
    participant B as Browser (WASM)
    participant S as Server
    participant R as Rust OPAQUE Library

    U->>B: Enter PIN for new secret
    B->>B: Start OPAQUE registration
    Note over B: client.start_registration(pin)
    B->>S: POST /opaque/register/start
    S->>R: opaque_registration_response()
    R->>R: Generate server response
    R->>S: Server registration data
    S->>B: Registration response
    B->>B: Complete registration
    Note over B: client.finish_registration(response)
    B->>B: Generate OPAQUE record
    B->>S: POST /secrets (with OPAQUE record)
    S->>S: Store OPAQUE record (no PIN info)
```

### Login Flow (Secret Retrieval)

When a user wants to retrieve a secret:

```mermaid
sequenceDiagram
    participant U as User
    participant B as Browser (WASM)
    participant S as Server
    participant R as Rust OPAQUE Library

    U->>B: Enter PIN to decrypt secret
    B->>B: Start OPAQUE login
    Note over B: client.start_login(pin)
    B->>S: POST /opaque/login/start
    S->>S: Load stored OPAQUE record
    S->>R: opaque_login_start(record, request)
    R->>R: Verify and generate challenge
    R->>S: Session ID + credential response
    S->>B: Return session + challenge
    B->>B: Complete login proof
    Note over B: client.finish_login(challenge)
    B->>S: POST /secrets/reveal (with proof)
    S->>R: opaque_login_finish(session, proof)
    R->>R: Verify PIN proof
    R->>S: Authentication success/failure
    S->>B: Return encrypted secret (if success)
    B->>B: Decrypt secret locally
```
