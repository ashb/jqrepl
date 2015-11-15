package jqrepl

/*
To install
$ ./configure --disable-maintainer-mode --prefix=$PWD/BUILD
$ make install-libLTLIBRARIES install-includeHEADERS
*/

/*
#cgo CFLAGS: -I ${SRCDIR}/jq-1.5/BUILD/include
#cgo LDFLAGS: ${SRCDIR}/jq-1.5/BUILD/lib/libjq.a

#include <jq.h>
#include <jv.h>
*/
import "C"
import "errors"

type Jq struct {
	_state *C.struct_jq_state
}

func New() (*Jq, error) {
	jq := new(Jq)

	var err error
	jq._state, err = C.jq_init()

	if err != nil {
		return nil, err
	} else if jq == nil {
		return nil, errors.New("jq_init returned nil -- out of memory?")
	}

	return jq, nil
}

func (jq *Jq) Close() {
	C.jq_teardown(&jq._state)
}
