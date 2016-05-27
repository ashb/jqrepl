// +build ignore

package main

import (
	"github.com/ashb/jqrepl"
	"github.com/ashb/jqrepl/jq"
)

func main() {
	repl, err := jqrepl.New()

	if err != nil {
		// TODO: don't use panic
		panic(err)
	}

	defer repl.Close()

	input, err := jq.JvFromJSONString(`
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

	repl.SetJvInput(input)

	repl.Loop()
}
