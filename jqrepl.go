package jqrepl

import (
	"fmt"
	"io"
	"os"

	"github.com/ashb/jqrepl/jq"
	"gopkg.in/chzyer/readline.v1"
)

var (
	jvStringName, jvStringValue, jvStringOut, jvStringUnderscore, jvStringDunderscore *jq.Jv
)

func init() {
	// Pre-create thre jv_string objects that we will use over and over again.
	jvStringName = jq.JvFromString("name")
	jvStringValue = jq.JvFromString("value")
	jvStringOut = jq.JvFromString("out")
	jvStringUnderscore = jq.JvFromString("_")
	jvStringDunderscore = jq.JvFromString("__")
}

type JqRepl struct {
	programCounter int
	promptTemplate string
	reader         *readline.Instance
	libJq          *jq.Jq
	input          *jq.Jv

	// a jv array of previous outputs
	results *jq.Jv
}

func StdinIsTTY() bool {
	return readline.IsTerminal(int(os.Stdin.Fd()))
}

// New creates a nwq JqRepl
//
// If stdin is not a tty then it will re-open the controlling tty ("/dev/tty"
// on unix) to be able to run in interactive mode
func New() (*JqRepl, error) {
	repl := JqRepl{
		promptTemplate: "\033[0;36m%3d »\033[0m ",
	}

	cfg, err := repl.readlineReplConfig()
	if err != nil {
		return nil, err
	}

	repl.reader, err = readline.NewEx(cfg)
	if err != nil {
		return nil, err
	}

	repl.libJq, err = jq.New()
	if err != nil {
		repl.reader.Close()
		return nil, err
	}

	repl.results = jq.JvArray()

	return &repl, nil
}

func (repl *JqRepl) readlineReplConfig() (*readline.Config, error) {
	cfg := readline.Config{
		Prompt: repl.currentPrompt(),
		Stdin:  os.Stdin,
	}

	// If stdin is not a tty (i.e. cos we've had input redirected) then we need
	// to re-open the controlling TTY to get interactive input from the user.
	if !StdinIsTTY() {
		tty, err := ReopenTTY()
		if err != nil {
			return nil, err
		}

		fd := int(tty.Fd())
		cfg.Stdin = tty
		cfg.ForceUseInteractive = true

		// The default impl of Make/ExitRaw operate on os.Stdin, which is not
		// re-settable
		var previousState *readline.State
		cfg.FuncMakeRaw = func() error {
			var err error
			previousState, err = readline.MakeRaw(fd)
			return err
		}
		cfg.FuncExitRaw = func() error {
			return readline.Restore(fd, previousState)
		}
	}

	cfg.Init()
	return &cfg, nil
}

func (repl *JqRepl) Close() {
	repl.reader.Close()
	repl.libJq.Close()
	if repl.input != nil {
		repl.input.Free()
	}
}

func (repl *JqRepl) currentPrompt() string {
	return fmt.Sprintf(repl.promptTemplate, repl.programCounter)
}

// JvInput returns the current input the JQ program will operate on
func (repl *JqRepl) JvInput() *jq.Jv {
	return repl.input
}

func (repl *JqRepl) SetJvInput(input *jq.Jv) {
	if repl.input != nil {
		repl.input.Free()
	}
	repl.input = input
}

func (repl *JqRepl) Loop() {
	for {
		repl.reader.SetPrompt(repl.currentPrompt())

		line, err := repl.reader.Readline()
		if err == io.EOF {
			break
		} else if err == readline.ErrInterrupt {
			// Stop the streaming of any results - if we were
			continue
		} else if err != nil {
			panic(fmt.Errorf("%#v", err))
		}

		repl.RunProgram(line)
	}
}

func (repl *JqRepl) Error(err error) {
	fmt.Fprintf(repl.reader.Stderr(), "\033[0;31m%s\033[0m\n", err)
}

func (repl *JqRepl) Output(o *jq.Jv) {
	fmt.Fprintln(repl.reader.Stdout(), o.Dump(jq.JvPrintPretty|jq.JvPrintSpace1|jq.JvPrintColour))
}

func (repl *JqRepl) RunProgram(program string) {
	args := makeProgramArgs(repl.results.Copy())
	chanIn, chanOut, chanErr := repl.libJq.Start(program, args)
	inCopy := repl.JvInput().Copy()

	// Run until the channels are closed
	for chanErr != nil && chanOut != nil {
		select {
		case e, ok := <-chanErr:
			if !ok {
				chanErr = nil
			} else {
				repl.Error(e)

			}
		case o, ok := <-chanOut:
			if !ok {
				chanOut = nil
			} else {
				// Store the result in the history
				repl.results = repl.results.ArrayAppend(o.Copy())

				repl.Output(o)
				repl.programCounter++
			}
		case chanIn <- inCopy:
			// We've sent our input, close the channel to tell Jq we're done
			close(chanIn)
			chanIn = nil
		}
	}
}

func makeProgramArgs(history *jq.Jv) *jq.Jv {
	// Create this structure:
	// programArgs = [
	//		{"name": "out", "value": history },
	//    {"name": "_", "value": history[-1] },
	//    {"name": "__", "value": history[-2] },
	// ]

	arg := jq.JvObject().ObjectSet(jvStringName.Copy(), jvStringOut.Copy()).ObjectSet(jvStringValue.Copy(), history.Copy())
	res := jq.JvArray().ArrayAppend(arg)

	len := history.Copy().ArrayLength()

	if len >= 1 {
		arg = jq.JvObject().ObjectSet(jvStringName.Copy(), jvStringUnderscore.Copy()).ObjectSet(jvStringValue.Copy(), history.Copy().ArrayGet(len-1))
		res = res.ArrayAppend(arg)
	}
	if len >= 2 {
		arg = jq.JvObject().ObjectSet(jvStringName.Copy(), jvStringDunderscore.Copy()).ObjectSet(jvStringValue.Copy(), history.Copy().ArrayGet(len-2))
		res = res.ArrayAppend(arg)
	}

	return res
}
