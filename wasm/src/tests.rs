use super::*;
use opaque_ke::{
    CredentialFinalization, CredentialRequest, RegistrationRequest, RegistrationUpload,
    ServerLogin, ServerLoginStartParameters, ServerRegistration, ServerSetup,
};
use wasm_bindgen_test::*;

fn assert_err_contains<T>(result: Result<T, JsValue>, needle: &str) {
    let err = match result {
        Ok(_) => panic!("expected Err"),
        Err(err) => err,
    };
    let msg = err
        .as_string()
        .unwrap_or_else(|| format!("{err:?}"));
    assert!(
        msg.contains(needle),
        "expected error containing `{needle}`, got `{msg}`"
    );
}

#[wasm_bindgen_test]
fn opaque_flow_encrypts_payload() {
    let pin = "A1B2C3";
    let mut rng = OsRng;
    let server_setup = ServerSetup::<Suite>::new(&mut rng);
    let user_id = b"user-123";

    let reg_start = start_registration(pin).expect("client start registration");
    let reg_request = RegistrationRequest::<Suite>::deserialize(&reg_start.message()).expect("req");
    let server_reg_start = ServerRegistration::<Suite>::start(&server_setup, reg_request, user_id)
        .expect("server reg start");

    let reg_finish = finish_registration(
        reg_start.handle(),
        pin,
        &server_reg_start.message.serialize().to_vec(),
    )
    .expect("client finish registration");
    let server_record = ServerRegistration::<Suite>::finish(
        RegistrationUpload::<Suite>::deserialize(&reg_finish.upload()).expect("upload"),
    );

    let login_start = start_login(pin).expect("client login start");
    let credential_request =
        CredentialRequest::<Suite>::deserialize(&login_start.message()).expect("cred req");
    let server_login_start = ServerLogin::<Suite>::start(
        &mut rng,
        &server_setup,
        Some(server_record),
        credential_request,
        user_id,
        ServerLoginStartParameters::default(),
    )
    .expect("server login start");

    let login_finish = finish_login(
        login_start.handle(),
        pin,
        &server_login_start.message.serialize().to_vec(),
    )
    .expect("client login finish");
    let finalization = CredentialFinalization::<Suite>::deserialize(&login_finish.finalization())
        .expect("finalization bytes");
    ServerLogin::finish(server_login_start.state, finalization).expect("server finish");

    let payload = b"super secret data";
    let ciphertext = encrypt(&login_finish.export_key(), payload).expect("encrypt");
    let restored = decrypt(&login_finish.export_key(), &ciphertext).expect("decrypt");
    assert_eq!(payload.to_vec(), restored);
    assert_eq!(login_finish.export_key().len(), 64);
}

#[wasm_bindgen_test]
fn finish_registration_rejects_unknown_handle() {
    let pin = "123456";
    let reg_start = start_registration(pin).expect("start registration");
    let bad_handle = reg_start.handle().wrapping_add(42);
    assert_err_contains(
        finish_registration(bad_handle, pin, &[]),
        "unknown registration handle",
    );
    let _ = finish_registration(reg_start.handle(), pin, &[]);
}

#[wasm_bindgen_test]
fn finish_login_rejects_unknown_handle() {
    let pin = "654321";
    let login_start = start_login(pin).expect("start login");
    let bad_handle = login_start.handle().wrapping_add(1);
    assert_err_contains(
        finish_login(bad_handle, pin, &[]),
        "unknown login handle",
    );
    let _ = finish_login(login_start.handle(), pin, &[]);
}

#[wasm_bindgen_test]
fn encrypt_rejects_empty_export_key() {
    assert_err_contains(encrypt(&[], b"payload"), "export_key must not be empty");
}

#[wasm_bindgen_test]
fn decrypt_rejects_short_payload() {
    let export_key = [0u8; 64];
    let short_payload = vec![0u8; NONCE_LEN - 1];
    assert_err_contains(decrypt(&export_key, &short_payload), "payload too small");
}

#[wasm_bindgen_test]
fn decrypt_fails_on_modified_ciphertext() {
    let export_key = [7u8; 64];
    let mut payload = encrypt(&export_key, b"secret bytes").expect("encrypt payload");
    payload[NONCE_LEN] ^= 0xFF;
    assert_err_contains(decrypt(&export_key, &payload), "aead::Error");
}
