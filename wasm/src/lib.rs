use aes_gcm::aead::{Aead, KeyInit, Nonce as AeadNonce};
use aes_gcm::Aes256Gcm;
use argon2::Argon2;
use hkdf::Hkdf;
use opaque_ke::ciphersuite::CipherSuite;
use opaque_ke::key_exchange::tripledh::TripleDh;
use opaque_ke::{
    ClientLogin, ClientLoginFinishParameters, ClientRegistration,
    ClientRegistrationFinishParameters, CredentialResponse, Identifiers, RegistrationResponse,
    Ristretto255,
};
use rand::rngs::OsRng;
use rand::RngCore;
use sha2::Sha512;
use std::cell::{Cell, RefCell};
use std::collections::HashMap;
use wasm_bindgen::prelude::*;

// Parameters for AES-GCM payloads and HKDF key schedule.
const NONCE_LEN: usize = 12;
const AES_KEY_LEN: usize = 32;
const HKDF_SALT: &[u8] = b"safex/opaque/export-key";
const HKDF_INFO: &[u8] = b"safex/aes256-gcm";

type CipherNonce = AeadNonce<Aes256Gcm>;
type Suite = SafexCipherSuite;

// State for logins and registrations between WASM calls.
thread_local! {
    static REGISTRATION_STATES: RefCell<HashMap<u32, ClientRegistration<Suite>>> =
        RefCell::new(HashMap::new());
    static LOGIN_STATES: RefCell<HashMap<u32, ClientLogin<Suite>>> =
        RefCell::new(HashMap::new());
    static NEXT_HANDLE: Cell<u32> = Cell::new(0);
}

// Cipher suite tying OPAQUE to Ristretto255 + TripleDH + Argon2.
struct SafexCipherSuite;

impl CipherSuite for SafexCipherSuite {
    type OprfCs = Ristretto255;
    type KeGroup = Ristretto255;
    type KeyExchange = TripleDh;
    type Ksf = Argon2<'static>;
}

// JS wrapper around OPAQUE registration start artifacts.
#[wasm_bindgen]
pub struct RegistrationStartResult {
    handle: u32,
    message: Vec<u8>,
}

#[wasm_bindgen]
impl RegistrationStartResult {
    #[wasm_bindgen(getter)]
    pub fn handle(&self) -> u32 {
        self.handle
    }

    #[wasm_bindgen(getter)]
    pub fn message(&self) -> Vec<u8> {
        self.message.clone()
    }
}

impl RegistrationStartResult {
    fn new(handle: u32, message: Vec<u8>) -> Self {
        Self { handle, message }
    }
}

// JS wrapper around OPAQUE registration finish artifacts.
#[wasm_bindgen]
pub struct RegistrationFinishResult {
    upload: Vec<u8>,
    export_key: Vec<u8>,
}

#[wasm_bindgen]
impl RegistrationFinishResult {
    #[wasm_bindgen(getter)]
    pub fn upload(&self) -> Vec<u8> {
        self.upload.clone()
    }

    #[wasm_bindgen(getter)]
    pub fn export_key(&self) -> Vec<u8> {
        self.export_key.clone()
    }
}

impl RegistrationFinishResult {
    fn new(upload: Vec<u8>, export_key: Vec<u8>) -> Self {
        Self { upload, export_key }
    }
}

// JS wrapper containing login KE1 message and state.
#[wasm_bindgen]
pub struct LoginStartResult {
    handle: u32,
    message: Vec<u8>,
}

#[wasm_bindgen]
impl LoginStartResult {
    #[wasm_bindgen(getter)]
    pub fn handle(&self) -> u32 {
        self.handle
    }

    #[wasm_bindgen(getter)]
    pub fn message(&self) -> Vec<u8> {
        self.message.clone()
    }
}

impl LoginStartResult {
    fn new(handle: u32, message: Vec<u8>) -> Self {
        Self { handle, message }
    }
}

// JS wrapper returned after login finishes.
#[wasm_bindgen]
pub struct LoginFinishResult {
    finalization: Vec<u8>,
    export_key: Vec<u8>,
    session_key: Vec<u8>,
}

#[wasm_bindgen]
impl LoginFinishResult {
    #[wasm_bindgen(getter)]
    pub fn finalization(&self) -> Vec<u8> {
        self.finalization.clone()
    }

    #[wasm_bindgen(getter)]
    pub fn export_key(&self) -> Vec<u8> {
        self.export_key.clone()
    }

    #[wasm_bindgen(getter)]
    pub fn session_key(&self) -> Vec<u8> {
        self.session_key.clone()
    }
}

impl LoginFinishResult {
    fn new(finalization: Vec<u8>, export_key: Vec<u8>, session_key: Vec<u8>) -> Self {
        Self {
            finalization,
            export_key,
            session_key,
        }
    }
}

// Step 1 of OPAQUE registration: generate a blinded password request.
#[wasm_bindgen]
pub fn start_registration(pin: &str) -> Result<RegistrationStartResult, JsValue> {
    let mut rng = OsRng;
    let start = ClientRegistration::<Suite>::start(&mut rng, pin.as_bytes()).map_err(to_js_err)?;
    let handle = store_registration_state(start.state);
    Ok(RegistrationStartResult::new(
        handle,
        start.message.serialize().to_vec(),
    ))
}

