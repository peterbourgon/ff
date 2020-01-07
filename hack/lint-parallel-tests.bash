#!/usr/bin/env bash

set -o pipefail

function for_all_test_files { find . -name '*_test.go' ; }
function first_line_of_test { xargs -n1 awk '/^func Test/{getline; print FILENAME ":" NR-1 " " $0}' ; }
function not_parallel       { grep -v 't.Parallel()' ; }

if for_all_test_files | first_line_of_test | not_parallel
then
	echo FAIL: not all tests call t.Parallel
	exit 1
fi

