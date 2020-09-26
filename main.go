//line README.md:65
package main

import (
//line README.md:149
	"fmt"
	"os"
	"io"
//line README.md:212
	"bufio"
//line README.md:385
	"regexp"
//line README.md:510
	"strings"
//line addons/002_SubdirectoryFiles.md:35
	"path/filepath"
//line addons/005_Flags.md:11
	"flag"
//line README.md:69
)

//line addons/003_LineNumbers.md:25
type File string
type CodeBlock []CodeLine
type BlockName string
type language string
//line addons/003_LineNumbers.md:36
type CodeLine struct {
	text   string
	file   File
	lang   language
	number int
}
//line addons/003_LineNumbers.md:30

var blocks map[BlockName]CodeBlock
var files map[File]CodeBlock
//line addons/004_MarkupExpansion.md:91
type codefence struct {
	char  string // This should probably be a rune for purity
	count int
}
//line addons/005_Flags.md:19
var flags struct {
//line addons/005_Flags.md:29
	outfile     string
	publishable bool
//line addons/005_Flags.md:21
}
//line README.md:402
var namedBlockRe *regexp.Regexp
//line README.md:432
var fileBlockRe *regexp.Regexp
//line README.md:516
var replaceRe *regexp.Regexp
//line README.md:72

//line addons/003_LineNumbers.md:118
// Updates the blocks and files map for the markdown read from r.
func ProcessFile(r io.Reader, inputfilename string) error {
//line addons/003_LineNumbers.md:82
	scanner := bufio.NewReader(r)
	var err error

	var line CodeLine
	line.file = File(inputfilename)

	var inBlock, appending bool
	var bname BlockName
	var fname File
	var block CodeBlock
//line addons/004_MarkupExpansion.md:193
	var fence codefence
//line addons/003_LineNumbers.md:99
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
//line addons/004_MarkupExpansion.md:210
		if !inBlock {
//line addons/004_MarkupExpansion.md:225
			if len(line.text) >= 3 && (line.text[0:3] == "```" || line.text[0:3] == "~~~") {
				inBlock = true
				// We were outside of a block and now we are in one,
				// so just blindly reset the block variable.
				block = make(CodeBlock, 0)
//line addons/004_MarkupExpansion.md:187
				fname, bname, appending, line.lang, fence = parseHeader(line.text)
//line addons/004_MarkupExpansion.md:231
			}
//line addons/004_MarkupExpansion.md:212
			continue
		}
		if l := strings.TrimSpace(line.text); len(l) >= fence.count && strings.Replace(l, fence.char, "", -1) == "" {
//line addons/003_LineNumbers.md:56
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
//line addons/004_MarkupExpansion.md:216
			continue
		}
//line addons/003_LineNumbers.md:48
		block = append(block, line)
//line addons/003_LineNumbers.md:111
	}
//line addons/003_LineNumbers.md:121
}
//line addons/004_MarkupExpansion.md:129
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
//line addons/001_WhitespacePreservation.md:34
// Replace expands all macros in a CodeBlock and returns a CodeBlock with no
// references to macros.
func (c CodeBlock) Replace(prefix string) (ret CodeBlock) {
//line addons/003_LineNumbers.md:251
	var line string
	for _, v := range c {
		line = v.text
//line addons/003_LineNumbers.md:234
		matches := replaceRe.FindStringSubmatch(line)
		if matches == nil {
			if v.text != "\n" {
				v.text = prefix + v.text
			}
			ret = append(ret, v)
			continue
		}
//line addons/003_LineNumbers.md:220
		bname := BlockName(matches[2])
		if val, ok := blocks[bname]; ok {
			ret = append(ret, val.Replace(prefix+matches[1])...)
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Block named %s referenced but not defined.\n", bname)
			ret = append(ret, v)
		}
//line addons/003_LineNumbers.md:255
	}
	return
//line addons/001_WhitespacePreservation.md:38
}
//line addons/005_Flags.md:84

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
//line addons/003_LineNumbers.md:306
			case "bash", "shell", "sh", "perl":
				formatstring = "#line %v \"%v\"\n"
			case "go", "golang":
				formatstring = "//line %[2]v:%[1]v\n"
			case "C", "c":
				formatstring = "#line %v \"%v\"\n"
//line addons/005_Flags.md:97
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
//line addons/004_MarkupExpansion.md:155

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
//line README.md:74

func main() {
//line addons/005_Flags.md:50

//line README.md:157
	// Initialize the maps
	blocks = make(map[BlockName]CodeBlock)
	files = make(map[File]CodeBlock)
//line addons/004_MarkupExpansion.md:104
	namedBlockRe = regexp.MustCompile("^(?P<fence>`{3,}|~{3,})\\s?(?P<language>\\w*)\\s*\"(?P<name>.+)\"\\s*(?P<append>[+][=])?$")
//line addons/004_MarkupExpansion.md:113
	fileBlockRe = regexp.MustCompile("^(?P<fence>`{3,}|~{3,})\\s?(?P<language>\\w+)\\s+(?P<file>[\\w\\.\\-\\/]+)\\s*(?P<append>[+][=])?$")
//line addons/004_MarkupExpansion.md:83
	replaceRe = regexp.MustCompile(`^(?P<prefix>\s*)(?:<<|//)<(?P<name>.+)>>>\s*$`)
//line addons/005_Flags.md:38
	flag.StringVar(&flags.outfile, "o", "", "output a specific file instead of all files.")
	flag.BoolVar(&flags.publishable, "p", false, "publishable output, without line directives.")
//line addons/005_Flags.md:52
	flag.Parse()

	for _, file := range flag.Args() {
//line addons/003_LineNumbers.md:127
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
//line addons/005_Flags.md:56
	}
//line addons/005_Flags.md:69
	if flags.outfile != "" {
		f := make(map[File]CodeBlock)
		if files[File(flags.outfile)] != nil {
			f[File(flags.outfile)] = files[File(flags.outfile)]
		} else {
			fmt.Fprintf(os.Stderr, "Warning: File named \"%s\" requested but not defined.\n", flags.outfile)
		}
		files = f
	}
//line addons/003_LineNumbers.md:318
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
//line README.md:77
}
