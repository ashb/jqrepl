package main

import (
	"fmt"
	"io"

	"github.com/ashb/jq-repl/jq"

	"gopkg.in/chzyer/readline.v1"
)

func main() {
	prompt := "\033[0;36m»\033[0m "
	l, err := readline.New(prompt)

	if err != nil {
		// TODO: don't use panic
		panic(err)
	}

	libjq, err := jq.New()
	if err != nil {
		// TODO: don't use panic
		panic(err)
	}
	defer libjq.Close()

	input, err := jq.JvFromJSONString(`
		{ "simple": 123,
			"nested": {
				"a": [1,2,"a"],
				"b": true,
				"c": null
			}
		}`)
	if err != nil {
		// TODO: don't use panic
		panic(err)
	}

	defer input.Free()

	for {
		line, err := l.Readline()
		if err == io.EOF {
			break
		} else if err == readline.ErrInterrupt {
			// Stop the streaming.
			continue
		} else if err != nil {
			panic(fmt.Errorf("%#v", err))
		}

		chanIn, chanOut, chanErr := libjq.Start(line)
		inCopy := input.Copy()

		// Run until the channels are closed
		for chanErr != nil && chanOut != nil {
			select {
			case e, ok := <-chanErr:
				if !ok {
					chanErr = nil
				} else {
					fmt.Printf("\033[0;31m%s\033[0m\n", e)
				}
			case o, ok := <-chanOut:
				if !ok {
					chanOut = nil
				} else {
					fmt.Printf("// Read from output %v %#v\n", o.Kind(), ok)
					fmt.Println(o.ToGoVal())
				}
			case chanIn <- inCopy:
				// We've sent our input, close the channel to tell Jq we're done
				close(chanIn)
				chanIn = nil
			}
		}
	}

	libjq.Close()
}
