package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Flag definitions.
var (
	pkg = flag.String("package", "main", "name of the package to generate")
	out = flag.String("out", "embedded_data.go", "file to write to")
)

func readContent(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadAll(f)
	return string(buf), err
}

func escapeContent(c string) string {
	// Escape
	//   `
	// with
	//   ` + "\x60" + `
	// which let us embed it in the template.
	return strings.Replace(c, "`", "` + \"\\x60\" + `", -1)
}

type File struct {
	Name           string
	Path           string
	EscapedContent string
}

const generatedTmpl = `
package {{.Pkg}}

// File auto-generated by embed.go.

{{range .Files}}
// {{.Path}} ----- 8< ----- 8< ----- 8< ----- 8< -----

// {{.Name}} contains the content of {{.Path}}.
const {{.Name}} = ` + "`" + `{{.EscapedContent}}` + "`" + `

{{end}}

`

// tmpl is the compiled version of generatedTmpl above.
var tmpl = template.Must(template.New("generated").Parse(generatedTmpl))

func Embed(pkg string, paths []string, outPath string) error {

	var vals struct {
		Pkg   string
		Files []*File
	}

	vals.Pkg = pkg

	for _, path := range paths {
		f := &File{
			Path: path,
			Name: strings.Replace(filepath.Base(path), ".", "_", -1),
		}

		content, err := readContent(path)
		if err != nil {
			return fmt.Errorf("Error reading %q: %v", path, err)
		}

		f.EscapedContent = escapeContent(content)

		vals.Files = append(vals.Files, f)
	}

	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("Error opening %q: %v", *out, err)
	}

	err = tmpl.Execute(out, vals)
	if err != nil {
		return fmt.Errorf("Error executing template: %v", err)
	}

	return nil
}

func main() {
	flag.Parse()

	paths := []string{}
	for _, glob := range flag.Args() {
		matches, err := filepath.Glob(glob)
		if err != nil {
			fmt.Printf("%v\n", err)
			os.Exit(1)
		}
		paths = append(paths, matches...)
	}

	err := Embed(*pkg, paths, *out)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}