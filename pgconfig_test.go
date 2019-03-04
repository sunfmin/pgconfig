package pgconfig_test

import (
	"testing"
	"time"

	"github.com/go-pg/pg"
	"github.com/sunfmin/pgconfig"
)

type Specification struct {
	Debug bool
	Port  int
	User  string
}

func TestReload(t *testing.T) {
	var myval *Specification

	var pgopts = &pg.Options{
		User:     "pgconfig",
		Password: "123",
		Database: "pgconfig_test",
		Addr:     "localhost:5001",
	}

	var db = pg.Connect(pgopts)

	pgc := pgconfig.New("myapp", pgopts)

	_, err := db.Exec("TRUNCATE TABLE app_configs")
	if err != nil {
		panic(err)
	}

	pgc.OnChange(&Specification{}, func(v interface{}) {
		myval = v.(*Specification)
	})

	if myval.User != "" {
		t.Errorf("user should have no value before update")
	}

	runm(
		db,
		`INSERT INTO app_configs (lookup_key, value) VALUES ('MYAPP_USER', 'Felix')`,
		`NOTIFY app_configs_myapp`,
	)

	time.Sleep(100 * time.Millisecond)

	if myval.User != "Felix" {
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