// Step 2 of OPAQUE registration: unblind server response and return upload.
#[wasm_bindgen]
pub fn finish_registration(
    handle: u32,
    pin: &str,
    server_response: &[u8],
) -> Result<RegistrationFinishResult, JsValue> {
    let state = take_registration_state(handle)?;
    let response =
        RegistrationResponse::<Suite>::deserialize(server_response).map_err(to_js_err)?;
    let mut rng = OsRng;
    let ksf = Argon2::default();
    let params = ClientRegistrationFinishParameters::new(Identifiers::default(), Some(&ksf));
    let result = state
        .finish(&mut rng, pin.as_bytes(), response, params)
        .map_err(to_js_err)?;
    Ok(RegistrationFinishResult::new(
        result.message.serialize().to_vec(),
        result.export_key.to_vec(),
    ))
}

// Step 1 of OPAQUE login: produce credential request and cache state.
#[wasm_bindgen]
pub fn start_login(pin: &str) -> Result<LoginStartResult, JsValue> {
    let mut rng = OsRng;
    let start = ClientLogin::<Suite>::start(&mut rng, pin.as_bytes()).map_err(to_js_err)?;
    let handle = store_login_state(start.state);
    Ok(LoginStartResult::new(
        handle,
        start.message.serialize().to_vec(),
    ))
}

// Step 2 of OPAQUE login: finalize handshake and return export/session keys.
#[wasm_bindgen]
pub fn finish_login(
    handle: u32,
    pin: &str,
    server_response: &[u8],
) -> Result<LoginFinishResult, JsValue> {
    let state = take_login_state(handle)?;
    let response = CredentialResponse::<Suite>::deserialize(server_response).map_err(to_js_err)?;
    let ksf = Argon2::default();
    let params = ClientLoginFinishParameters::new(None, Identifiers::default(), Some(&ksf));
    let result = state
        .finish(pin.as_bytes(), response, params)
        .map_err(to_js_err)?;
    Ok(LoginFinishResult::new(
        result.message.serialize().to_vec(),
        result.export_key.to_vec(),
        result.session_key.to_vec(),
    ))
}

// Encrypt arbitrary bytes using the export key-derived AES key.
#[wasm_bindgen]
pub fn encrypt(export_key: &[u8], data: &[u8]) -> Result<Vec<u8>, JsValue> {
    let key = derive_aes_key(export_key)?;
    let mut nonce = [0u8; NONCE_LEN];
    let mut rng = OsRng;
    rng.fill_bytes(&mut nonce);
    let cipher = Aes256Gcm::new_from_slice(&key).map_err(to_js_err)?;
    let nonce_ga = CipherNonce::from(nonce);
    let ciphertext = cipher.encrypt(&nonce_ga, data).map_err(to_js_err)?;
    let mut payload = Vec::with_capacity(NONCE_LEN + ciphertext.len());
    payload.extend_from_slice(&nonce);
    payload.extend_from_slice(&ciphertext);
    Ok(payload)
}

// Decrypt bytes previously produced by `encrypt`.
#[wasm_bindgen]
pub fn decrypt(export_key: &[u8], payload: &[u8]) -> Result<Vec<u8>, JsValue> {
    if payload.len() < NONCE_LEN {
        return Err(JsValue::from_str("payload too small"));
    }
    let key = derive_aes_key(export_key)?;
    let (nonce_bytes, ciphertext) = payload.split_at(NONCE_LEN);
    let mut nonce = [0u8; NONCE_LEN];
    nonce.copy_from_slice(nonce_bytes);
    let cipher = Aes256Gcm::new_from_slice(&key).map_err(to_js_err)?;
    let nonce_ga = CipherNonce::from(nonce);
    cipher.decrypt(&nonce_ga, ciphertext).map_err(to_js_err)
}

// Deterministically derive an AES-256 key from the 64-byte export key.
fn derive_aes_key(export_key: &[u8]) -> Result<[u8; AES_KEY_LEN], JsValue> {
    if export_key.is_empty() {
        return Err(JsValue::from_str("export_key must not be empty"));
    }
    let hkdf = Hkdf::<Sha512>::new(Some(HKDF_SALT), export_key);
    let mut key = [0u8; AES_KEY_LEN];
    hkdf.expand(HKDF_INFO, &mut key)
        .map_err(|_| JsValue::from_str("unable to derive key from export_key"))?;
    Ok(key)
}

// Keep registration state alive between WASM calls.
fn store_registration_state(state: ClientRegistration<Suite>) -> u32 {
    let handle = next_handle();
    REGISTRATION_STATES.with(|states| {
        states.borrow_mut().insert(handle, state);
    });
    handle
}

// Remove stored registration state; errors if handle missing.
fn take_registration_state(handle: u32) -> Result<ClientRegistration<Suite>, JsValue> {
    REGISTRATION_STATES
        .with(|states| states.borrow_mut().remove(&handle))
        .ok_or_else(|| JsValue::from_str("unknown registration handle"))
}

// Keep login state alive between WASM calls.
fn store_login_state(state: ClientLogin<Suite>) -> u32 {
    let handle = next_handle();
    LOGIN_STATES.with(|states| {
        states.borrow_mut().insert(handle, state);
    });
    handle
}

// Remove stored login state; errors if handle missing.
fn take_login_state(handle: u32) -> Result<ClientLogin<Suite>, JsValue> {
    LOGIN_STATES
        .with(|states| states.borrow_mut().remove(&handle))
        .ok_or_else(|| JsValue::from_str("unknown login handle"))
}

// Simple monotonically increasing handle generator.
fn next_handle() -> u32 {
    NEXT_HANDLE.with(|counter| {
        let current = counter.get();
        let next = if current == u32::MAX { 1 } else { current + 1 };
        counter.set(next);
        current
    })
}

fn to_js_err<E: core::fmt::Display>(err: E) -> JsValue {
    JsValue::from_str(&err.to_string())
}

#[cfg(all(test, target_arch = "wasm32"))]
mod tests;
