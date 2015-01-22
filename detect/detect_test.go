package detect

import (
	"fmt"
	"testing"
)

func TestDetect(t *testing.T) {
	cases := map[string]string{
		"hello.c":      "c",
		"hello.coffee": "coffeescript",
		"hello.go":     "go",
		"hello.hs":     "haskell",
		"hello.idr":    "idris",
		"hello.jl":     "julia",
		"hello.rb":     "ruby",
		"hello.rs":     "rust",
	}

	for file, type_ := range cases {
		p, err := Detect(fmt.Sprintf("../examples/%s", file))
		if err != nil {
			t.Errorf("Could not detect project type of '%s'", file)
		} else {
			if p.Id != type_ {
				t.Errorf("Expected type of '%s' to be '%s', but detected '%s'", file, type_, p.Id)
			}
		}
	}
}
