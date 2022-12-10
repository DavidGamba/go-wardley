module github.com/DavidGamba/go-wardley

go 1.18

require (
	cuelang.org/go v0.4.3
	github.com/DavidGamba/go-getoptions v0.26.0
	github.com/ajstarks/svgo v0.0.0-20191124160048-bd5c74aaa11c
	github.com/davecgh/go-spew v1.1.1
	github.com/fsnotify/fsnotify v1.4.7
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/zclconf/go-cty v1.5.0
)

require (
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/apparentlymart/go-textseg/v12 v12.0.0 // indirect
	github.com/cockroachdb/apd/v2 v2.0.1 // indirect
	github.com/google/go-cmp v0.4.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7 // indirect
	github.com/mpvl/unique v0.0.0-20150818121801-cbe035fff7de // indirect
	github.com/pkg/errors v0.8.1 // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect

	// workaround for error: //go:linkname must refer to declared function or variable
	golang.org/x/sys v0.0.0-20220422013727-9388b58f7150 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)
