package main

import (
	"flag"
	"github.com/go-ap/fedbox/app"
	"github.com/go-ap/fedbox/internal/config"
	"github.com/go-ap/fedbox/internal/log"
	"github.com/go-ap/fedbox/storage"
	"github.com/go-ap/storage/boltdb"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgx"
	"time"
)

var version = "HEAD"

const defaultTimeout = time.Second * 15

func main() {
	var wait time.Duration

	flag.DurationVar(&wait, "graceful-timeout", defaultTimeout, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	l := log.New()
	a := app.New(l, version)
	r := chi.NewRouter()

	var onClose func()

	if a.Config().Storage == config.BOLTDB {
		if b, err := boltdb.New(boltdb.Config{
			Path: a.Config().BoltDBPath,
			BucketName: "fedbox",
		}); err == nil {
			r.Use(app.Repo(b))
		}
	}
	if a.Config().Storage == config.POSTGRES {
		dbConf := a.Config().DB
		conn, err := pgx.NewConnPool(pgx.ConnPoolConfig{
			ConnConfig: pgx.ConnConfig{
				Host:     dbConf.Host,
				Port:     uint16(dbConf.Port),
				Database: dbConf.Name,
				User:     dbConf.User,
				Password: dbConf.Pw,
				Logger:   storage.DBLogger(l),
				//PreferSimpleProtocol: true,
			},
			MaxConnections: 3,
		})
		// TODO(marius) this defer Close doesn't seem right any more as we're not on top scope any more
		onClose = func () {
			l.Info("closing DB %v", conn.Close)
			conn.Close()
		}
		if err == nil {
			r.Use(app.Repo(storage.New(conn, a.Config().BaseURL, l)))
		} else {
			l.Errorf("invalid db connection")
		}
	}

	r.Use(middleware.RequestID)
	r.Use(log.NewStructuredLogger(l))

	r.Route("/", app.Routes())

	a.Run(r, wait, onClose)
}
