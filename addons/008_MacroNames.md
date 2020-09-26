# Add Macro as comments

Linenumber directives are great, but often when reading the intermittent code
we are more interrested in the name of the macro (codeblock) which provided the
code.

We start by extracting the different parts of Finalize; first we set up the
prev variable where we save the previous codeline, then we loop over each line,
taking care of lines where the metadata has changed and finally output the new metadata.

```go "Finalize Declaration"

// Finalize extract the textual lines from CodeBlocks and (if needed) prepend a
// notice about "unexpected" filename or line changes, which is extracted from
// the contained CodeLines. The result is a string with newlines ready to be
// pasted into a file.
func (block CodeBlock) Finalize() (ret string) {
	var prev CodeLine
	var lineformatstring string
	var macroformatstring string

	for _, current := range block {
		if !flags.publishable && (prev.number+1 != current.number || prev.file != current.file) {
			//<Finalize format>>>
		}
		ret += current.text
		prev = current
	}
	return
}
```

Add a flag and a variable for switching on macro names.

```go "flags for cli" +=
	macro bool
```
```go "Initialize" +=
flag.BoolVar(&flags.macro, "m", false, "macro names added in comments")
```

We based the default formatting on how the c family handles line directive and
comments. Of course it's a jungle out there, so we provide an easy way of
adding new formattings, by adding to "Finalize format languages". It is not
beautiful, but it gets the job done.

```go "Finalize format"
switch current.lang {
//<Finalize format languages>>>
}
if flags.macro && macroformatstring != "" && prev.macro != current.macro {
	ret += fmt.Sprintf(macroformatstring, current.macro)
}
if lineformatstring != "" {
	ret += fmt.Sprintf(lineformatstring, current.number, current.file)
}
```

Let's add a few languages the author cares about.

```go "Finalize format languages" +=
case "bash", "shell", "sh", "zsh", "python", "perl":
	macroformatstring = "# <<< %v >>>\n"
	lineformatstring = "\n#line %v \"%v\"\n"
case "go", "golang":
	macroformatstring = "//// <<< %v >>>\n"
	lineformatstring = "\n//line %[2]v:%[1]v\n"
case "CPP", "cpp", "Cpp":
	macroformatstring = "// <<< %v >>>\n"
	lineformatstring = "\n#line %v \"%v\"\n"
case "C", "c":
	// No surefire way to make line comments in c, we might be in a comment block already.
	lineformatstring = "\n#line %v \"%v\"\n"
```

The type definition of CodeLine needs to be extended with "macro" which is a
BlockName.

```go "Codeline type definition"
type CodeLine struct {
	text   string
	file   File
	lang   language
	number int
	macro  BlockName
}
```

The parseHeader function is starting to get out of hand, it might be time to
start looking at refactoring it as a method on CodeBlock. One of fname and
bname contains the name of the block, which we collects.

```go "Check block header"
fname, bname, appending, line.lang, fence = parseHeader(line.text)
if fname != "" {
	line.macro = BlockName(fname)
}
if bname != "" {
	line.macro = BlockName(fmt.Sprintf(`"%v"`, bname))
}
```

Lets make a new outputfile with macros printed out.

```go macro.go
// Code generated with lmt DO NOT EDIT.
//go:generate sh -c "go run main.go -m -o $GOFILE README.md addons/*.md && go fmt $GOFILE && echo please use main.go to produce a binary."
// This file is full of line directives, they are very useful when compiling and/or in user reports.
// If you are unconfortable with them, please look in lmt.go in the same directory.

//<main code>>>
```
