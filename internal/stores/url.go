package stores

import (
	"encoding/json"
	"github.com/rs/zerolog"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/Aligator77/go_practice/internal/config"
	"github.com/Aligator77/go_practice/internal/models"
)

type URLStore struct {
	DB         *config.ConnectionPool
	BaseURL    string
	Logger     zerolog.Logger
	LocalStore string
	DisableDB  string
	EmulateDB  map[string]models.Redirect
}

func NewURLService(db *config.ConnectionPool, Logger zerolog.Logger, BaseURL string, localStore string, DisableDBStore string) (us *URLStore) {
	us = &URLStore{
		DB:         db,
		BaseURL:    BaseURL,
		Logger:     Logger,
		LocalStore: localStore,
		DisableDB:  DisableDBStore,
		EmulateDB:  make(map[string]models.Redirect, 0),
	}

	return us
}

func (u *URLStore) Shutdown() error {
	err := u.DB.Close()

	return err
}

func (u *URLStore) MakeFullURL(link string) string {
	if !strings.Contains(link, "http") && len(u.BaseURL) > 0 {
		fullRedirect, _ := url.Parse(u.BaseURL)
		fullRedirect.Path = link
		return fullRedirect.String()
	} else {
		return link
	}
}

func (u *URLStore) StoreToFile(link string) error {
	if len(u.LocalStore) > 0 {
		f, err := os.OpenFile(u.LocalStore, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			u.Logger.Error().Err(err).Msg("Cannot open localStoreFile")
			return err
		}

		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				u.Logger.Error().Err(err).Msg("Cannot close localStoreFile")
			}
		}(f)

		if _, err = f.WriteString(link); err != nil {
			u.Logger.Error().Err(err).Msg("Cannot write to localStoreFile")
			return err
		}
	}
	return nil
}

func (u *URLStore) GetRedirect(id string) (redirect models.Redirect, err error) {
	if u.DisableDB == "0" {
		sqlRequest, ctx, cancel := Get(GetRedirect)
		defer cancel()

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.Logger.Error().Err(err).Msg("GetRedirect get connection failure")
			return redirect, err
		}
		defer conn.Close()

		row, err := conn.QueryContext(ctx, sqlRequest, id)
		if err != nil {
			u.Logger.Error().Err(err).Str("data", id).Msg("GetRedirect exec failure")
			return redirect, err
		}

		for row.Next() {

			if err := row.Scan(
				&redirect.URL,
				&redirect.Redirect,
				&redirect.DateCreate,
				&redirect.DateUpdate,
				&redirect.IsDelete,
				&redirect.User,
			); err != nil {
				u.Logger.Error().Err(err).Msg("scan failure")
				continue
			}
		}

	} else {
		redirect = u.EmulateDB[id]
	}

	return redirect, nil
}

func (u *URLStore) GetRedirectByURL(url string) (redirect models.Redirect, err error) {
	if u.DisableDB == "0" {
		sqlRequest, ctx, cancel := Get(GetRedirectByURL)
		defer cancel()

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.Logger.Error().Err(err).Msg("GetRedirect get connection failure")
			return redirect, err
		}
		defer conn.Close()

		row, err := conn.QueryContext(ctx, sqlRequest, url)
		if err != nil {
			u.Logger.Error().Err(err).Str("data", url).Msg("GetRedirect exec failure")
			return redirect, err
		}

		for row.Next() {

			if err := row.Scan(
				&redirect.ID,
				&redirect.URL,
				&redirect.Redirect,
				&redirect.DateCreate,
				&redirect.DateUpdate,
				&redirect.IsDelete,
				&redirect.User,
			); err != nil {
				u.Logger.Error().Err(err).Msg("scan failure")
				continue
			}
		}

	} else {
		redirect = u.EmulateDB[url]
	}

	return redirect, nil
}

func (u *URLStore) NewRedirect(redirect models.Redirect) (res models.Redirect, err error) {

	if u.DisableDB == "0" {
		sqlRequest, ctx, cancel := Get(InsertRedirect)
		defer cancel()

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.Logger.Error().Err(err).Msg("NewRedirect get connection failure")
			return redirect, err
		}
		defer conn.Close()

		res, err := conn.ExecContext(ctx, sqlRequest, redirect.ID, redirect.IsDelete, redirect.URL, redirect.Redirect, redirect.User)
		if err != nil {
			u.Logger.Error().Err(err).Str("data", redirect.String()).Msg("NewRedirect get connection failure")
			return redirect, err
		}

		if affected, err := res.RowsAffected(); affected > 0 {
			u.Logger.Warn().Str("affected", strconv.FormatInt(affected, 10)).Msg("NewRedirect exec has affected rows")
		} else if err != nil {
			u.Logger.Error().Err(err).Str("data", redirect.String()).Msg("NewRedirect get connection failure")
		}

	} else {
		u.EmulateDB[redirect.Redirect] = redirect
		u.EmulateDB[redirect.URL] = redirect
	}

	dataFile, _ := json.Marshal(redirect)
	_ = u.StoreToFile(string(dataFile))

	return redirect, nil
}

