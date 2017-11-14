package main

import (
	"io/ioutil"
	"os"

	"github.com/ashb/jqrepl"
	"github.com/ashb/jqrepl/jq"
)

func main() {

	var (
		jv  *jq.Jv
		err error
	)

	repl, err := jqrepl.New()

	if err != nil {
		// TODO: don't use panic
		panic(err)
	}

	defer repl.Close()

	if err != nil {
		// TODO: don't use panic
		panic(err)
	}

	if jqrepl.StdinIsTTY() {
		// TODO: Get input from a file, or exec a command!
		jv, err = jq.JvFromJSONString(`
							{ "simple": 123,
								"nested": {
									"a": [1,2,"a"],
									"b": true,
									"c": null
								},
							"non_printable": "\ud83c\uddec\ud83c\udde7"
							}`)
		if err != nil {
			// TODO: don't use panic
			panic(err)
		}
	} else {

		input, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			// TODO: don't use panic
			panic(err)
		}

		jv, err = jq.JvFromJSONBytes(input)

		if err != nil {
			// TODO: don't use panic
			panic(err)
		}
	}

	repl.SetJvInput(jv)

	repl.Loop()
}
