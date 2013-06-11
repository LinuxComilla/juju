// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package environs_test

import (
	"io/ioutil"
	. "launchpad.net/gocheck"
	"launchpad.net/juju-core/environs"
	"launchpad.net/juju-core/testing"
)

type EmptyStorageSuite struct{}

var _ = Suite(&EmptyStorageSuite{})

func (s *EmptyStorageSuite) TestGet(c *C) {
	f, err := environs.EmptyStorage.Get("anything")
	c.Assert(f, IsNil)
	c.Assert(err, ErrorMatches, `file "anything" not found`)
}

func (s *EmptyStorageSuite) TestURL(c *C) {
	url, err := environs.EmptyStorage.URL("anything")
	c.Assert(url, Equals, "")
	c.Assert(err, ErrorMatches, `file "anything" not found`)
}

func (s *EmptyStorageSuite) TestList(c *C) {
	names, err := environs.EmptyStorage.List("anything")
	c.Assert(names, IsNil)
	c.Assert(err, IsNil)
}


type VerifyStorageSuite struct{}

var _ = Suite(&VerifyStorageSuite{})

const existingEnv = `
environments:
    test:
        type: dummy
        state-server: false
        authorized-keys: i-am-a-key
`

func (s *VerifyStorageSuite) TestVerifyStorage(c *C) {
	defer testing.MakeFakeHome(c, existingEnv, "existing").Restore()

	environ, err := environs.NewFromName("test")
	c.Assert(err, IsNil)
	storage := environ.Storage()
	err = environs.VerifyStorage(storage)
	c.Assert(err, IsNil)
	reader, err := storage.Get("bootstrap-verify")
	c.Assert(err, IsNil)
	defer reader.Close()
	contents, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	c.Check(string(contents), Equals,
		"juju-core storage writing verified: ok\n")
}
