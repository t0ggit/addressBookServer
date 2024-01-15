package psg

import (
	"addressBookServer/pkg"
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/url"
)

type Psg struct {
	conn *pgxpool.Pool
}

func NewPsg(dburl string, login, pass string) (psg *Psg, err error) {
	wErr := pkg.NewWrappedError("NewPsg()")

	psg = &Psg{}
	psg.conn, err = parseConnectionString(dburl, login, pass)
	if err != nil {
		wErr.Specify(err, "parseConnectionString(dburl, login, pass)").LogError()
		return nil, err
	}

	err = psg.conn.Ping(context.Background())
	if err != nil {
		wErr.Specify(err, "psg.conn.Ping(context.Background())").LogError()
		return nil, err
	}

	return psg, nil
}

func parseConnectionString(dburl, user, password string) (db *pgxpool.Pool, err error) {
	wErr := pkg.NewWrappedError("parseConnectionString()")

	var u *url.URL
	if u, err = url.Parse(dburl); err != nil {
		wErr.Specify(err, "url.Parse(dburl)").LogError()
		return nil, err
	}
	u.User = url.UserPassword(user, password)

	db, err = pgxpool.New(context.Background(), u.String())
	if err != nil {
		wErr.Specify(err, "pgxpool.New(context.Background(), u.String())").LogError()
		return nil, err
	}
	return db, nil
}
