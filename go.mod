module github.com/denysvitali/mpeg-dash-tools

go 1.18

require (
	github.com/alexflint/go-arg v1.4.2
	github.com/mc2soft/mpd v1.1.1
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.6.1
)

require (
	github.com/alexflint/go-scalar v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/mc2soft/mpd => ../mpd/
