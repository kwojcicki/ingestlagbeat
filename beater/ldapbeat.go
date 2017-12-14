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

func (bt *Ldapbeat) query() {
	conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", bt.config.Server, bt.config.Port))
	if err != nil {
		logp.Warn("Couldn't connect to the ldap server: %s", err)
		return
	}
	logp.Info("Connected")

	defer conn.Close()
	err = conn.Bind(bt.config.Username, bt.config.Password)
	if err != nil {
		logp.Warn("Couldn't bind to the ldap server: %s", err)
		return
	}

	for _, query := range bt.config.Queries {
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
			return
		}

		logp.Info("%s", sr)
		for _, result := range sr.Entries {
			logp.Info("%s", result)
		}
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

		bt.query()
		// event := beat.Event{
		// 	Timestamp: time.Now(),
		// 	Fields: common.MapStr{
		// 		"type":    b.Info.Name,
		// 		"counter": counter,
		// 	},
		// }
		// bt.client.Publish(event)
		// logp.Info("Event sent")
		// counter++
	}
}

// Stop - stops beat
func (bt *Ldapbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
