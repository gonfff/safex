use crate::common::{CommonError, Suite};
use argon2::Argon2;
use hex::encode as hex_encode;
use opaque_ke::ciphersuite::CipherSuite;
use opaque_ke::keypair::{KeyPair, SecretKey};
use opaque_ke::{
    ClientLogin, ClientLoginFinishParameters, ClientRegistration,
    ClientRegistrationFinishParameters, CredentialFinalization, CredentialRequest,
    CredentialResponse, Identifiers, RegistrationRequest, RegistrationResponse, RegistrationUpload,
    ServerLogin, ServerLoginStartParameters, ServerRegistration, ServerSetup,
};
use rand::rngs::OsRng;
use rand::RngCore;
use sha2::{Digest, Sha512};
use std::collections::HashMap;
use std::ffi::{c_char, CString};
use std::ptr;
use std::slice;
use std::sync::{Mutex, OnceLock};
use std::time::{Duration, Instant};

pub struct SafexOpaqueManager {
    setup: ServerSetup<Suite>,
    session_ttl: Duration,
    sessions: Mutex<HashMap<String, SessionEntry>>,
}

struct SessionEntry {
    secret_id: Vec<u8>,
    state: ServerLogin<Suite>,
    expires_at: Instant,
}

#[derive(Default)]
struct ClientState {
    next_handle: u32,
    registrations: HashMap<u32, ClientRegistration<Suite>>,
    logins: HashMap<u32, ClientLogin<Suite>>,
}

impl ClientState {
    fn next_handle(&mut self) -> u32 {
        let mut next = self.next_handle.wrapping_add(1);
        if next == 0 {
            next = 1;
        }
        self.next_handle = next;
        next
    }
}

static CLIENT_STATE: OnceLock<Mutex<ClientState>> = OnceLock::new();

fn client_state() -> &'static Mutex<ClientState> {
    CLIENT_STATE.get_or_init(|| Mutex::new(ClientState::default()))
}

#[repr(C)]
pub struct SafexOpaqueBuffer {
    pub ptr: *mut u8,
    pub len: usize,
}

#[repr(C)]
pub struct SafexOpaqueLoginStart {
    pub session_id: SafexOpaqueBuffer,
    pub response: SafexOpaqueBuffer,
}

#[repr(C)]
pub struct SafexOpaqueClientStart {
    pub handle: u32,
    pub message: SafexOpaqueBuffer,
}

#[repr(C)]
pub struct SafexOpaqueRegistrationFinish {
    pub upload: SafexOpaqueBuffer,
    pub export_key: SafexOpaqueBuffer,
}

#[repr(C)]
pub struct SafexOpaqueLoginFinish {
    pub finalization: SafexOpaqueBuffer,
    pub export_key: SafexOpaqueBuffer,
    pub session_key: SafexOpaqueBuffer,
}

