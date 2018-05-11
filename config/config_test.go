// Copyright 2012, 2013, 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package config_test

import (
	"io/ioutil"
	"path"
	"testing"
	"time"

	jujutesting "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
	"gopkg.in/goose.v2/identity"
	"gopkg.in/macaroon-bakery.v2-unstable/bakery"

	"gopkg.in/juju/charmstore.v5/config"
)

func TestPackage(t *testing.T) {
	gc.TestingT(t)
}

type ConfigSuite struct {
	jujutesting.IsolationSuite
}

var _ = gc.Suite(&ConfigSuite{})

const testConfig = `
audit-log-file: /var/log/charmstore/audit.log
audit-log-max-size: 500
audit-log-max-age: 1
mongo-url: localhost:23456
api-addr: blah:2324
foo: 1
bar: false
auth-username: myuser
auth-password: mypasswd
identity-location: localhost:18082
identity-public-key: +qNbDWly3kRTDVv2UN03hrv/CBt4W6nxY5dHdw+KJFA=
identity-api-url: "http://example.com/identity"
terms-public-key: +qNbDWly3kRTDVv2UN03hrv/CBt4W6nxY5dHdw+KJFB=
terms-location: localhost:8092
agent-username: agentuser
agent-key:
  private: lsvcDkapKoFxIyjX9/eQgb3s41KVwPMISFwAJdVCZ70=
  public: +qNbDWly3kRTDVv2UN03hrv/CBt4W6nxY5dHdw+KJFA=
stats-cache-max-age: 1h
search-cache-max-age: 15m
request-timeout: 500ms
max-mgo-sessions: 10
blobstore: swift
swift-auth-url: 'https://foo.com'
swift-username: bob
swift-secret: secret
swift-bucket: bucket
swift-region: somewhere
swift-tenant: a-tenant
swift-authmode: userpass
logging-config: INFO
`

func (s *ConfigSuite) readConfig(c *gc.C, content string) (*config.Config, error) {
	// Write the configuration content to file.
	path := path.Join(c.MkDir(), "charmd.conf")
	err := ioutil.WriteFile(path, []byte(content), 0666)
	c.Assert(err, gc.Equals, nil)

	// Read the configuration.
	return config.Read(path)
}

func (s *ConfigSuite) TestRead(c *gc.C) {
	conf, err := s.readConfig(c, testConfig)
	c.Assert(err, gc.Equals, nil)
	c.Assert(conf, jc.DeepEquals, &config.Config{
		AuditLogFile:     "/var/log/charmstore/audit.log",
		AuditLogMaxAge:   1,
		AuditLogMaxSize:  500,
		MongoURL:         "localhost:23456",
		APIAddr:          "blah:2324",
		AuthUsername:     "myuser",
		AuthPassword:     "mypasswd",
		IdentityLocation: "localhost:18082",
		IdentityPublicKey: &bakery.PublicKey{
			Key: mustParseKey("+qNbDWly3kRTDVv2UN03hrv/CBt4W6nxY5dHdw+KJFA="),
		},
		TermsLocation: "localhost:8092",
		TermsPublicKey: &bakery.PublicKey{
			Key: mustParseKey("+qNbDWly3kRTDVv2UN03hrv/CBt4W6nxY5dHdw+KJFB="),
		},
		AgentUsername: "agentuser",
		AgentKey: &bakery.KeyPair{
			Public: bakery.PublicKey{
				Key: mustParseKey("+qNbDWly3kRTDVv2UN03hrv/CBt4W6nxY5dHdw+KJFA="),
			},
			Private: bakery.PrivateKey{
				mustParseKey("lsvcDkapKoFxIyjX9/eQgb3s41KVwPMISFwAJdVCZ70="),
			},
		},
		StatsCacheMaxAge:  config.DurationString{time.Hour},
		RequestTimeout:    config.DurationString{500 * time.Millisecond},
		MaxMgoSessions:    10,
		SearchCacheMaxAge: config.DurationString{15 * time.Minute},
		BlobStore:         config.SwiftBlobStore,
		SwiftAuthURL:      "https://foo.com",
		SwiftUsername:     "bob",
		SwiftSecret:       "secret",
		SwiftBucket:       "bucket",
		SwiftRegion:       "somewhere",
		SwiftTenant:       "a-tenant",
		SwiftAuthMode:     &config.SwiftAuthMode{identity.AuthUserPass},
		LoggingConfig:     "INFO",
	})
}

func (s *ConfigSuite) TestReadConfigError(c *gc.C) {
	cfg, err := config.Read(path.Join(c.MkDir(), "charmd.conf"))
	c.Assert(err, gc.ErrorMatches, ".* no such file or directory")
	c.Assert(cfg, gc.IsNil)
}

func (s *ConfigSuite) TestValidateConfigError(c *gc.C) {
	cfg, err := s.readConfig(c, "")
	c.Assert(err, gc.ErrorMatches, "missing fields mongo-url, api-addr, auth-username, auth-password in config file")
	c.Assert(cfg, gc.IsNil)

	cfg, err = s.readConfig(c, "blobstore: swift\n")
	c.Assert(err, gc.ErrorMatches, "missing fields mongo-url, api-addr, auth-username, auth-password, swift-auth-url, swift-username, swift-secret, swift-bucket, swift-region, swift-tenant, swift-auth-mode in config file")
	c.Assert(cfg, gc.IsNil)
}

func mustParseKey(s string) bakery.Key {
	var k bakery.Key
	err := k.UnmarshalText([]byte(s))
	if err != nil {
		panic(err)
	}
	return k
}
