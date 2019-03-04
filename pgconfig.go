package pgconfig

import (
	"fmt"
	"log"

	"github.com/go-pg/pg"
	"github.com/sunfmin/pgconfig/envconfig"
)

var originalLookupEnv = envconfig.LookupEnv

type PgConfig interface {
	OnChange(v interface{}, onchange OnChangerFunc)
}

type AppConfig struct {
	LookupKey string
	Value     string
}

var createSQL = `
CREATE TABLE IF NOT EXISTS app_configs (lookup_key VARCHAR(100) PRIMARY KEY, value TEXT)
`

func (im *impl) lookupEnv(key string) (value string, found bool) {
	value, found = im.getFromDbCache(key)
	if !found {
		value, found = originalLookupEnv(key)
	}
	return
}

func (im *impl) getFromDbCache(key string) (value string, found bool) {
	for _, v := range im.dbvalues {
		if v.LookupKey == key {
			return v.Value, true
		}
	}

	return "", false
}

type OnChangerFunc func(vfilled interface{})

type impl struct {
	prefix    string
	pgOptions *pg.Options
	dbvalues  []*AppConfig
	db        *pg.DB
	chname    string
	ln        *pg.Listener
}

func New(prefix string, pgOptions *pg.Options) PgConfig {
	im := &impl{
		prefix:    prefix,
		pgOptions: pgOptions,
		db:        pg.Connect(pgOptions),
		chname:    fmt.Sprintf("app_configs_%s", prefix),
	}

	_, err := im.db.Exec(createSQL)
	if err != nil {
		panic(err)
	}

	log.Printf("Listen postgres channel at %s\n", im.chname)
	im.ln = im.db.Listen(im.chname)

	envconfig.LookupEnv = im.lookupEnv

	return im
}

func (im *impl) OnChange(v interface{}, onchange OnChangerFunc) {
	im.reload(&v)
	onchange(v)
	ch := im.ln.Channel()
	go func() {
		for {
			<-ch
			log.Printf("Got change notification from %s\n", im.chname)
			im.reload(v)
			onchange(v)
		}
	}()
	return
}

func (im *impl) reload(v interface{}) {
	err := im.db.Model(&im.dbvalues).Select()
	if err != nil {
		log.Println("Select()", err)
		return
	}

	err = envconfig.Process(im.prefix, v)
	if err != nil {
		log.Println("envconfig.Process", err)
		return
	}
	return
}
