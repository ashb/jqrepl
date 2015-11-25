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
	"unsafe"
)

// The object representing the complete JQ state.
type Jq struct {
	_state       *C.struct_jq_state
	errorChannel chan error
}

// Create a new JQ object. errorChannel will be sent any "recoverable errorrs"
// - i.e. ones caused by invalid input or invalid programs, but not out of
// memory situations
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
	if jq._state != nil {
		C.jq_teardown(&jq._state)
		jq._state = nil
	}
}

//export go_error_handler
func go_error_handler(data unsafe.Pointer, jv C.jv) {
	ch := *(*chan<- error)(data)

	err := _ConvertError(jv)
	ch <- err
}

// Start will compile `program` and return a three channels: input, output and
// error. Sending a jq.Jv* to input cause the program to be run to it and
// results returned as jq.Jv* on the output channel, or one or more error
// values sent to the error channel. When you are done sending values close the
// input channel.
//
// This function is not reentereant -- in that you cannot and should not call
// Start again until you have closed the previous input channel.
//
// If there is a problem compiling the JQ program then the errors will be
// reported on error channel before any input is read so makle sure you account
// for this case.
//
// Any jq.Jv* values passed to the input channel will be owned by the channel.
// If you want to keep them afterwards ensure you Copy() them before passing to
// the channel
func (jq *Jq) Start(program string) (in chan<- *Jv, out <-chan *Jv, errs <-chan error) {

	// Create out two way copy of the channels. We need to be able to recv from
	// input, so need to store the original channel
	cIn := make(chan *Jv)
	cOut := make(chan *Jv)
	cErr := make(chan error)

	// And assign the read/write only versions to the output fars
	in = cIn
	out = cOut
	errs = cErr

	// Because we can't pass a function pointer to an exported Go func we have to
	// call a C function which uses the exported fund for us.
	// https://github.com/golang/go/wiki/cgo#function-variables
	C.install_jq_error_cb(jq._state, unsafe.Pointer(&cErr))

	go func() {
		if jq._Compile(program) == false {
			// Even if compile failed follow the contract. Read any inputs and take
			// ownership of them (aka free them)
			//
			// Errors from compile will be sent to the error channel
			for jv := range cIn {
				jv.Free()
			}
		} else {
			for jv := range cIn {
				jq._Execute(jv, cOut, jq.errorChannel)
			}
		}
		// Once we've read all the inputs close the output to signal to caller that
		// we are done.
		close(cOut)
		close(cErr)
		C.install_jq_error_cb(jq._state, nil)
	}()

	return
}

// Process a single input and send the results on `out`
func (jq *Jq) _Execute(jv *Jv, out chan<- *Jv, err chan<- error) {
	flags := C.int(0)

	C.jq_start(jq._state, jv.jv, flags)
	result := &Jv{C.jq_next(jq._state)}
	for result.IsValid() {
		out <- result
		result = &Jv{C.jq_next(jq._state)}
	}
	if msg := result.GetInvalidMessage(); msg.Kind() != JV_KIND_NULL {
		// Uncaught jq exception
		// TODO: get file:line position in input somehow.
		if msg.Kind() == JV_KIND_STRING {
			defer msg.Free()
			err <- errors.New(msg._string())
		} else {
			msg := Jv{C.jv_dump_string(msg.jv, 0)}
			defer msg.Free()
			err <- errors.New(msg._string())
		}
	}
}

func (jq *Jq) _Compile(prog string) bool {
	cs := C.CString(prog)
	defer C.free(unsafe.Pointer(cs))

	compiled := C.jq_compile(jq._state, cs) != 0
	// If there was an error it will have been sent to errorChannel
	return compiled
}
