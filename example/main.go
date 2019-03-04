package main

import (
	"fmt"
	"time"

	"github.com/go-pg/pg"
	"github.com/sunfmin/pgconfig"
)

type Specification struct {
	Debug bool
	Port  int
	User  string
}

func main() {
	var myval *Specification

	var pgopts = &pg.Options{
		User:     "pgconfig",
		Password: "123",
		Database: "pgconfig_test",
		Addr:     "localhost:5001",
	}

	pgc := pgconfig.New("myapp", pgopts)

	pgc.OnChange(&Specification{}, func(v interface{}) {
		myval = v.(*Specification)
	})

	for {
		fmt.Println(myval.User)
		time.Sleep(1 * time.Second)
	}
}
