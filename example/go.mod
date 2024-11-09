module example

go 1.22.0

require github.com/brycensranch/go-aptabase/pkg v0.0.0
replace github.com/brycensranch/go-aptabase/pkg => ../pkg // This tells Go to use the local directory

require (
	golang.org/x/exp v0.0.0-20240909161429-701f63a606c0 // indirect
	golang.org/x/sys v0.26.0 // indirect
)
