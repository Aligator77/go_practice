package stores

import (
	"context"
	"time"
)

const (
	GetRedirect = iota
	InsertRedirect
	InsertBatchRedirects
	GetRedirectByURL
)

type SQLQuery struct {
	SQLRequest string
	ctxTimeout time.Duration
}

var queryMap = make(map[int]SQLQuery)

func init() {
	queryMap[InsertRedirect] = SQLQuery{
		SQLRequest: `
			insert into redirects
			(id
			, is_active
			, url
			, redirect
			, date_create
			, date_update)
			values ($1, $2, $3, $4, NOW(), NOW())
		`,
		ctxTimeout: 2 * time.Minute}
	queryMap[InsertBatchRedirects] = SQLQuery{
		SQLRequest: `
			insert into redirects
			(id
			, is_active
			, url
			, redirect
			, date_create
			, date_update)
			values
		`,
		ctxTimeout: 2 * time.Minute}
	queryMap[GetRedirect] = SQLQuery{
		SQLRequest: `
			select url
			     , redirect
			     , date_create
				 , date_update
			from redirects
			where is_active = B'1' and redirect = $1 limit 1
		`,
		ctxTimeout: 2 * time.Minute,
	}
	queryMap[GetRedirectByURL] = SQLQuery{
		SQLRequest: `
			select id 
			     ,	url
			     , redirect
			     , date_create
				 , date_update
			from redirects
			where is_active = B'1' and url = $1 limit 1
		`,
		ctxTimeout: 2 * time.Minute,
	}
}

func Get(name int) (string, context.Context, context.CancelFunc) {
	sqlQuery := queryMap[name]
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, sqlQuery.ctxTimeout)

	return sqlQuery.SQLRequest, ctx, cancel
}
