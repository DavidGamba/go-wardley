module github.com/DavidGamba/go-wardley

go 1.13

require (
	github.com/DavidGamba/go-getoptions v0.27.0
	github.com/ajstarks/svgo v0.0.0-20211024235047-1546f124cd8b
	github.com/davecgh/go-spew v1.1.1
	github.com/fsnotify/fsnotify v1.6.0
	github.com/hashicorp/hcl/v2 v2.16.2
	github.com/zclconf/go-cty v1.13.1

	// workaround for error: //go:linkname must refer to declared function or variable
	golang.org/x/sys v0.6.0 // indirect
)
