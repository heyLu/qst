# qst - run things quickly (and easily)

intended to be run in unfamilar environments, you pass it a file or a
directory and it tries to detect what it is and how to run it.

- run `qst` or `qst -h` to see options and support project types
- `qst hello_world.go`: compiles and runs `hello_world.go`, rerunning
	after it exits or the file is saved

	quite fun for small things, just throw some code in a file, have `qst`
	watch and restart when appropriate.
- `qst -step=test ...` runs the tests for projects that support it
- `qst -type=make ...` to choose a specific type
- `qst -command="custom-build {file}" ...`, e.g. you can specify your own
	commands and only use the restart feature
- `qst -remote github.com/heyLu/qst/examples/hello_web.rb`, fetches a project
	from github and runs it
- `qst -detect ...` just displays the detected types (first would be chosen)

## Why?

- for simple things, for example "run go build whenever i change this"
- when you forgot how to start something
- to learn go

`qst` is a simple tool, and will stay simple. it is not intended to replace
anything, but to make your life a little bit simpler. there are more interesting
things to remember.

for more, see ["introducing quickstart"](https://github.com/heyLu/w/blob/master/introducing-quickstart.md)

## I want that

Great! You can either download precompile binaries ([bsd64][], [linux64][], [mac64][],
install it using `go get github.com/heyLu/qst` or clone the repo and build
it yourself to hack on it.

[bsd64]: https://github.com/heyLu/qst/releases/download/v0.1.0/qst-freebsd-amd64
[linux64]: https://github.com/heyLu/qst/releases/download/v0.1.0/qst-linux-amd64
[mac64]: https://github.com/heyLu/qst/releases/download/v0.1.0/qst-darwin-amd64

### Building it yourself

(Just a reminder: Only do this if you want to improve it, otherwise
`go get github.com/heyLu/qst` might be what you want instead.)

	# set up $GOPATH as desired
	$ export GOPATH=$PWD/.go         # choose whatever you want
	$ go build qst
	...
	$ ./qst -h
	Usage: qst <file>
	...
	$ ./qst examples/hello_web.rb
	...
	^C
	$ ./qst examples/hello_web.go
	...

Try changing something in the files, it's fun. :)

## Ideas/todo

- watch many files (select by globbing)
- sometimes restarts twice after one save?
- more project types
- tests (how? lots of shellscripts could do it, but would be very
	cumbersome. current "architecture" doesn't allow mocking.)
- more steps:
	* install (packages using the dependency manager of the type)
	* init (create a new blueprint for projects)

any further ideas? if so, open an issue or better yet, implement it yourself
and send me a pull request.