package main

import (
	_ "embed"

	"github.com/code-to-go/safepool/sql"
)

//go:embed sql/sqlite.sql
var sqlliteDDL string

func init() {
	sql.InitDDL = sqlliteDDL
}

// func main() {
// 	fmt.Print("This is just a library! ")
// }
