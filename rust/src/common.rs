use aes_gcm::aead::{Aead, Error as AeadError, KeyInit, Nonce as AeadNonce};
use aes_gcm::Aes256Gcm;
use argon2::Argon2;
use hkdf::Hkdf;
use opaque_ke::ciphersuite::CipherSuite;
use opaque_ke::key_exchange::tripledh::TripleDh;
use opaque_ke::Ristretto255;
use rand::rngs::OsRng;
use rand::RngCore;
use sha2::{digest::InvalidLength, Sha512};
use std::fmt;

pub const NONCE_LEN: usize = 12;
pub const AES_KEY_LEN: usize = 32;
const HKDF_SALT: &[u8] = b"safex/opaque/export-key";
const HKDF_INFO: &[u8] = b"safex/aes256-gcm";

pub type CipherNonce = AeadNonce<Aes256Gcm>;
pub type Suite = SafexCipherSuite;

pub struct SafexCipherSuite;

impl CipherSuite for SafexCipherSuite {
    type OprfCs = Ristretto255;
    type KeGroup = Ristretto255;
    type KeyExchange = TripleDh;
    type Ksf = Argon2<'static>;
}

pub type CommonResult<T> = Result<T, CommonError>;

#[derive(Debug)]
pub struct CommonError(String);

impl CommonError {
    pub fn new(msg: impl Into<String>) -> Self {
        Self(msg.into())
    }
}

impl fmt::Display for CommonError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl std::error::Error for CommonError {}

impl From<AeadError> for CommonError {
    fn from(err: AeadError) -> Self {
        CommonError::new(err.to_string())
    }
}

impl From<InvalidLength> for CommonError {
    fn from(err: InvalidLength) -> Self {
        CommonError::new(err.to_string())
    }
}

pub fn derive_aes_key(export_key: &[u8]) -> CommonResult<[u8; AES_KEY_LEN]> {
    if export_key.is_empty() {
        return Err(CommonError::new("export_key must not be empty"));
    }
    let hkdf = Hkdf::<Sha512>::new(Some(HKDF_SALT), export_key);
    let mut key = [0u8; AES_KEY_LEN];
    hkdf.expand(HKDF_INFO, &mut key)
        .map_err(|_| CommonError::new("unable to derive key from export_key"))?;
    Ok(key)
}

pub fn encrypt_payload(export_key: &[u8], data: &[u8]) -> CommonResult<Vec<u8>> {
    let key = derive_aes_key(export_key)?;
    let mut nonce = [0u8; NONCE_LEN];
    let mut rng = OsRng;
    rng.fill_bytes(&mut nonce);
    let cipher = Aes256Gcm::new_from_slice(&key)?;
    let nonce_ga = CipherNonce::from(nonce);
    let ciphertext = cipher.encrypt(&nonce_ga, data)?;
    let mut payload = Vec::with_capacity(NONCE_LEN + ciphertext.len());
    payload.extend_from_slice(&nonce);
    payload.extend_from_slice(&ciphertext);
    Ok(payload)
}

pub fn decrypt_payload(export_key: &[u8], payload: &[u8]) -> CommonResult<Vec<u8>> {
    if payload.len() < NONCE_LEN {
        return Err(CommonError::new("payload too small"));
    }
    let key = derive_aes_key(export_key)?;
    let (nonce_bytes, ciphertext) = payload.split_at(NONCE_LEN);
    let mut nonce = [0u8; NONCE_LEN];
    nonce.copy_from_slice(nonce_bytes);
    let cipher = Aes256Gcm::new_from_slice(&key)?;
    let nonce_ga = CipherNonce::from(nonce);
    Ok(cipher.decrypt(&nonce_ga, ciphertext)?)
}
