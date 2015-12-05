#include <stdio.h>

#include <jv.h>
#include <jq.h>

extern void goLibjqErrorHandler(void*, jv);

static inline void callGoErrorHandler(void *data, jv it) {
	goLibjqErrorHandler(data, it);
}

void install_jq_error_cb(jq_state *jq, void* go_jq) {
	jq_set_error_cb(jq, callGoErrorHandler, go_jq);
}
