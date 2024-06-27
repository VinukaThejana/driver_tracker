package connections

import (
	"database/sql"
	"fmt"
	"net/url"

	// Driver for the mssql server
	_ "github.com/denisenkom/go-mssqldb/azuread"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/enums"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
)

// InitDB is a function that is used to initialize databases
func (c *C) InitDB(e *env.Env) {
	username := e.DBUser
	password := ""
	if e.Env == string(enums.Prd) {
		password = e.DBPassword1
	} else {
		password = url.QueryEscape(fmt.Sprintf("%s#$%s@#$%d", e.DBPassword1, e.DBPassword2, e.DBPassword3))
	}
	host := e.DBHost
	port := e.DBPort
	database := e.DBDatabase
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&encrypt=true&TrustServerCertificate=true", username, password, host, port, database)

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		lib.LogFatal(err)
	}
	if err := db.Ping(); err != nil {
		lib.LogFatal(err)
	}

	c.DB = db
}
