package main

import (
	"fmt"
	"log"
	"os"

	"github.com/piotrkowalczuk/pqt/pqtgo"
	"github.com/piotrkowalczuk/pqt/pqtsql"
)

const version = 9.5

var (
	acronyms = map[string]string{
		"id":   "ID",
		"http": "HTTP",
		"ip":   "IP",
		"net":  "NET",
		"irc":  "IRC",
		"uuid": "UUID",
		"url":  "URL",
		"html": "HTML",
		"db":   "DB",
	}
)

func main() {
	file, err := os.Create("schema.pqt.go")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// For simplicity it is going to be empty.
	sch := schema("example")
	fmt.Fprint(file, `
        // Code generated by pqt.
        // source: cmd/appg/main.go
        // DO NOT EDIT!
    `)
	goGen := pqtgo.Generator{
		Formatter: &pqtgo.Formatter{
			Visibility: pqtgo.Public,
			Acronyms:   acronyms,
		},
		Pkg:     "model",
		Version: version,
		Plugins: []pqtgo.Plugin{
			&generator{},
		},
	}
	sqlGen := &pqtsql.Generator{Version: version}

	err = goGen.GenerateTo(file, sch)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(file, "/// SQL ...\n")
	fmt.Fprint(file, "const SQL = `\n")

	if err := sqlGen.GenerateTo(sch, file); err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(file, "`")
}
