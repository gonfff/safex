use crate::common::{decrypt_payload, encrypt_payload, Suite};
use argon2::Argon2;
use opaque_ke::{
    ClientLogin, ClientLoginFinishParameters, ClientRegistration,
    ClientRegistrationFinishParameters, CredentialResponse, Identifiers, RegistrationResponse,
};
use rand::rngs::OsRng;
use std::cell::{Cell, RefCell};
use std::collections::HashMap;
use wasm_bindgen::prelude::*;

type JsResult<T> = Result<T, JsValue>;

thread_local! {
    static REGISTRATION_STATES: RefCell<HashMap<u32, ClientRegistration<Suite>>> =
        RefCell::new(HashMap::new());
    static LOGIN_STATES: RefCell<HashMap<u32, ClientLogin<Suite>>> =
        RefCell::new(HashMap::new());
    static NEXT_HANDLE: Cell<u32> = Cell::new(0);
}

fn next_handle() -> u32 {
    NEXT_HANDLE.with(|counter| {
        let current = counter.get();
        let next = if current == u32::MAX { 1 } else { current + 1 };
        counter.set(next);
        current
    })
}

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

#[wasm_bindgen]
pub fn start_registration(pin: &str) -> JsResult<RegistrationStartResult> {
    let mut rng = OsRng;
    let start = ClientRegistration::<Suite>::start(&mut rng, pin.as_bytes()).map_err(to_js_err)?;
    let handle = store_registration_state(start.state);
    Ok(RegistrationStartResult::new(
        handle,
        start.message.serialize().to_vec(),
    ))
}

#[wasm_bindgen]
pub fn finish_registration(
    handle: u32,
    pin: &str,
    server_response: &[u8],
) -> JsResult<RegistrationFinishResult> {
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

#[wasm_bindgen]
pub fn start_login(pin: &str) -> JsResult<LoginStartResult> {
    let mut rng = OsRng;
    let start = ClientLogin::<Suite>::start(&mut rng, pin.as_bytes()).map_err(to_js_err)?;
    let handle = store_login_state(start.state);
    Ok(LoginStartResult::new(
        handle,
        start.message.serialize().to_vec(),
    ))
}

#[wasm_bindgen]
pub fn finish_login(handle: u32, pin: &str, server_response: &[u8]) -> JsResult<LoginFinishResult> {
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

#[wasm_bindgen]
pub fn encrypt(export_key: &[u8], data: &[u8]) -> JsResult<Vec<u8>> {
    encrypt_payload(export_key, data).map_err(to_js_err)
}

#[wasm_bindgen]
pub fn decrypt(export_key: &[u8], payload: &[u8]) -> JsResult<Vec<u8>> {
    decrypt_payload(export_key, payload).map_err(to_js_err)
}

fn store_registration_state(state: ClientRegistration<Suite>) -> u32 {
    let handle = next_handle();
    REGISTRATION_STATES.with(|states| {
        states.borrow_mut().insert(handle, state);
    });
    handle
}

fn take_registration_state(handle: u32) -> JsResult<ClientRegistration<Suite>> {
    REGISTRATION_STATES
        .with(|states| states.borrow_mut().remove(&handle))
        .ok_or_else(|| JsValue::from_str("unknown registration handle"))
}

fn store_login_state(state: ClientLogin<Suite>) -> u32 {
    let handle = next_handle();
    LOGIN_STATES.with(|states| {
        states.borrow_mut().insert(handle, state);
    });
    handle
}

fn take_login_state(handle: u32) -> JsResult<ClientLogin<Suite>> {
    LOGIN_STATES
        .with(|states| states.borrow_mut().remove(&handle))
        .ok_or_else(|| JsValue::from_str("unknown login handle"))
}

fn to_js_err<E: core::fmt::Display>(err: E) -> JsValue {
    JsValue::from_str(&err.to_string())
}

#[cfg(all(test, target_arch = "wasm32"))]
mod tests;
