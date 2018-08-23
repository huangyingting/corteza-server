package sam

import (
	"github.com/namsral/flag"
	"github.com/pkg/errors"

	"github.com/crusttech/crust/sam/websocket"
)

type (
	configuration struct {
		http struct {
			addr    string
			logging bool
		}
		websocket websocket.Configuration
		db struct {
			dsn string
		}
	}
)

var config *configuration

func (c *configuration) Validate() error {
	if c == nil {
		return errors.New("SAM config is not initialized, need to call Flags()")
	}
	if c.http.addr == "" {
		return errors.New("No HTTP Addr is set, can't listen for HTTP")
	}
	if c.db.dsn == "" {
		return errors.New("No DB DSN is set, can't connect to database")
	}
	if err := c.websocket.Validate(); err != nil {
		return err
	}
	return nil
}

func Flags(prefix ...string) {
	if config != nil {
		return
	}
	if len(prefix) == 0 {
		panic("sam.Flags() needs prefix on first call")
	}
	config := new(configuration)
	(&config.websocket).Init()

	p := func(s string) string {
		return prefix[0] + "-" + s
	}

	flag.StringVar(&config.http.addr, p("http-addr"), ":3000", "Listen address for HTTP server")
	flag.BoolVar(&config.http.logging, p("http-log"), true, "Enable/disable HTTP request log")
	flag.StringVar(&config.db.dsn, p("db-dsn"), "crust:crust@tcp(db1:3306)/crust?collation=utf8mb4_general_ci", "DSN for database connection")
}