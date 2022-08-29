package main

import (
	_ "embed"
	"weshare/cli"
	"weshare/sql"
)

//go:embed sql/sqlite.sql
var sqlliteDDL string

func init() {
	sql.InitDDL = sqlliteDDL
}

func main() {
	cli.ProcessArgs()
}
