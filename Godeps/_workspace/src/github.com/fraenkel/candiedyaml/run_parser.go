package candiedyaml

import (
	"fmt"
	"os"
)

func Run_parser(cmd string, args []string) {
	for i := 0; i < len(args); i++ {
		fmt.Printf("[%d] Scanning '%s'", i, args[i])
		file, err := os.Open(args[i])
		if err != nil {
			panic(fmt.Sprintf("Invalid file '%s': %s", args[i], err.Error()))
		}

		parser := yaml_parser_t{}
		yaml_parser_initialize(&parser)
		yaml_parser_set_input_reader(&parser, file)

		failed := false
		token := yaml_token_t{}
		count := 0
		for {
			if !yaml_parser_scan(&parser, &token) {
				failed = true
				break
			}

			if token.token_type == yaml_STREAM_END_TOKEN {
				break
			}
			count++
		}

		file.Close()

		msg := "SUCCESS"
		if failed {
			msg = "FAILED"
			if parser.error != yaml_NO_ERROR {
				m := parser.problem_mark
				fmt.Printf("ERROR: (%s) %s @ line: %d  col: %d\n",
					parser.context, parser.problem, m.line, m.column)
			}
		}
		fmt.Printf("%s (%d tokens)\n", msg, count)
	}
}
