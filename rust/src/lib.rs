pub mod common;

#[cfg(feature = "wasm")]
pub mod wasm;

#[cfg(feature = "wasm")]
pub use wasm::*;

#[cfg(feature = "server")]
pub mod server;

pub use common::{decrypt_payload, derive_aes_key, encrypt_payload, CommonError, Suite, NONCE_LEN};
