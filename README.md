# pqt [![GoDoc](https://godoc.org/github.com/piotrkowalczuk/pqt?status.svg)](http://godoc.org/github.com/piotrkowalczuk/pqt)&nbsp;[![Build Status](https://travis-ci.org/piotrkowalczuk/pqt.svg)](https://travis-ci.org/piotrkowalczuk/pqt)&nbsp;[![codecov.io](https://codecov.io/github/piotrkowalczuk/pqt/coverage.svg?branch=master)](https://codecov.io/github/piotrkowalczuk/pqt?branch=master)
This package is a toolbox for postgres driven applications.
It provides multiple tools to help to work with postgres efficiently.
In comparison to other currently available libraries instead of pushing struct tags into anti pattern or parsing SQL it allow to define schema programmatically.

It relies to a large degree on packages:

* [ntypes](http://github.com/piotrkowalczuk/ntypes)
* [qtypes](http://github.com/piotrkowalczuk/qtypes)

## Features:

- __array support__ - golang postgres driver do not support arrays natively, pqt comes with help:
	- [pqt.ArrayInt64](https://godoc.org/github.com/piotrkowalczuk/pqt#ArrayInt64) - wrapper for []int64
	- [pqt.ArrayFloat64](https://godoc.org/github.com/piotrkowalczuk/pqt#ArrayFloat64) - wrapper for []float64
	- [pqt.ArrayString](https://godoc.org/github.com/piotrkowalczuk/pqt#ArrayString) - wrapper for []string
- __sql generation__
- __go generation__ - it includes:
	- `entity` - struct that reflects single row within the database
	- `criteria` - object that can be passed to the `Find` method, it allows to create complex queries
	- `patch` - structure used by `UpdateBy<primary-key>` methods to modify existing entity
	- `constants`:
		- `table names`
		- `column names`
		- `constraints` - library generates exact names of each constraint and corresponding constant that allow to easily handle query errors using `pqt.ErrorConstraint` helper function
	- `repository` - data access layer that expose API to manipulate entities:
		- `Find` - returns collection of entities that match given criteria
		- `Insert` - saves given entity into the database
		- `FindOneBy<primary-key>` - retrieves single entity
		- `UpdateBy<primary-key>` - modifies single entity
		- `DeleteBy<primary-key>` - modifies single entity
- __schema definition__ - allow to programmatically define database schema, that includes:
	- `schemas`
	- `tables`
	- `columns`
	- `constraints`
	- `relationships`


## Example

### Table
```go
s := pqt.Schema("custom")
u := pqt.NewTable("user", pqt.WithIfNotExists()).
	AddColumn(pqt.NewColumn("password", pqt.TypeBytea(), pqt.WithNotNull())).
	AddColumn(pqt.NewColumn("username", pqt.TypeText(), pqt.WithNotNull(), pqt.WithUnique())).
	AddColumn(pqt.NewColumn("first_name", pqt.TypeText(), pqt.WithNotNull())).
	AddColumn(pqt.NewColumn("last_name", pqt.TypeText(), pqt.WithNotNull())).
	AddColumn(pqt.NewColumn("is_superuser", pqt.TypeBool(), pqt.WithNotNull(), pqt.WithDefault("FALSE"))).
	AddColumn(pqt.NewColumn("is_active", pqt.TypeBool(), pqt.WithNotNull(), pqt.WithDefault("FALSE"))).
	AddColumn(pqt.NewColumn("is_staff", pqt.TypeBool(), pqt.WithNotNull(), pqt.WithDefault("FALSE"))).
	AddColumn(pqt.NewColumn("is_confirmed", pqt.TypeBool(), pqt.WithNotNull(), pqt.WithDefault("FALSE"))).
	AddColumn(pqt.NewColumn("confirmation_token", pqt.TypeBytea())).
	AddColumn(pqt.NewColumn("last_login_at", pqt.TypeTimestampTZ()))
s.AddTable(u)
```

### Generation
Package itself do not provide any command line application that would generate output out of given input.
Instead it encourage to write local generation application next to the proper package.
```go
package appg

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/piotrkowalczuk/pqt"
	"github.com/piotrkowalczuk/pqt/pqtgo"
	"github.com/piotrkowalczuk/pqt/pqtsql"
)

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
	}
)

func main() {
	file, err := os.Create("schema.pqt.go")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// For simplicity it is going to be empty.
	sch := pqt.NewSchema("app")
	fmt.Fprintf(file, `
		// Code generated by pqt.
		// source: appg/main.go
		// DO NOT EDIT!
	`)
	err := pqtgo.NewGenerator().
		AddImport("github.com/piotrkowalczuk/ntypes").
		AddImport("github.com/piotrkowalczuk/qtypes").
		SetAcronyms(acronyms).
		SetPackage("appd").
		GenerateTo(sch, file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(file, "const SQL = `\n")
	if err := pqtsql.NewGenerator().GenerateTo(sch, file); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(file, "`")
}

```

