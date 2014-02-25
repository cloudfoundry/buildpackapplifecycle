package candiedyaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
)

var parses = func(filename string) {
	It("parses "+filename, func() {
		file, err := os.Open(filename)
		Ω(err).To(BeNil())

		parser := yaml_parser_t{}
		yaml_parser_initialize(&parser)
		yaml_parser_set_input_reader(&parser, file)

		failed := false
		event := yaml_event_t{}

		for {
			if !yaml_parser_parse(&parser, &event) {
				failed = true
				println("---", parser.error, parser.problem, parser.context, "line", parser.problem_mark.line, "col", parser.problem_mark.column)
				break
			}

			if event.event_type == yaml_STREAM_END_EVENT {
				break
			}
		}

		file.Close()

		// msg := "SUCCESS"
		// if failed {
		// 	msg = "FAILED"
		// 	if parser.error != yaml_NO_ERROR {
		// 		m := parser.problem_mark
		// 		fmt.Printf("ERROR: (%s) %s @ line: %d  col: %d\n",
		// 			parser.context, parser.problem, m.line, m.column)
		// 	}
		// }
		Ω(failed).To(BeFalse())
	})
}

var parseYamls = func(dirname string) {
	fileInfos, err := ioutil.ReadDir(dirname)
	Ω(err).To(BeNil())
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {
			parses(filepath.Join(dirname, fileInfo.Name()))
		}
	}
}

var _ = Describe("Parser", func() {
	parseYamls("fixtures/specification")
	parseYamls("fixtures/specification/types")
})
