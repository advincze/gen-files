gen-files
=========

code generator that uses template files to create a code skeleton 

### install 
if you have [go](golang.org/dl) installed:
```
$ go get -u -v github.com/advincze/gen-files
```
you'll find `gen-files` in `$GOPATH/bin/gen-files`

### use

```
$ gen-files

Usage of gen-files:
  -config string
    	config file (required)
  -index string
    	index file (required)
  -to string
    	target dir (required)
```

the config file is a generic JSON file where you can store arbitratry configuration for 

```
{
	"foo":"bar"
}
```

the index file contains the files to create:

```
{
  "files": [
    {
      "from": "foo/Hello.java",
      "to": "{{ .foo }}/World.java"
    },
    {
      "from": "foo/Hello.java",
      "to": "{{ .foo }}2/World.java"
    },
    { ...
```

the template files and the names of the "to" paths can use go template syntax:

https://golang.org/pkg/text/template/
