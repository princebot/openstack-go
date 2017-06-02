config [![GoDoc](https://godoc.org/gitghub.com/princebot/openstack-go/config?status.svg)](https://godoc.org/github.com/princebot/openstack-go/config)
====
Config is a [Go](https://golang.org/doc) library
for loading OpenStack configuration from a
[clouds.yaml](http://docs.openstack.org/developer/python-openstackclient/configuration.html)
file.

This acts as an extension for [gophercloud](github.com/gophercloud/gophercloud.),
which does not currently include support for parsing <tt>clouds.yaml</tt>.

## Example
```go
package main

import (
	"log"

	"github.com/princebot/openstack-go/config"
	"github.com/gophercloud/gophercloud/openstack"
)

func main() {
	// Load clouds.yaml from default lookup paths.
	c, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// Retrieve configuration for a named cloud.
	options, err := c.Cloud("foo")
	if err != nil {
		log.Fatal(err)
	}

	// Now, we’re all set for gophercloud.
	provider, err := openstack.AuthenticatedClient(options)
	…
}
```

## Requirements
* [Go 1.7+](https://golang.org/doc/install)
* A valid [clouds.yaml](http://docs.openstack.org/developer/python-openstackclient/configuration.html) file

## Installation
```bash
go get -u github.com/princebot/openstack-go/config
```
