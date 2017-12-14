// +build !integration

package config

import (
	"path/filepath"
	"testing"

	"github.com/elastic/beats/libbeat/cfgfile"
	"github.com/stretchr/testify/assert"
)

// Test_ReadConfig - testing config
func Test_ReadConfig(t *testing.T) {
	absPath, err := filepath.Abs("../")

	assert.NotNil(t, absPath)
	assert.Nil(t, err)

	config, err := cfgfile.Load(absPath + "/ldapbeat.yml")
	assert.Nil(t, err)

	queries := ConfigSettings{}
	err = config.Unpack(&queries)
	assert.Nil(t, err)

	assert.Equal(t, 2, len(queries.Ldapbeat.Queries))
	assert.Equal(t, "(&(objectClass=*)(objectClass=groupOfUniqueNames))", queries.Ldapbeat.Queries[0].Query)
	assert.Equal(t, "dc=example,dc=com", queries.Ldapbeat.Queries[0].BaseDN)
	assert.Equal(t, 2, queries.Ldapbeat.Queries[0].Scope)
	assert.Equal(t, 0, queries.Ldapbeat.Queries[0].DeRefAliases)
	assert.Equal(t, 0, queries.Ldapbeat.Queries[0].Sizelimit)
	assert.Equal(t, 0, queries.Ldapbeat.Queries[0].Timelimit)
	assert.Equal(t, false, queries.Ldapbeat.Queries[0].Typesonly)
	assert.Equal(t, []string{"dn", "cn", "objectClass"}, queries.Ldapbeat.Queries[0].Attributes)
}
