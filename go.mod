module github.com/sporkops/cli

go 1.24.7

require (
	github.com/99designs/keyring v1.2.2
	github.com/spf13/cobra v1.10.2
	github.com/sporkops/spork-go v0.8.0
	golang.org/x/term v0.28.0
	gopkg.in/yaml.v3 v3.0.1
)

// Local development override — the SDK at v0.8.0 is the multi-org
// release; while it is being prepared for tagging, the CLI builds
// against the working tree alongside this repo. Remove the replace
// directive once spork-go v0.8.0 is published.
replace github.com/sporkops/spork-go => ../spork-go

require (
	github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4 // indirect
	github.com/danieljoos/wincred v1.1.2 // indirect
	github.com/dvsekhvalnov/jose2go v1.5.0 // indirect
	github.com/godbus/dbus v0.0.0-20190726142602-4481cbc300e2 // indirect
	github.com/gsterjov/go-libsecret v0.0.0-20161001094733-a6f4afe4910c // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mtibben/percent v0.2.1 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/sys v0.29.0 // indirect
)