#[no_mangle]
pub extern "C" fn safex_opaque_manager_new(
    server_id_ptr: *const u8,
    server_id_len: usize,
    secret_key_ptr: *const u8,
    secret_key_len: usize,
    oprf_seed_ptr: *const u8,
    oprf_seed_len: usize,
    session_ttl_secs: u64,
    err_out: *mut *mut c_char,
) -> *mut SafexOpaqueManager {
    set_error(err_out, None);
    let result = (|| {
        let server_id = unsafe { read_slice(server_id_ptr, server_id_len)? };
        let secret_key = unsafe { read_slice(secret_key_ptr, secret_key_len)? };
        let oprf_seed = unsafe { read_slice(oprf_seed_ptr, oprf_seed_len)? };
        build_manager(server_id, secret_key, oprf_seed, session_ttl_secs)
    })();
    match result {
        Ok(manager) => Box::into_raw(Box::new(manager)),
        Err(err) => {
            set_error(err_out, Some(err));
            ptr::null_mut()
        }
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_manager_free(ptr: *mut SafexOpaqueManager) {
    if ptr.is_null() {
        return;
    }
    unsafe {
        drop(Box::from_raw(ptr));
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_registration_response(
    manager: *mut SafexOpaqueManager,
    secret_id_ptr: *const u8,
    secret_id_len: usize,
    request_ptr: *const u8,
    request_len: usize,
    err_out: *mut *mut c_char,
) -> SafexOpaqueBuffer {
    set_error(err_out, None);
    match unsafe { manager.as_ref() } {
        Some(mgr) => {
            let result = (|| {
                let secret_id = unsafe { read_slice(secret_id_ptr, secret_id_len)? };
                let request = unsafe { read_slice(request_ptr, request_len)? };
                mgr.registration_response(secret_id, request)
            })();
            match result {
                Ok(bytes) => to_buffer(bytes),
                Err(err) => {
                    set_error(err_out, Some(err));
                    SafexOpaqueBuffer {
                        ptr: ptr::null_mut(),
                        len: 0,
                    }
                }
            }
        }
        None => {
            set_error(
                err_out,
                Some(CommonError::new("opaque manager not initialized")),
            );
            SafexOpaqueBuffer {
                ptr: ptr::null_mut(),
                len: 0,
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_login_start(
    manager: *mut SafexOpaqueManager,
    secret_id_ptr: *const u8,
    secret_id_len: usize,
    record_ptr: *const u8,
    record_len: usize,
    request_ptr: *const u8,
    request_len: usize,
    err_out: *mut *mut c_char,
) -> SafexOpaqueLoginStart {
    set_error(err_out, None);
    match unsafe { manager.as_ref() } {
        Some(mgr) => {
            let result = (|| {
                let secret_id = unsafe { read_slice(secret_id_ptr, secret_id_len)? };
                let record = unsafe { read_slice(record_ptr, record_len)? };
                let request = unsafe { read_slice(request_ptr, request_len)? };
                mgr.login_start(secret_id, record, request)
            })();
            match result {
                Ok((session_id, response)) => SafexOpaqueLoginStart {
                    session_id: to_buffer(session_id.into_bytes()),
                    response: to_buffer(response),
                },
                Err(err) => {
                    set_error(err_out, Some(err));
                    SafexOpaqueLoginStart {
                        session_id: SafexOpaqueBuffer {
                            ptr: ptr::null_mut(),
                            len: 0,
                        },
                        response: SafexOpaqueBuffer {
                            ptr: ptr::null_mut(),
                            len: 0,
                        },
                    }
                }
            }
        }
        None => {
            set_error(
                err_out,
                Some(CommonError::new("opaque manager not initialized")),
            );
            SafexOpaqueLoginStart {
                session_id: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
                response: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_login_finish(
    manager: *mut SafexOpaqueManager,
    session_id_ptr: *const u8,
    session_id_len: usize,
    ke3_ptr: *const u8,
    ke3_len: usize,
    err_out: *mut *mut c_char,
) -> SafexOpaqueBuffer {
    set_error(err_out, None);
    match unsafe { manager.as_ref() } {
        Some(mgr) => {
            let result = (|| {
                let session_id = unsafe { read_slice(session_id_ptr, session_id_len)? };
                let ke3 = unsafe { read_slice(ke3_ptr, ke3_len)? };
                mgr.login_finish(session_id, ke3)
            })();
            match result {
                Ok(secret_id) => to_buffer(secret_id),
                Err(err) => {
                    set_error(err_out, Some(err));
                    SafexOpaqueBuffer {
                        ptr: ptr::null_mut(),
                        len: 0,
                    }
                }
            }
        }
        None => {
            set_error(
                err_out,
                Some(CommonError::new("opaque manager not initialized")),
            );
            SafexOpaqueBuffer {
                ptr: ptr::null_mut(),
                len: 0,
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_buffer_free(buffer: SafexOpaqueBuffer) {
    if buffer.ptr.is_null() {
        return;
    }
    unsafe {
        let slice = ptr::slice_from_raw_parts_mut(buffer.ptr, buffer.len);
        drop(Box::from_raw(slice));
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_string_free(ptr: *mut c_char) {
    if ptr.is_null() {
        return;
    }
    unsafe {
        drop(CString::from_raw(ptr));
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_client_start_registration(
    pin_ptr: *const u8,
    pin_len: usize,
    err_out: *mut *mut c_char,
) -> SafexOpaqueClientStart {
    set_error(err_out, None);
    let result = (|| {
        let pin = unsafe { read_slice(pin_ptr, pin_len)? };
        let mut rng = OsRng;
        let start = ClientRegistration::<Suite>::start(&mut rng, pin)
            .map_err(|e| CommonError::new(e.to_string()))?;
        let mut state = client_state().lock().unwrap();
        let handle = state.next_handle();
        state.registrations.insert(handle, start.state);
        Ok(SafexOpaqueClientStart {
            handle,
            message: to_buffer(start.message.serialize().to_vec()),
        })
    })();
    match result {
        Ok(val) => val,
        Err(err) => {
            set_error(err_out, Some(err));
            SafexOpaqueClientStart {
                handle: 0,
                message: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_client_finish_registration(
    handle: u32,
    pin_ptr: *const u8,
    pin_len: usize,
    response_ptr: *const u8,
    response_len: usize,
    err_out: *mut *mut c_char,
) -> SafexOpaqueRegistrationFinish {
    set_error(err_out, None);
    let result = (|| {
        let pin = unsafe { read_slice(pin_ptr, pin_len)? };
        let response = RegistrationResponse::<Suite>::deserialize(unsafe {
            read_slice(response_ptr, response_len)?
        })
        .map_err(|e| CommonError::new(e.to_string()))?;
        let state = {
            let mut global = client_state().lock().unwrap();
            global
                .registrations
                .remove(&handle)
                .ok_or_else(|| CommonError::new("unknown registration handle"))?
        };
        let mut rng = OsRng;
        let ksf = Argon2::default();
        let params = ClientRegistrationFinishParameters::new(Identifiers::default(), Some(&ksf));
        let finish = state
            .finish(&mut rng, pin, response, params)
            .map_err(|e| CommonError::new(e.to_string()))?;
        Ok(SafexOpaqueRegistrationFinish {
            upload: to_buffer(finish.message.serialize().to_vec()),
            export_key: to_buffer(finish.export_key.to_vec()),
        })
    })();
    match result {
        Ok(val) => val,
        Err(err) => {
            set_error(err_out, Some(err));
            SafexOpaqueRegistrationFinish {
                upload: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
                export_key: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_client_start_login(
    pin_ptr: *const u8,
    pin_len: usize,
    err_out: *mut *mut c_char,
) -> SafexOpaqueClientStart {
    set_error(err_out, None);
    let result = (|| {
        let pin = unsafe { read_slice(pin_ptr, pin_len)? };
        let mut rng = OsRng;
        let start = ClientLogin::<Suite>::start(&mut rng, pin)
            .map_err(|e| CommonError::new(e.to_string()))?;
        let mut state = client_state().lock().unwrap();
        let handle = state.next_handle();
        state.logins.insert(handle, start.state);
        Ok(SafexOpaqueClientStart {
            handle,
            message: to_buffer(start.message.serialize().to_vec()),
        })
    })();
    match result {
        Ok(val) => val,
        Err(err) => {
            set_error(err_out, Some(err));
            SafexOpaqueClientStart {
                handle: 0,
                message: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
            }
        }
    }
}

#[no_mangle]
pub extern "C" fn safex_opaque_client_finish_login(
    handle: u32,
    pin_ptr: *const u8,
    pin_len: usize,
    response_ptr: *const u8,
    response_len: usize,
    err_out: *mut *mut c_char,
) -> SafexOpaqueLoginFinish {
    set_error(err_out, None);
    let result = (|| {
        let pin = unsafe { read_slice(pin_ptr, pin_len)? };
        let response = CredentialResponse::<Suite>::deserialize(unsafe {
            read_slice(response_ptr, response_len)?
        })
        .map_err(|e| CommonError::new(e.to_string()))?;
        let state = {
            let mut global = client_state().lock().unwrap();
            global
                .logins
                .remove(&handle)
                .ok_or_else(|| CommonError::new("unknown login handle"))?
        };
        let ksf = Argon2::default();
        let params = ClientLoginFinishParameters::new(None, Identifiers::default(), Some(&ksf));
        let finish = state
            .finish(pin, response, params)
            .map_err(|e| CommonError::new(e.to_string()))?;
        Ok(SafexOpaqueLoginFinish {
            finalization: to_buffer(finish.message.serialize().to_vec()),
            export_key: to_buffer(finish.export_key.to_vec()),
            session_key: to_buffer(finish.session_key.to_vec()),
        })
    })();
    match result {
        Ok(val) => val,
        Err(err) => {
            set_error(err_out, Some(err));
            SafexOpaqueLoginFinish {
                finalization: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
                export_key: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
                session_key: SafexOpaqueBuffer {
                    ptr: ptr::null_mut(),
                    len: 0,
                },
            }
        }
    }
}

impl SafexOpaqueManager {
    fn registration_response(
        &self,
        secret_id: &[u8],
        request_bytes: &[u8],
    ) -> Result<Vec<u8>, CommonError> {
        if secret_id.is_empty() {
            return Err(CommonError::new("secret ID is required"));
        }
        let request = RegistrationRequest::<Suite>::deserialize(request_bytes)
            .map_err(|e| CommonError::new(e.to_string()))?;
        let response = ServerRegistration::<Suite>::start(&self.setup, request, secret_id)
            .map_err(|e| CommonError::new(format!("opaque registration response: {e}")))?;
        Ok(response.message.serialize().to_vec())
    }

    fn login_start(
        &self,
        secret_id: &[u8],
        record_blob: &[u8],
        request_bytes: &[u8],
    ) -> Result<(String, Vec<u8>), CommonError> {
        if secret_id.is_empty() {
            return Err(CommonError::new("secret ID is required"));
        }
        if record_blob.is_empty() {
            return Err(CommonError::new("opaque record is required"));
        }
        let record_upload = RegistrationUpload::<Suite>::deserialize(record_blob)
            .map_err(|e| CommonError::new(e.to_string()))?;
        let record = ServerRegistration::<Suite>::finish(record_upload);
        let credential_request = CredentialRequest::<Suite>::deserialize(request_bytes)
            .map_err(|e| CommonError::new(e.to_string()))?;
        let mut rng = OsRng;
        let start = ServerLogin::<Suite>::start(
            &mut rng,
            &self.setup,
            Some(record),
            credential_request,
            secret_id,
            ServerLoginStartParameters::default(),
        )
        .map_err(|e| CommonError::new(format!("opaque login init: {e}")))?;
        let session_id = generate_session_id();
        let mut sessions = self.sessions.lock().unwrap();
        sessions.insert(
            session_id.clone(),
            SessionEntry {
                secret_id: secret_id.to_vec(),
                state: start.state,
                expires_at: Instant::now() + self.session_ttl,
            },
        );
        Ok((session_id, start.message.serialize().to_vec()))
    }

    fn login_finish(
        &self,
        session_id_bytes: &[u8],
        ke3_bytes: &[u8],
    ) -> Result<Vec<u8>, CommonError> {
        if session_id_bytes.is_empty() {
            return Err(CommonError::new("session ID is required"));
        }
        let session_id = String::from_utf8(session_id_bytes.to_vec())
            .map_err(|_| CommonError::new("session ID must be valid UTF-8"))?;
        let finalization = CredentialFinalization::<Suite>::deserialize(ke3_bytes)
            .map_err(|e| CommonError::new(e.to_string()))?;
        let mut sessions = self.sessions.lock().unwrap();
        if let Some(entry) = sessions.remove(&session_id) {
            if Instant::now() > entry.expires_at {
                return Err(CommonError::new("opaque session expired"));
            }
            ServerLogin::<Suite>::finish(entry.state, finalization)
                .map_err(|e| CommonError::new(format!("opaque login finish: {e}")))?;
            Ok(entry.secret_id)
        } else {
            Err(CommonError::new("opaque session not found"))
        }
    }
}

fn build_manager(
    _server_id: &[u8],
    secret_key: &[u8],
    oprf_seed: &[u8],
    session_ttl_secs: u64,
) -> Result<SafexOpaqueManager, CommonError> {
    let ttl = Duration::from_secs(session_ttl_secs.max(1));
    let setup = build_server_setup(secret_key, oprf_seed)?;
    Ok(SafexOpaqueManager {
        setup,
        session_ttl: ttl,
        sessions: Mutex::new(HashMap::new()),
    })
}

fn build_server_setup(
    secret_key: &[u8],
    oprf_seed: &[u8],
) -> Result<ServerSetup<Suite>, CommonError> {
    let mut rng = OsRng;
    let server_sk = canonicalize_private_key(secret_key);
    let fake_sk = random_fake_private_key(&mut rng, server_sk.len())?;
    let mut serialized = Vec::with_capacity(oprf_seed.len() + server_sk.len() + fake_sk.len());
    serialized.extend_from_slice(oprf_seed);
    serialized.extend_from_slice(&server_sk);
    serialized.extend_from_slice(&fake_sk);
    ServerSetup::deserialize(&serialized)
        .map_err(|e| CommonError::new(format!("init opaque server: {e}")))
}

fn canonicalize_private_key(secret_key: &[u8]) -> Vec<u8> {
    if let Ok(pair) =
        KeyPair::<<Suite as CipherSuite>::KeGroup>::from_private_key_slice(secret_key)
    {
        return pair.private().serialize().to_vec();
    }
    let key_len = secret_key.len();
    let mut seed = Sha512::digest(secret_key).to_vec();
    loop {
        if seed.len() < key_len {
            seed = Sha512::digest(&seed).to_vec();
            continue;
        }
        if let Ok(pair) =
            KeyPair::<<Suite as CipherSuite>::KeGroup>::from_private_key_slice(&seed[..key_len])
        {
            return pair.private().serialize().to_vec();
        }
        seed = Sha512::digest(&seed).to_vec();
    }
}

fn random_fake_private_key(rng: &mut OsRng, key_len: usize) -> Result<Vec<u8>, CommonError> {
    let mut candidate = vec![0u8; key_len];
    loop {
        rng.fill_bytes(&mut candidate);
        if let Ok(pair) =
            KeyPair::<<Suite as CipherSuite>::KeGroup>::from_private_key_slice(&candidate)
        {
            return Ok(pair.private().serialize().to_vec());
        }
    }
}

fn generate_session_id() -> String {
    let mut bytes = [0u8; 16];
    OsRng.fill_bytes(&mut bytes);
    hex_encode(bytes)
}

unsafe fn read_slice<'a>(ptr: *const u8, len: usize) -> Result<&'a [u8], CommonError> {
    if len == 0 {
        return Ok(&[]);
    }
    if ptr.is_null() {
        return Err(CommonError::new("received null pointer argument"));
    }
    Ok(slice::from_raw_parts(ptr, len))
}

fn to_buffer(data: Vec<u8>) -> SafexOpaqueBuffer {
    let boxed = data.into_boxed_slice();
    let len = boxed.len();
    let ptr = Box::into_raw(boxed) as *mut u8;
    SafexOpaqueBuffer { ptr, len }
}

fn set_error(err_out: *mut *mut c_char, err: Option<CommonError>) {
    if err_out.is_null() {
        return;
    }
    unsafe {
        if let Some(err) = err {
            if !(*err_out).is_null() {
                drop(CString::from_raw(*err_out));
            }
            let msg = CString::new(err.to_string())
                .unwrap_or_else(|_| CString::new("opaque error").unwrap());
            *err_out = msg.into_raw();
        } else {
            if !(*err_out).is_null() {
                drop(CString::from_raw(*err_out));
            }
            *err_out = ptr::null_mut();
        }
    }
}
