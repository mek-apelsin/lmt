# Extract snippets from literate code

When doing refactoring of code it is often very useful to see the full content
of a concatenate codeblock without expansions straight in the terminal. If we
are to implement this we could probably include one flag for extraction with
expansions and a listing of all the codeblock `lmt` knows of. The addition of a
flag for listing all the files is probably good to add for completeness.

```go "Initialize" +=
flag.StringVar(&flags.concatenate, "c", "", "Concatenate a codeblock and print to standard out.")
flag.StringVar(&flags.extract, "e", "", "Extract, expand a codeblock and print to standard out.")
flag.BoolVar(&flags.listblocks, "l", false, "List all codeblocks.")
flag.BoolVar(&flags.listfiles, "f", false, "List all output files.")
```

These flags are added to the global variable `flags`.

```go "flags for cli" +=
	concatenate string
	extract     string
	listblocks  bool
	listfiles   bool
```

Since we introduce other mode of functioning than writing files to disk, our
main implementation needs to change. The simplest way is probably by adding a
switch, and setting the default to output files, which have been the only mode
of operation before.

```go "main implementation"

//<Initialize>>>
flag.Parse()

for _, file := range flag.Args() {
	//<Open and process file>>>
}
//<Override filelist>>>
switch {
//<Output files override>>>
default:
	//<Output files>>>
}
```

We described three different new modes of output which overrides the outputting
of files, lets list them.

```go "Output files override" +=
//<Implement flags to list files>>>
//<Implement flags to list codeblocks>>>
//<Check flags to print content to standard out>>>
```


We do want the `-c` and `-e` flags to work kind of the same, so we try to
handle both cases at the same time, in an attempt to make the differences
between them clearer.

```go "Check flags to print content to standard out"
case flags.concatenate != "", flags.extract != "":
	for i, v := range map[rune]string{'c': flags.concatenate, 'e': flags.extract} {
		if v != "" {
			cb, err := getBlockByName(v)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Block named \"%s\" requested but not defined.\n", v)
				return
			}
			switch i {
			case 'c':
				fmt.Fprintf(os.Stdout, "%s", cb.Finalize())
			case 'e':
				fmt.Fprintf(os.Stdout, "%s", cb.Replace("").Finalize())
			}
		}
	}
```

If all we have is a name of a block, we can't really be sure if it is placed in
files map or the blocks map. Let's make the files take precedent over blocks,
and return the first one it finds.

```go "Extract a codeblock by a name"

// getBlockByName takes a string as a name and use it as a key in files and
// blocks and return the first codeblock it could find. If no codeblocks are
// found by that name getBlockByName returns an error.
func getBlockByName(bn string) (CodeBlock, error) {
	// TODO: Why not make files a simple list and store all codeblocks in blocks?
	if _, filesiscb := files[File(bn)]; filesiscb {
		return files[File(bn)], nil
	}
	if _, blockiscb := blocks[BlockName(bn)]; blockiscb {
		return blocks[BlockName(bn)], nil
	}
	return nil, errors.New("No CodeBlock by that name")
}
```

```go "other functions" +=
//<Extract a codeblock by a name>>>
```

We COULD make two functions which takes map[X]Codeblock (where X is either a
File or a BlockName), converts all of it to strings, and runs the two snippets
of codes below as one function. We might even make them methods on two new
types and make blocks and files implementations of these. It would add up to a
lot more code though. It would probably be a idea to look inte making files a
simple list of the files `lmt` knows about and save it all in blocks and leave
the difference between them in UI alone. For now we make a simple slice of all
the names, sorts them and print it out on standard out.

```go "Implement flags to list codeblocks"
case flags.listblocks:
	bn := make([]string, 0, len(blocks))
	for n := range blocks {
		bn = append(bn, string(n))
	}
	sort.Strings(bn)
	fmt.Println(strings.Join(bn, "\n"))
```

```go "Implement flags to list files"
case flags.listfiles:
	fn := make([]string, 0, len(files))
	for n := range files {
		fn = append(fn, string(n))
	}
	sort.Strings(fn)
	fmt.Println(strings.Join(fn, "\n"))
```

We've added sort as a new dependency, for listing code and files and errors
when writing getBlockByName.

```go "main.go imports" +=
"sort"
"errors"
```
