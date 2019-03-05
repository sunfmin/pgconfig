package main

import (
	"fmt"
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

type dbLogger struct{}

func (d dbLogger) BeforeQuery(q *pg.QueryEvent) {}

func (d dbLogger) AfterQuery(q *pg.QueryEvent) {
	fmt.Println(q.FormattedQuery())
}

func main() {
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

	pgc.OnChange(&Specification{}, func(v interface{}) {
		myval = v.(*Specification)
	})

	for {
		fmt.Println(myval.User)
		time.Sleep(1 * time.Second)
	}
}
