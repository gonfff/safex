#pragma once

#include <stdint.h>
#include <stddef.h>

#ifdef __cplusplus
extern "C" {
#endif

typedef struct SafexOpaqueManager SafexOpaqueManager;

typedef struct {
    uint8_t *ptr;
    size_t len;
} SafexOpaqueBuffer;

typedef struct {
    SafexOpaqueBuffer session_id;
    SafexOpaqueBuffer response;
} SafexOpaqueLoginStart;

typedef struct {
    uint32_t handle;
    SafexOpaqueBuffer message;
} SafexOpaqueClientStart;

typedef struct {
    SafexOpaqueBuffer upload;
    SafexOpaqueBuffer export_key;
} SafexOpaqueRegistrationFinish;

typedef struct {
    SafexOpaqueBuffer finalization;
    SafexOpaqueBuffer export_key;
    SafexOpaqueBuffer session_key;
} SafexOpaqueLoginFinish;

SafexOpaqueManager *safex_opaque_manager_new(
    const uint8_t *server_id,
    size_t server_id_len,
    const uint8_t *secret_key,
    size_t secret_key_len,
    const uint8_t *oprf_seed,
    size_t oprf_seed_len,
    uint64_t session_ttl_secs,
    char **err_out);

void safex_opaque_manager_free(SafexOpaqueManager *manager);

SafexOpaqueBuffer safex_opaque_registration_response(
    SafexOpaqueManager *manager,
    const uint8_t *secret_id,
    size_t secret_id_len,
    const uint8_t *request,
    size_t request_len,
    char **err_out);

SafexOpaqueLoginStart safex_opaque_login_start(
    SafexOpaqueManager *manager,
    const uint8_t *secret_id,
    size_t secret_id_len,
    const uint8_t *record_blob,
    size_t record_len,
    const uint8_t *request,
    size_t request_len,
    char **err_out);

SafexOpaqueBuffer safex_opaque_login_finish(
    SafexOpaqueManager *manager,
    const uint8_t *session_id,
    size_t session_id_len,
    const uint8_t *ke3_payload,
    size_t ke3_len,
    char **err_out);

SafexOpaqueClientStart safex_opaque_client_start_registration(
    const uint8_t *pin,
    size_t pin_len,
    char **err_out);

SafexOpaqueRegistrationFinish safex_opaque_client_finish_registration(
    uint32_t handle,
    const uint8_t *pin,
    size_t pin_len,
    const uint8_t *response,
    size_t response_len,
    char **err_out);

SafexOpaqueClientStart safex_opaque_client_start_login(
    const uint8_t *pin,
    size_t pin_len,
    char **err_out);

SafexOpaqueLoginFinish safex_opaque_client_finish_login(
    uint32_t handle,
    const uint8_t *pin,
    size_t pin_len,
    const uint8_t *response,
    size_t response_len,
    char **err_out);

void safex_opaque_buffer_free(SafexOpaqueBuffer buffer);
void safex_opaque_string_free(char *err_ptr);

#ifdef __cplusplus
}
#endif
