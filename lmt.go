// Code generated with lmt DO NOT EDIT.
//go:generate sh -c "go run main.go -p -o $GOFILE README.md addons/*.md && go fmt $GOFILE && echo please use main.go to produce a binary."
// This file is without line directives and is primarily for reading.
// When building and executable, please use main.go as it leaves information
// about the literate programming sources if you ever experience a crash,
// or having problem compiling.

package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type File string
type CodeBlock []CodeLine
type BlockName string
type language string
type CodeLine struct {
	text   string
	file   File
	lang   language
	number int
}

var blocks map[BlockName]CodeBlock
var files map[File]CodeBlock

type codefence struct {
	char  string // This should probably be a rune for purity
	count int
}

var flags struct {
	outfile     string
	publishable bool
	concatenate string
	extract     string
	listblocks  bool
	listfiles   bool
}
var namedBlockRe *regexp.Regexp
var fileBlockRe *regexp.Regexp
var replaceRe *regexp.Regexp

// Updates the blocks and files map for the markdown read from r.
func ProcessFile(r io.Reader, inputfilename string) error {
	scanner := bufio.NewReader(r)
	var err error

	var line CodeLine
	line.file = File(inputfilename)

	var inBlock, appending bool
	var bname BlockName
	var fname File
	var block CodeBlock
	var fence codefence
	for {
		line.number++
		line.text, err = scanner.ReadString('\n')
		switch err {
		case io.EOF:
			return nil
		case nil:
			// Nothing special
		default:
			return err
		}
		if !inBlock {
			if len(line.text) >= 3 && (line.text[0:3] == "```" || line.text[0:3] == "~~~") {
				inBlock = true
				// We were outside of a block and now we are in one,
				// so just blindly reset the block variable.
				block = make(CodeBlock, 0)
				fname, bname, appending, line.lang, fence = parseHeader(line.text)
			}
			continue
		}
		if l := strings.TrimSpace(line.text); len(l) >= fence.count && strings.Replace(l, fence.char, "", -1) == "" {
			inBlock = false
			// Update the files map if it's a file.
			if fname != "" {
				if appending {
					files[fname] = append(files[fname], block...)
				} else {
					files[fname] = block
				}
			}

			// Update the named block map if it's a named block.
			if bname != "" {
				if appending {
					blocks[bname] = append(blocks[bname], block...)
				} else {
					blocks[bname] = block
				}
			}
			continue
		}
		block = append(block, line)
	}
}
func parseHeader(line string) (File, BlockName, bool, language, codefence) {
	line = strings.TrimSpace(line) // remove indentation and trailing spaces

	// lets iterate over the regexps we have.
	for _, re := range []*regexp.Regexp{namedBlockRe, fileBlockRe} {
		if m := namedMatchesfromRe(re, line); m != nil {
			var fence codefence
			fence.char = m["fence"][0:1]
			fence.count = len(m["fence"])
			return File(m["file"]), BlockName(m["name"]), (m["append"] == "+="), language(m["language"]), fence
		}
	}

	// An empty return value for unnamed or broken fences to codeblocks.
	return "", "", false, "", codefence{}
}

// Replace expands all macros in a CodeBlock and returns a CodeBlock with no
// references to macros.
func (c CodeBlock) Replace(prefix string) (ret CodeBlock) {
	var line string
	for _, v := range c {
		line = v.text
		matches := replaceRe.FindStringSubmatch(line)
		if matches == nil {
			if v.text != "\n" {
				v.text = prefix + v.text
			}
			ret = append(ret, v)
			continue
		}
		bname := BlockName(matches[2])
		if val, ok := blocks[bname]; ok {
			ret = append(ret, val.Replace(prefix+matches[1])...)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Block named %s referenced but not defined.\n", bname)
			ret = append(ret, v)
		}
	}
	return
}

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
			case "bash", "shell", "sh", "perl":
				formatstring = "\n#line %v \"%v\"\n"
			case "go", "golang":
				formatstring = "\n//line %[2]v:%[1]v\n"
			case "C", "c":
				formatstring = "\n#line %v \"%v\"\n"
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

// namedMatchesfromRe takes an regexp and a string to match and returns a map
// of named groups to the matches. If not matches are found it returns nil.
func namedMatchesfromRe(re *regexp.Regexp, toMatch string) (ret map[string]string) {
	substrings := re.FindStringSubmatch(toMatch)
	if substrings == nil {
		return nil
	}

	ret = make(map[string]string)
	names := re.SubexpNames()

	for i, s := range substrings {
		ret[names[i]] = s
	}
	// The names[0] and names[x] from unnamed regex grous are an empty string.
	// Instead of checking every names[x] we simply overwrite the previous
	// ret[""] and discard it at the end.
	delete(ret, "")
	return
}

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

func main() {

	// Initialize the maps
	blocks = make(map[BlockName]CodeBlock)
	files = make(map[File]CodeBlock)
	namedBlockRe = regexp.MustCompile("^(?P<fence>`{3,}|~{3,})\\s?(?P<language>\\w*)\\s*\"(?P<name>.+)\"\\s*(?P<append>[+][=])?$")
	fileBlockRe = regexp.MustCompile("^(?P<fence>`{3,}|~{3,})\\s?(?P<language>\\w+)\\s+(?P<file>[\\w\\.\\-\\/]+)\\s*(?P<append>[+][=])?$")
	replaceRe = regexp.MustCompile(`^(?P<prefix>\s*)(?:<<|//)<(?P<name>.+)>>>\s*$`)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] files...\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&flags.outfile, "o", "", "output a specific file instead of all files.")
	flag.BoolVar(&flags.publishable, "p", false, "publishable output, without line directives.")
	flag.StringVar(&flags.concatenate, "c", "", "Concatenate a codeblock and print to standard out.")
	flag.StringVar(&flags.extract, "e", "", "Extract, expand a codeblock and print to standard out.")
	flag.BoolVar(&flags.listblocks, "l", false, "List all codeblocks.")
	flag.BoolVar(&flags.listfiles, "f", false, "List all output files.")
	flag.Parse()

	for _, file := range flag.Args() {
		f, err := os.Open(file)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error: ", err)
			continue
		}

		if err := ProcessFile(f, file); err != nil {
			fmt.Fprintln(os.Stderr, "error: ", err)
		}
		// Don't defer since we're in a loop, we don't want to wait until the function
		// exits.
		f.Close()
	}
	if flags.outfile != "" {
		f := make(map[File]CodeBlock)
		if files[File(flags.outfile)] != nil {
			f[File(flags.outfile)] = files[File(flags.outfile)]
		} else {
			fmt.Fprintf(os.Stderr, "Warning: File named \"%s\" requested but not defined.\n", flags.outfile)
		}
		files = f
	}
	switch {
	case flags.listfiles:
		fn := make([]string, 0, len(files))
		for n := range files {
			fn = append(fn, string(n))
		}
		sort.Strings(fn)
		fmt.Println(strings.Join(fn, "\n"))
	case flags.listblocks:
		bn := make([]string, 0, len(blocks))
		for n := range blocks {
			bn = append(bn, string(n))
		}
		sort.Strings(bn)
		fmt.Println(strings.Join(bn, "\n"))
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
	default:
		for filename, codeblock := range files {
			if dir := filepath.Dir(string(filename)); dir != "." {
				if err := os.MkdirAll(dir, 0775); err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
			}

			f, err := os.Create(string(filename))
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
			fmt.Fprintf(f, "%s", codeblock.Replace("").Finalize())
			// We don't defer this so that it'll get closed before the loop finishes.
			f.Close()
		}
	}
}
