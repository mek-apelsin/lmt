#!/bin/bash -e

errexit() { echo "$*" ; exit 1 ;}
has() {
	for c in "$@"; do
		command -v "$c" &>/dev/null || errexit "Missing command $c"
	done
}
test "$1" == "--help" || test "$1" == -h && { echo "subcommands: reseed and nofail. reseed takes parameter 'all' or number denoting source to reseed. Nofails continues without stopping at failing tests."; exit; }

has go lmt

mkdir -p ./BUILD
DIR=$(mktemp -d --tmpdir=./BUILD/ lmttest.XXXXXX)
# shellcheck disable=SC2064
trap "rm -rf \"$DIR\"" EXIT
mkdir -p "$DIR"
cd "$DIR"

fn=0
for f in ../../README.md ../../addons/*; do
	fa+=("$f")
	lmt -p -e main.go "${fa[@]}" > ./base.go
	unset "fi"
	in=0
	for i in ../../README.md ../../addons/* ;do
		fi+=("$i")
		test -f ../../tests/output/"$fn.$in" && go run ./base.go "${fi[@]}" || continue
		test "$1" == reseed && { test "$2" == "$in" || test $2 == "all" ;} && cp main.go ../../tests/output/"$fn.$in" ||
			diff --ignore-matching-lines='^//line' -u ../../tests/output/"$fn.$in" main.go || test "$1" == "nofail" ||
			{ cp main.go ../../err.out.go  ; errexit "Build failed with lmt from \"$f\" and input \"$i\". Output saved in err.out.go" ;}
		in=$((in+1))
	done
	fn=$((fn+1))
done

