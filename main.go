package main

import (
	"os"

	"github.com/EdgarH78/jurassic-park/api"
	"github.com/EdgarH78/jurassic-park/data"
	"github.com/gin-gonic/gin"
)

func getEnvWithFallback(variableName, fallback string) string {
	value := os.Getenv(variableName)
	if value != "" {
		return value
	}
	return fallback
}

var (
	sqlHost         = getEnvWithFallback("SQL_HOST", "localhost")
	sqlUser         = getEnvWithFallback("SQL_USER", "admin")
	sqlPassword     = getEnvWithFallback("SQL_PASSWORD", "password")
	sqlDatabaseName = getEnvWithFallback("SQL_DATABASE_NAME", "jurassicpark")
)

func main() {
	sqlConfig := data.SQLConfig{
		User:         sqlUser,
		Password:     sqlPassword,
		Host:         sqlHost,
		DatabaseName: sqlDatabaseName,
	}
	parkSqlDao, err := data.NewParkSqlDao(sqlConfig)
	if err != nil {
		panic(err)
	}
	engine := gin.Default()
	api := api.NewAPI(parkSqlDao, engine)
	api.Run()
}
