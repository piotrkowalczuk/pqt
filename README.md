# pqt [![GoDoc](https://godoc.org/github.com/piotrkowalczuk/pqt?status.svg)](http://godoc.org/github.com/piotrkowalczuk/pqt)&nbsp;[![Build Status](https://travis-ci.org/piotrkowalczuk/pqt.svg?branch=master)](https://travis-ci.org/piotrkowalczuk/pqt)&nbsp;[![codecov.io](https://codecov.io/github/piotrkowalczuk/pqt/coverage.svg?branch=master)](https://codecov.io/github/piotrkowalczuk/pqt?branch=master)

This package is a toolbox for Postgres driven applications.
It provides multiple tools to help to work with Postgres efficiently.
In comparison to other currently available libraries instead of pushing struct tags into anti-pattern or parsing SQL, it allows defining schema programmatically.

## Documentation

* wiki
    * [Features](https://github.com/piotrkowalczuk/pqt/wiki/Features)
    * [Entities](https://github.com/piotrkowalczuk/pqt/wiki/Entities)
    * [Types](https://github.com/piotrkowalczuk/pqt/wiki/Types)
    * [Repositories](https://github.com/piotrkowalczuk/pqt/wiki/Repositories)
    * [Error Handling](https://github.com/piotrkowalczuk/pqt/wiki/Error-Handling)
* godoc 
    * [pqt](http://godoc.org/github.com/piotrkowalczuk/pqt)
    * [pqtgo](http://godoc.org/github.com/piotrkowalczuk/pqt/pqtgo)
    * [pqtsql](http://godoc.org/github.com/piotrkowalczuk/pqt/pqtsql)

## Example

The package itself does not provide any command line application that would generate output out of given input. 
Instead, it encourages to write local generation application next to the proper package. 
A good example of how such an application could be structured can be found in [examples](https://github.com/piotrkowalczuk/pqt/tree/master/example).

By default, the example is trying to connect to local `test` database on the default port.
To run it simply call:

```bash
$ make gen // not necessary, since generated code is already part of the repo
$ make run
```

## Plugins 

[pqtgo](github.com/piotrkowalczuk/pqt/pqtgo) supports plugins over the [interface](https://godoc.org/github.com/piotrkowalczuk/pqt/pqtgo#Plugin).

* [ntypespqt](github.com/piotrkowalczuk/ntypes)
* [qtypespqt](github.com/piotrkowalczuk/qtypes)

## Contribution

Very welcome in general. Especially in fields like:

## TODO

* [x] Change `<entity-name>FindExpr.OrderBy` to slice.
* [ ] Postgres types better support.
* [x] Support for functions.
* [x] Selective go/sql generation.
* [x] Logical operations (`AND`, `OR`)
* [ ] Refactor `WithXXX` functions to be prefixed by the type they return for example `TableWithIfNotExists` or `ColumnWithNotNull`.
* [ ] Constraint.
    * [x] Index
    * [x] Unique
    * [x] Primary Key
    * [x] Foreign Key
    * [x] Check
    * [ ] Exclusion
