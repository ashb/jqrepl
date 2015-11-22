package jq

/*
To install
$ ./configure --disable-maintainer-mode --prefix=$PWD/BUILD
$ make install-libLTLIBRARIES install-includeHEADERS
*/

/*
#cgo CFLAGS: -I ${SRCDIR}/../jq-1.5/BUILD/include
#cgo LDFLAGS: ${SRCDIR}/../jq-1.5/BUILD/lib/libjq.a

#include <jq.h>
#include <jv.h>

#include <stdlib.h>

void install_jq_error_cb(jq_state *jq, void* go_jq);
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// The object representing the complete JQ state.
type Jq struct {
	_state       *C.struct_jq_state
	errorChannel chan error
	input        *Jv
}

// Create a new JQ object. errorChannel will be sent any "recoverable errorrs"
// - i.e. ones caused by invalid input or invalid programs, but not out of
// memory situations
func New(errorChannel chan error) (*Jq, error) {
	jq := new(Jq)

	var err error
	jq._state, err = C.jq_init()

	if err != nil {
		return nil, err
	} else if jq == nil {
		return nil, errors.New("jq_init returned nil -- out of memory?")
	}

	jq.errorChannel = errorChannel

	// Because we can't pass a function pointer to an exported Go func we have to
	// call a C function which uses the exported fund for us.
	// https://github.com/golang/go/wiki/cgo#function-variables
	C.install_jq_error_cb(jq._state, unsafe.Pointer(jq))

	return jq, nil
}

func (jq *Jq) Close() {
	if jq.input != nil {
		jq.input.Free()
		jq.input = nil
	}
	if jq._state != nil {
		C.jq_teardown(&jq._state)
		close(jq.errorChannel)
		jq._state = nil
	}
}

//export go_error_handler
func go_error_handler(data unsafe.Pointer, jv C.jv) {
	jq := (*Jq)(data)

	jq.errorChannel <- _ConvertError(jv)
}

func (jq *Jq) SetJsonInput(input string) error {
	if jq.input != nil {
		jq.input.Free()
	}

	var err error
	jq.input, err = JvFromJsonString(input)
	return err
}

// Execute program against the provided input. On error will return a
// placeholder error with the real error(s) sent to the errorChannel provided
// to New. Caller is responsible for freeing the result
func (jq *Jq) Execute(program string) (*Jv, error) {

	if jq._Compile(program) == false {
		return nil, errors.New("JQ compile errors sent to channel")
	}

	if jq.input == nil {
		return nil, errors.New("SetJsonInput must be called before calling Execute")
	}

	// TODO: Make this use channels, as JQ by defualt deals with multiple JSON
	// documents as input and filters each in turn.

	flags := C.int(0)

	C.jq_start(jq._state, C.jv_copy(jq.input.jv), flags)
	result := &Jv{C.jq_next(jq._state)}

	if C.jv_is_valid(result.jv) == 1 {
		return result, nil
	} else {
		// TODO: Why am I copying here
		msg := &Jv{C.jv_invalid_get_msg(C.jv_copy(result.jv))}

		//input_pos := C.jq_util_input_get_position(jq)
		if msg.Kind() == JV_KIND_STRING {
			defer msg.Free()
			return nil, errors.New(msg._string())
		} else {
			msg := &Jv{C.jv_dump_string(msg.jv, 0)}
			defer msg.Free()
			return nil, fmt.Errorf("(not a string): %s", msg._string())
		}
		//jv_free(input_pos)
	}
}

func (jq *Jq) _Compile(prog string) bool {
	cs := C.CString(prog)
	defer C.free(unsafe.Pointer(cs))

	compiled := C.jq_compile(jq._state, cs) != 0
	// If there was an error it will have been sent to errorChannel
	return compiled
}
