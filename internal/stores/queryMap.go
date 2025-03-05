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
	DisableRedirects
	GetRedirectsByUser
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
			, is_deleted
			, url
			, redirect
			, date_create
			, date_update
			, user_id)
			values ($1, $2, $3, $4, NOW(), NOW(), $5)
		`,
		ctxTimeout: 2 * time.Minute}
	queryMap[InsertBatchRedirects] = SQLQuery{
		SQLRequest: `
			insert into redirects
			(id
			, is_deleted
			, url
			, redirect
			, date_create
			, date_update
			, user_id
			)
			values
		`,
		ctxTimeout: 2 * time.Minute}
	queryMap[GetRedirect] = SQLQuery{
		SQLRequest: `
			select url
			     , redirect
			     , date_create
				 , date_update
				 , is_deleted
				 , user_id
			from redirects
			where redirect = $1 limit 1
		`,
		ctxTimeout: 2 * time.Minute,
	}
	queryMap[GetRedirectByURL] = SQLQuery{
		SQLRequest: `
			select id 
			     , url
			     , redirect
			     , date_create
				 , date_update
				 , is_deleted
				 , user_id
			from redirects
			where is_deleted = B'0' and url = $1 limit 1
		`,
		ctxTimeout: 2 * time.Minute,
	}
	queryMap[DisableRedirects] = SQLQuery{
		SQLRequest: `
			Update redirects
			set is_deleted = 1
			where url in $1
		`,
		ctxTimeout: 2 * time.Minute,
	}
	queryMap[GetRedirectsByUser] = SQLQuery{
		SQLRequest: `
			select id 
			     , url
			     , redirect
			     , date_create
				 , date_update
				 , is_deleted
				 , user_id
			from redirects
			where user_id = $1 
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
