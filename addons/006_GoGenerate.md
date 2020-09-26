# Go generate

Go provides the ability to generate code with `go generate`. By providing
directives within a specifik go-file we are able to describe how it is built
from our code, a very useful feature, especially for litarate programming.

A go generate stanza is `^//go:generate COMMAND$` and every file which has been
generated should have a line matching the predifined regex below.

    `^// Code generated .* DO NOT EDIT.$`

With our new ability to extract one single file from `lmt`, and to choose
wether it is pretty-printed (without any line directives, ready to be read and
consumed by any golang user) or with line directives, we could produce both.
When we are releasing binary builds or showing of the general goodiness of
literate programming we continue to use main.go for meaningful bug reports,
crash reports or compile errors, but when proving we are producing compliant go
code or somebody just happens to read a through our go code it is simpler to
look at lmt.go.

Sadly go:generate doesn't seem to do "globbing" at all and lmt doesn't recurse
into directives, we have to call out to bourne shell once to do the globbing
for us. The user might not already have a binary of lmt, a problem we solve by
running main.go (the one providing meaningful errors) and let it produce the
generated files. If the build of main.go succeeds we abuse the system to
produce a "helpful" message to the user.

```go main.go
// Code generated with lmt DO NOT EDIT.
//go:generate sh -c "go run main.go -o $GOFILE README.md addons/*.md && echo run '`go build -o lmt main.go`' to produce a binary."
// This file is full of line directives, they are very useful when compiling and/or in user reports.
// If you are unconfortable with them, please look in lmt.go in the same directory.

//<main code>>>
```

Our implementation for lmt.go only differ in the addition of the `-p` flag.
Which prettyprints the result and without linestancas and other comments used
by Go in debug outputs.

```go lmt.go
// Code generated with lmt DO NOT EDIT.
//go:generate sh -c "go run main.go -p -o $GOFILE README.md addons/*.md && go fmt $GOFILE && echo please use main.go to produce a binary."
// This file is without line directives and is primarily for reading.
// When building and executable, please use main.go as it leaves information
// about the literate programming sources if you ever experience a crash,
// or having problem compiling.

//<main code>>>
```

Everything else has been split out to a new code block without any changes.

```go "main code"
package main

import (
	<<<main.go imports>>>
)

<<<global variables>>>

<<<other functions>>>

func main() {
	<<<main implementation>>>
}
```
