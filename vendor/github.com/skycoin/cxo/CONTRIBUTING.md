CONTRIBUTING
============

- all `.go` souce files should be formated with `-s` flag
- length of a code line must not exceed 80 characters
- follow Golang naming convention: CamelCase, `Value()/SetValue()`,
  `IsExist()`, etc
- tests, examples and benchmarks should follow Golagn naming convention

More then that you should check you code with:

- `go vet` (`go vet ./...`)
- [`gocyclo`](https://github.com/fzipp/gocyclo) (`gocyclo -over 15 .`)
  It's not so strict
- [golint](https://github.com/golang/lint)
- [ineffassign](https://github.com/gordonklaus/ineffassign)
  (`ls ./**/*.go | xargs -L 1 ineffassign` or `ineffassign file.go`)
- [misspell](https://github.com/client9/misspell) (`misspell .`)

You features should have test case, that proof that it works.
