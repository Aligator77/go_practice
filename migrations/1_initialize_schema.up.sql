-- +goose Up
-- +goose StatementBegin
create database yapr
    with owner postgres;

create table yapr.public.redirects
(
    id          text,
    is_active   bit,
    url         text,
    redirect    text,
    date_create timestamp,
    date_update timestamp
);

alter table yapr.public.redirects
    owner to postgres;

create index redirects_url_index
    on yapr.public.redirects (url);

create index redirects_redirect_index
    on yapr.public.redirects (redirect);

create index redirects_id_index
    on yapr.public.redirects (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS yapr.public.redirects;
DROP DATABASE IF EXISTS yapr;
-- +goose StatementEnd