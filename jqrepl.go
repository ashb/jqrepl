package jqrepl

import (
	"fmt"
	"io"

	"github.com/ashb/jqrepl/jq"
	"gopkg.in/chzyer/readline.v1"
)

type JqRepl struct {
	programCounter int
	promptTemplate string
	reader         *readline.Instance
	libJq          *jq.Jq
	input          *jq.Jv
}

func New() (*JqRepl, error) {
	repl := JqRepl{
		promptTemplate: "\033[0;36m%3d »\033[0m",
	}
	var err error
	repl.reader, err = readline.New(repl.currentPrompt())
	if err != nil {
		return nil, err
	}

	repl.libJq, err = jq.New()
	if err != nil {
		repl.reader.Close()
		return nil, err
	}

	return &repl, nil
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

		repl.programCounter++
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
	chanIn, chanOut, chanErr := repl.libJq.Start(program)
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
				repl.Output(o)
			}
		case chanIn <- inCopy:
			// We've sent our input, close the channel to tell Jq we're done
			close(chanIn)
			chanIn = nil
		}
	}
}
