use aes_gcm::aead::{Aead, KeyInit, Nonce as AeadNonce};
use aes_gcm::Aes256Gcm;
use argon2::Argon2;
use rand::rngs::OsRng;
use rand::TryRngCore;
use wasm_bindgen::prelude::*;

const SALT_LEN: usize = 16;
const NONCE_LEN: usize = 12;

type CipherNonce = AeadNonce<Aes256Gcm>;

#[wasm_bindgen]
pub fn encrypt(code: &str, data: &[u8]) -> Result<Vec<u8>, JsValue> {
    validate_code(code)?;
    let mut salt = [0u8; SALT_LEN];
    let mut nonce = [0u8; NONCE_LEN];
    let mut rng = OsRng;
    rng.try_fill_bytes(&mut salt)
        .expect("os rng should always be available");
    rng.try_fill_bytes(&mut nonce)
        .expect("os rng should always be available");
    let key = derive_key(code, &salt);
    let nonce_ga = CipherNonce::from(nonce);
    let cipher = Aes256Gcm::new_from_slice(&key).map_err(to_js_err)?;
    let ciphertext = cipher.encrypt(&nonce_ga, data).map_err(to_js_err)?;
    let mut payload = Vec::with_capacity(SALT_LEN + NONCE_LEN + ciphertext.len());
    payload.extend_from_slice(&salt);
    payload.extend_from_slice(&nonce);
    payload.extend_from_slice(&ciphertext);
    Ok(payload)
}

#[wasm_bindgen]
pub fn decrypt(code: &str, payload: &[u8]) -> Result<Vec<u8>, JsValue> {
    validate_code(code)?;
    if payload.len() < SALT_LEN + NONCE_LEN {
        return Err(JsValue::from_str("payload too small"));
    }
    let (salt, rest) = payload.split_at(SALT_LEN);
    let (nonce_bytes, ciphertext) = rest.split_at(NONCE_LEN);
    let key = derive_key(code, salt);
    let mut nonce = [0u8; NONCE_LEN];
    nonce.copy_from_slice(nonce_bytes);
    let nonce_ga = CipherNonce::from(nonce);
    let cipher = Aes256Gcm::new_from_slice(&key).map_err(to_js_err)?;
    cipher.decrypt(&nonce_ga, ciphertext).map_err(to_js_err)
}

fn derive_key(code: &str, salt: &[u8]) -> [u8; 32] {
    let mut key = [0u8; 32];
    Argon2::default()
        .hash_password_into(code.as_bytes(), salt, &mut key)
        .expect("argon2 key derivation failure");
    key
}

fn validate_code(code: &str) -> Result<(), JsValue> {
    if code.len() != 6 || !code.chars().all(|c| c.is_ascii_alphanumeric()) {
        return Err(JsValue::from_str("code must be 6 alphanumeric characters"));
    }
    Ok(())
}

fn to_js_err<E: core::fmt::Display>(err: E) -> JsValue {
    JsValue::from_str(&err.to_string())
}

#[cfg(all(test, target_arch = "wasm32"))]
mod tests {
    use super::*;
    use wasm_bindgen_test::*;

    #[wasm_bindgen_test]
    fn roundtrip() {
        let code = "123456";
        let plain = b"hello world";
        let payload = encrypt(code, plain).expect("encrypt");
        let restored = decrypt(code, &payload).expect("decrypt");
        assert_eq!(plain.to_vec(), restored);
    }

    #[wasm_bindgen_test]
    fn accepts_alphanumeric() {
        let code = "A1B2C3";
        let payload = encrypt(code, b"hi").expect("encrypt");
        let restored = decrypt(code, &payload).expect("decrypt");
        assert_eq!(restored, b"hi");
    }

    #[wasm_bindgen_test]
    fn bad_code() {
        let plain = b"hi";
        assert!(encrypt("12", plain).is_err());
    }
}
