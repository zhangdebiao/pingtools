module pingtools

go 1.17

replace (
	github.com/axgle/pflag => ./pflag
	github.com/axgle/util => ./util
)

require (
	github.com/axgle/mahonia v0.0.0-20180208002826-3358181d7394
	github.com/axgle/pflag v0.0.0-00010101000000-000000000000
	github.com/axgle/util v0.0.0-00010101000000-000000000000
	github.com/kardianos/service v1.2.1
)

require golang.org/x/sys v0.0.0-20201015000850-e3ed0017c211 // indirect
