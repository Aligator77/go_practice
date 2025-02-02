package stores

import "database/sql"

type UrlService struct {
	Db sql.DB
}

func CreateUrlService(db sql.DB) (us *UrlService) {
	us = &UrlService{
		Db: db,
	}

	return us
}

func (u *UrlService) Shutdown() error {
	err := u.Db.Close()

	return err
}
