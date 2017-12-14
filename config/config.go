// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import (
	"time"
)

type LdapBeatConfig struct {
	Period  time.Duration `config:"period"`
	Queries []LDAPQueries `config:"queries"`
}

type LDAPQueries struct {
	Query        string   `config:"filter"`
	BaseDN       string   `config:"basedn"`
	Scope        int      `config:"scope"`
	DeRefAliases int      `config:"derefaliases"`
	Sizelimit    int      `config:"sizelimit"`
	Timelimit    int      `config:"timelimit"`
	Typesonly    bool     `config:"typesonly"`
	Attributes   []string `config:"attributes"`
}

var DefaultConfig = LdapBeatConfig{
	Period: 1 * time.Second,
}

type ConfigSettings struct {
	Ldapbeat LdapBeatConfig
}
