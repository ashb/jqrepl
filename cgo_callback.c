#include <stdio.h>

#include <jv.h>
#include <jq.h>

extern void go_error_handler(void*, jv);

static inline void Call_go_error_handler(void *data, jv it) {
	go_error_handler(data, it);
}

void install_jq_error_cb(jq_state *jq, void* go_jq) {
	jq_set_error_cb(jq, Call_go_error_handler, go_jq);
}
