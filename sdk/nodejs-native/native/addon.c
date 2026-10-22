#include <node_api.h>
#include <stdlib.h>
#include "libcodex.h"

static napi_value call(napi_env env, napi_callback_info info) {
 size_t argc = 1, size = 0;
 napi_value argv[1], result;
 if (napi_get_cb_info(env, info, &argc, argv, NULL, NULL) != napi_ok || argc != 1 ||
     napi_get_value_string_utf8(env, argv[0], NULL, 0, &size) != napi_ok) {
  napi_throw_type_error(env, NULL, "call expects a JSON string"); return NULL;
 }
 char *input = malloc(size + 1);
 if (!input) { napi_throw_error(env, NULL, "allocation failed"); return NULL; }
 if (napi_get_value_string_utf8(env, argv[0], input, size + 1, &size) != napi_ok) {
  free(input); napi_throw_error(env, NULL, "invalid input"); return NULL;
 }
 char *output = CodexCall(input);
 free(input);
 napi_status status = napi_create_string_utf8(env, output, NAPI_AUTO_LENGTH, &result);
 CodexFree(output);
 if (status != napi_ok) { napi_throw_error(env, NULL, "invalid response"); return NULL; }
 return result;
}
static napi_value init(napi_env env, napi_value exports) {
 napi_value fn;
 if (napi_create_function(env, "call", NAPI_AUTO_LENGTH, call, NULL, &fn) != napi_ok ||
     napi_set_named_property(env, exports, "call", fn) != napi_ok) return NULL;
 return exports;
}
NAPI_MODULE(codex_native, init)
