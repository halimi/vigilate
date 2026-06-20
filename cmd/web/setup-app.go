package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/v2"
	"github.com/halimi/vigilate/internal/driver"
)

var session *scs.SessionManager

func setupApp() (*string, error) {
	insecurePort := flag.String("port", ":4000", "port to listen on")
	identifier := flag.String("identifier", "vigilate", "unique identifier")
	domain := flag.String("domain", "localhost", "domain name (e.g. example.com)")
	inProduction := flag.Bool("production", false, "application is in production")
	dbHost := flag.String("dbhost", "localhost", "database host")
	dbPort := flag.String("dbport", "5432", "database port")
	dbUser := flag.String("dbuser", "", "database user")
	dbPass := flag.String("dbpass", "", "database password")
	dbName := flag.String("dbname", "vigilate", "database name")
	dbSsl := flag.String("dbssl", "disable", "database ssl setting")

	flag.Parse()

	if *dbUser == "" || *dbPass == "" || *dbHost == "" || *dbPort == "" || *dbName == "" || *identifier == "" {
		log.Fatal("Missing required flags.")
	}

	log.Println("Connecting to database...")
	dsnString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5", *dbHost, *dbPort, *dbUser, *dbPass, *dbName, *dbSsl)
	db, err := driver.ConnectPostgres(dsnString)
	if err != nil {
		log.Fatal("Cannot connect to databse!", err)
	}

	log.Println("Initializing session manager...")
	session = scs.New()
	session.Store = postgresstore.New(db.SQL)
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.Name = fmt.Sprintf("gbsession_id_%s", *identifier)
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = *inProduction

	log.Println(domain)

	return insecurePort, nil
}
