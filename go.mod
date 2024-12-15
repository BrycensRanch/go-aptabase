module github.com/brycensranch/go-aptabase

go 1.22.0

require github.com/brycensranch/go-aptabase/pkg v0.0.0

replace github.com/brycensranch/go-aptabase/pkg => ./pkg // This tells Go to use the local directory. This is for internal testing.

require (
	golang.org/x/exp v0.0.0-20241210194714-1829a127f884 // indirect
	golang.org/x/sys v0.28.0 // indirect
)
