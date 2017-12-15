package beater

import (
	"fmt"
	"time"

	ldap "gopkg.in/ldap.v2"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/kwojcicki/ldapbeat/config"
)

// Ldapbeat - struct for beater
type Ldapbeat struct {
	done   chan struct{}
	config config.LdapBeatConfig
	client beat.Client
}

// New - Creates Beater
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	config := config.DefaultConfig
	if err := cfg.Unpack(&config); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Ldapbeat{
		done:   make(chan struct{}),
		config: config,
	}
	return bt, nil
}

func (bt *Ldapbeat) query(conn *ldap.Conn, query config.LDAPQuery) *ldap.SearchResult {
	searchRequest := ldap.NewSearchRequest(
		query.BaseDN,
		query.Scope, query.DeRefAliases, query.Sizelimit, query.Timelimit, query.Typesonly,
		query.Query,
		query.Attributes,
		nil,
	)
	sr, err := conn.Search(searchRequest)
	if err != nil {
		logp.Warn("Couldn't query ldap server: %s", err)
		return nil
	}
	return sr
}

func (bt *Ldapbeat) connectToLDAP() (*ldap.Conn, error) {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", bt.config.Server, bt.config.Port))
	if err != nil {
		logp.Warn("Couldn't connect to the ldap server: %s", err)
		return nil, err
	}
	logp.Info("Connected")

	err = conn.Bind(bt.config.Username, bt.config.Password)
	if err != nil {
		logp.Warn("Couldn't bind to the ldap server: %s", err)
		return nil, err
	}
	return conn, nil
}

func (bt *Ldapbeat) publishEvent(result *ldap.SearchResult, query config.LDAPQuery) {
	for _, entry := range result.Entries {
		fields := common.MapStr{
			"query": query.Query,
		}
		for _, attribute := range query.Attributes {
			if attribute == "dn" || attribute == "DN" {
				fields["dn"] = entry.DN
			} else {
				fields[attribute] = entry.GetAttributeValue(attribute)
			}
		}
		event := beat.Event{
			Timestamp: time.Now(),
			Fields:    fields,
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
	}
}

// Run - basic run loop for beat
func (bt *Ldapbeat) Run(b *beat.Beat) error {
	logp.Info("ldapbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}
		func() {
			for _, query := range bt.config.Queries {
				conn, err := bt.connectToLDAP()
				if err != nil {
					continue
				}
				defer conn.Close()
				result := bt.query(conn, query)
				bt.publishEvent(result, query)
			}
		}()
	}
}

// Stop - stops beat
func (bt *Ldapbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
