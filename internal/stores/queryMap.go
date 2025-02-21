package stores

import (
	"context"
	"time"
)

const (
	GetRedirect = iota
	InsertRedirects
)

type SQLQuery struct {
	SQLRequest string
	ctxTimeout time.Duration
}

var queryMap = make(map[int]SQLQuery)

func init() {
	queryMap[InsertRedirects] = SQLQuery{
		SQLRequest: `
			insert into redirects
			(is_active
			, url
			, refirect
			, date_create
			, date_update)
			values @Redirects
		`,
		ctxTimeout: 2 * time.Minute}
	queryMap[GetRedirect] = SQLQuery{
		SQLRequest: `
			select url, redirect
			from redirects
			where is_active = 1 and url = @URL limit 1
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
