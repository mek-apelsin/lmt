# Calling conventions for the lmt executable

When applying the previous patches we get very (very) noisy output. This is not
in everyones interest, and we sometimes want to publish clean code without line
directives and the like. A few flags with options for the level of noise in the
output would probably be useful. Lets import "flag" from the standard library.
The pflag package from spf13 might be objectively better(?) but flag is
"standard" and not too shabby.

```go "main.go imports" +=
"flag"
```

We save all flag option variables in a single global struct. It might be
expanded later so lets make it easy, by adding it in the global block variables
block and putting the implementation in its own codeblock.

```go "global block variables" +=
var flags struct {
//<flags for cli>>>
}
```

We start of with one flag for chosing a single outfile, which is awesome
whenever we are using //go:generate, but also a flag for disabling all the
noise and another for extra noise (the name of the code blocks).

```go "flags for cli"
	outfile     string
	publishable bool
```

The flag pachage needs to connect our flags-variables to a flag letter, a
default value and a helpful comment. Lets put it last in the initalize section
in main.

```go "Initialize" +=
flag.StringVar(&flags.outfile, "o", "", "output a specific file instead of all files.")
flag.BoolVar(&flags.publishable, "p", false, "publishable output, without line directives.")
```

We need to override the main implementation since we don't want to use the
flags as inputfiles, luckily flag.Args is a list without the flags. But if we
ever want to add more flags, we need to handle them with flag.Parse(), lets
make sure we always run it after initalize by putting it after initalizem
instead adding flag as a codeblock in Initalize.  Lastly, we probably should
remove any and all files from the output when only requesting a single file.

```go "main implementation"

//<Initialize>>>
flag.Parse()

for _, file := range flag.Args() {
	//<Open and process file>>>
}
//<Override filelist>>>
//<Output files>>>
```

If we are requesting a single file, we could remove all the other mentions of
files, but it is way simpler to just create a new map for files to codeblocks,
and (if the requested file is in the original) copy the single file-codeblock
pair over and throwing the old "files" away. If the requested file is not
available in the code, the user should probably be warned. An implementation to
override the filelist might look like this.

```go "Override filelist"
if flags.outfile != "" {
	f := make(map[File]CodeBlock)
	if files[File(flags.outfile)] != nil {
		f[File(flags.outfile)] = files[File(flags.outfile)]
	} else {
		fmt.Fprintf(os.Stderr, "Warning: File named \"%s\" requested but not defined.\n", flags.outfile)
	}
	files = f
}
```

Lastly we don't write the line directives to file if the `publishable` flag is
set.

```go "Finalize Declaration"

// Finalize reads the textual lines from CodeBlocks and (if needed) prepend a
// notice about "unexpected" filename or line changes, which is extracted from
// the contained CodeLines. The result is a string with newlines ready to be
// pasted into a file.
func (block CodeBlock) Finalize() (ret string) {
	var prev CodeLine
	var formatstring string

	for _, current := range block {
		if !flags.publishable && (prev.number+1 != current.number || prev.file != current.file) {
			switch current.lang {
			//<Format strings for languages>>>
			}
			if formatstring != "" {
				ret += fmt.Sprintf(formatstring, current.number, current.file)
			}
		}
		ret += current.text
		prev = current
	}
	return
}
```
