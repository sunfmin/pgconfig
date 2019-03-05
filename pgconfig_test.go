package pgconfig_test

import (
	"log"
	"testing"
	"time"

	"github.com/go-pg/pg"
	"github.com/sunfmin/pgconfig"
	"github.com/sunfmin/pgconfig/envconfig"
)

type Specification struct {
	Debug bool
	Port  int
	User  string
}

type dbLogger struct {
	start time.Time
}

func (d dbLogger) BeforeQuery(q *pg.QueryEvent) {

}

func (d dbLogger) AfterQuery(q *pg.QueryEvent) {
	sql, _ := q.FormattedQuery()

	log.Printf("%s\n", sql)
}

func TestReload(t *testing.T) {
	var myval *Specification

	var opts = &pg.Options{}
	err := envconfig.Process("myapp_db", opts)
	if err != nil {
		panic(err)
	}

	var db = pg.Connect(&pg.Options{
		Addr:     opts.Addr,
		User:     opts.User,
		Password: opts.Password,
		Database: opts.Database,
	})

	db.AddQueryHook(dbLogger{})

	pgc := pgconfig.New("myapp", db)
	_, err = db.Exec("TRUNCATE TABLE app_configs")
	if err != nil {
		panic(err)
	}

	runm(
		db,
		`INSERT INTO app_configs (lookup_key, value) VALUES ('MYAPP_USER', 'Felix1')`,
	)

	pgc.OnChange(&Specification{}, func(v interface{}) {
		myval = v.(*Specification)
	})

	if myval.User != "Felix1" {
		t.Errorf("user should load initial value")
	}

	runm(
		db,
		`UPDATE app_configs SET value = 'Felix2' WHERE lookup_key = 'MYAPP_USER'`,
		`NOTIFY app_configs_myapp`,
	)

	time.Sleep(100 * time.Millisecond)

	if myval.User != "Felix2" {
		t.Errorf("wrong value loaded from db %+v", myval)
	}
}

func runm(db *pg.DB, qs ...string) {
	for _, q := range qs {
		_, err := db.Exec(q)
		if err != nil {
			panic(err)
		}
	}
}
