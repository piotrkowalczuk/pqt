# pqt [![GoDoc](https://godoc.org/github.com/piotrkowalczuk/pqt?status.svg)](http://godoc.org/github.com/piotrkowalczuk/pqt)&nbsp;[![Build Status](https://travis-ci.org/piotrkowalczuk/pqt.svg)](https://travis-ci.org/piotrkowalczuk/pqt)&nbsp;[![codecov.io](https://codecov.io/github/piotrkowalczuk/pqt/coverage.svg?branch=master)](https://codecov.io/github/piotrkowalczuk/pqt?branch=master)

This package is a toolbox for postgres driven applications.
It provides multiple tools to help to work with postgres efficiently.
In comparison to other currently available libraries instead of pushing struct tags into anti pattern or parsing SQL, it allows to define schema programmatically.

## Features:

- __query builder__:
	- `Composer` - builder like object that keeps buffer and arguments but also tracks positional parameters.
- __array support__
	- `JSONArrayInt64` - wrapper for []int64, it generates JSONB compatible array `[]` instead of `{}`
	- `JSONArrayFloat64` - wrapper for []float64, it generates JSONB compatible array `[]` instead of `{}`
	- `JSONArrayString` - wrapper for []string, it generates JSONB compatible array `[]` instead of `{}`
- __sql generation__
- __go generation__ - it includes:
	- `<table-name>Entity` - struct that reflects single row within the database
	- `<table-name>Criteria` - object that can be passed to the `Find` method, it allows to create complex queries
	- `<table-name>Patch` - structure used by `UpdateOneBy<primary-key>` methods to modify existing cri
	- `<table-name>Iterator` - structure used by `FindIter` methods as a result, it wraps `sql.Rows`
	- `constants`:
		- `complete names`
		- `column names`
		- `constraints` - library generates exact names of each constraint and corresponding constant that allow to easily handle query errors using `ErrorConstraint` helper function
	- `<table-name>Repository` - data access layer that expose API to manipulate entities:
		- `Count` - returns number of entities for given cri
		- `Find` - returns collection of entities that match given cri
		- `FindIter` - works like `Find` but returns `iterator`
		- `Insert` - saves given cri into the database
		- `FindOneBy<primary-key>` - retrieves single cri, search by primary key
		- `FindOneBy<unique-key>` - retrieves single cri, search by unique key
		- `UpdateOneBy<primary-key>` - modifies single cri, search by primary key
		- `UpdateOneBy<unique-key>` - modifies single cri, search by unique key
		- `DeleteOneBy<primary-key>` - modifies single cri, search by primary key
	- `func Scan<Entity>Rows(rows *sql.Rows) ([]*<cri>Entity, error) {` helper function
- __schema definition__ - allow to programmatically define database schema, that includes:
	- `schemas`
	- `tables`
	- `columns`
	- `constraints`
	- `relationships`
- __helper functions__
    - `ErrorConstraint` - if possible extracts constraint from [pq.Error](https://godoc.org/github.com/lib/pq#Error) so it's easy to build switch statements using generated constraints.

## Documentation

* [wiki](https://github.com/piotrkowalczuk/pqt/wiki)
* godoc [pqt](http://godoc.org/github.com/piotrkowalczuk/pqt)
* godoc [pqtgo](http://godoc.org/github.com/piotrkowalczuk/pqt/pqtgo)
* godoc [pqtsql](http://godoc.org/github.com/piotrkowalczuk/pqt/pqtsql)

## Plugins 

[pqtgo](github.com/piotrkowalczuk/pqt/pqtgo) supports plugins over the [interface](https://godoc.org/github.com/piotrkowalczuk/pqt/pqtgo#Plugin).

* [ntypespqt](github.com/piotrkowalczuk/ntypes)
* [qtypespqt](github.com/piotrkowalczuk/qtypes)

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

## TODO

[ ] Postgres types better support.
[x] Support for functions.
[ ] Refactor `WithXXX` functions to be prefixed by the type they return for example `TableWithIfNotExists` or `ColumnWithNotNull`.
[ ] Constraint.
    [x] Index
    [x] Unique
    [x] Primary Key
    [x] Foreign Key
    [x] Check
    [ ] Exclusion
