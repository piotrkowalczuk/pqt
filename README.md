# pqt [![GoDoc](https://godoc.org/github.com/piotrkowalczuk/pqt?status.svg)](http://godoc.org/github.com/piotrkowalczuk/pqt)&nbsp;[![Build Status](https://travis-ci.org/piotrkowalczuk/pqt.svg)](https://travis-ci.org/piotrkowalczuk/pqt)&nbsp;[![codecov.io](https://codecov.io/github/piotrkowalczuk/pqt/coverage.svg?branch=master)](https://codecov.io/github/piotrkowalczuk/pqt?branch=master)
This package is a toolbox for postgres driven applications.
It provides multiple tools to help to work with postgres efficiently.
In comparison to other currently available libraries instead of pushing struct tags into anti pattern or parsing SQL, it allows to define schema programmatically.

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
	- `iterator` - structure used by `FindIter` methods as a result, it wraps `sql.Rows`
	- `constants`:
		- `table names`
		- `column names`
		- `constraints` - library generates exact names of each constraint and corresponding constant that allow to easily handle query errors using [ErrorConstraint](https://godoc.org/github.com/piotrkowalczuk/pqt#ErrorConstraint) helper function
	- `repository` - data access layer that expose API to manipulate entities:
		- `Count` - returns number of entities for given criteria
		- `Find` - returns collection of entities that match given criteria
		- `FindIter` - works like `Find` but returns `iterator`
		- `Insert` - saves given entity into the database
		- `FindOneBy<primary-key>` - retrieves single entity
		- `UpdateBy<primary-key>` - modifies single entity
		- `DeleteBy<primary-key>` - modifies single entity
	- `func Scan<Entity>Rows(rows *sql.Rows) ([]*<entity>Entity, error) {` helper function
- __schema definition__ - allow to programmatically define database schema, that includes:
	- `schemas`
	- `tables`
	- `columns`
	- `constraints`
	- `relationships`

## Documentation

* [wiki](https://github.com/piotrkowalczuk/pqt/wiki)
* godoc [pqt](http://godoc.org/github.com/piotrkowalczuk/pqt)
* godoc [pqtgo](http://godoc.org/github.com/piotrkowalczuk/pqt/pqtgo)
* godoc [pqtsql](http://godoc.org/github.com/piotrkowalczuk/pqt/pqtsql)

## Example

Package itself do not provide any command line application that would generate output out of given input.
Instead it encourage to write local generation application next to the proper package.
Good example how such application could be structured can be found in [examples](https://github.com/piotrkowalczuk/pqt/tree/master/example).

By default example is trying to connect to local `test` database on default port.
To run it simply call:

```bash
$ make gen // not necessary, since generated code is already part of the repo
$ make run
```

## Contribution

Very welcome in general. Especially in fields like:

* postgres types better support
* support for functions
* better control over if generated code is private or public

