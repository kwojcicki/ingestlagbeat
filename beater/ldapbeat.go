package beater

import (
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/go-ldap/ldap"
	"github.com/kwojcicki/ldapbeat/config"
)

type Ldapbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// Creates beater
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

		conn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", "ldap.forumsys.com", 389))
		if err != nil {
			continue
		}
		logp.Info("Connected")

		defer conn.Close()
		err = conn.Bind("cn=read-only-admin,dc=example,dc=com", "password")
		if err != nil {
			continue
		}
		searchRequest := ldap.NewSearchRequest(
			"dc=example,dc=com",
			ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
			"(&(%s)(|(objectClass=group)(objectClass=groupofnames)))",
			[]string{"dn", "cn"},
			nil,
		)
		sr, err := conn.Search(searchRequest)
		if err != nil {
			continue
		}

		sr.PrettyPrint(0)

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

func (bt *Ldapbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