func (u *URLStore) NewRedirectsBatch(redirects []*models.Redirect) (id int64, err error) {
	if len(redirects) == 0 {
		return 0, nil
	}

	if u.DisableDB == "0" {
		sqlRequest, ctx, cancel := Get(InsertBatchRedirects)
		defer cancel()

		var queryStr strings.Builder
		queryStr.WriteString(sqlRequest)

		for i, r := range redirects {
			queryStr.WriteString(" (")
			queryStr.WriteString(`'` + r.ID + `', B'` + strconv.Itoa(r.IsDelete) + `', '` + r.URL + `', '` + r.Redirect + `', NOW(), NOW(), '` + r.User + `'`)
			queryStr.WriteString(")")
			if i != len(redirects)-1 {
				queryStr.WriteString(",")
			}
		}

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.Logger.Error().Err(err).Msg("NewRedirectsBatch get connection failure")
			return id, err
		}
		defer conn.Close()
		res, err := conn.ExecContext(ctx, queryStr.String())
		if err != nil {
			u.Logger.Error().Err(err).Str("data", strconv.FormatInt(id, 10)).Msg("NewRedirectsBatch ExecContext failure")
			return id, err
		}

		if affected, err := res.RowsAffected(); affected > 0 {
			u.Logger.Warn().Str("affected", strconv.FormatInt(affected, 10)).Msg("NewRedirectsBatch exec has affected rows")
		} else if err != nil {
			u.Logger.Error().Err(err).Str("data", strconv.FormatInt(id, 10)).Msg("NewRedirectsBatch RowsAffected = 0")
		}
	}
	for _, r := range redirects {
		u.EmulateDB[r.Redirect] = *r
		u.EmulateDB[r.URL] = *r
		dataFile, _ := json.Marshal(r)
		_ = u.StoreToFile(string(dataFile))

	}

	return id, nil
}

func (u *URLStore) DeleteRedirect(redirects []string) (affected bool, err error) {
	sqlRequest, ctx, cancel := Get(DisableRedirects)
	defer cancel()

	var queryStr strings.Builder
	queryStr.WriteString(sqlRequest)
	queryStr.WriteString("where url in (")

	for i, r := range redirects {
		queryStr.WriteString(`'` + r + `'`)
		if i != len(redirects)-1 {
			queryStr.WriteString(",")
		}
		if redirect, ok := u.EmulateDB[r]; ok {
			redirect.IsDelete = 1
			u.EmulateDB[r] = redirect
		}
	}
	queryStr.WriteString(")")
	u.Logger.Warn().Msg("DisableRedirects query " + queryStr.String())

	conn, err := u.DB.Conn(ctx)
	if err != nil {
		u.Logger.Error().Err(err).Msg("DisableRedirects get connection failure")
		return false, err
	}
	defer conn.Close()
	res, err := conn.ExecContext(ctx, queryStr.String())
	if err != nil {
		u.Logger.Error().Err(err).Str("data", queryStr.String()).Msg("DisableRedirects get connection failure")
		return false, err
	}
	a, err := res.RowsAffected()
	if a > 0 {
		u.Logger.Warn().Str("affected", strconv.FormatInt(a, 10)).Msg("DisableRedirects exec has affected rows")
	}

	return true, nil
}

func (u *URLStore) GetRedirectsByUser(userID string) (redirects []models.Redirect, err error) {
	if u.DisableDB == "0" {

		sqlRequest, ctx, cancel := Get(GetRedirectsByUser)
		defer cancel()

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.Logger.Error().Err(err).Msg("NewRedirect get connection failure")
			return redirects, err
		}
		defer conn.Close()
		row, err := conn.QueryContext(ctx, sqlRequest, userID)
		if err != nil {
			u.Logger.Error().Err(err).Str("userID", userID).Msg("GetRedirect exec failure")
			return redirects, err
		}

		for row.Next() {
			var redirect models.Redirect
			if err := row.Scan(
				&redirect.ID,
				&redirect.URL,
				&redirect.Redirect,
				&redirect.DateCreate,
				&redirect.DateUpdate,
				&redirect.IsDelete,
				&redirect.User,
			); err != nil {
				u.Logger.Error().Err(err).Msg("scan failure")
				continue
			}
			redirects = append(redirects, redirect)
		}
	} else {
		for _, r := range u.EmulateDB {
			if r.User == userID {
				redirects = append(redirects, r)
			}
		}
	}

	return redirects, nil
}
