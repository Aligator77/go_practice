-- +goose Up
-- +goose StatementBegin
create table public.redirects
(
    id          text,
    is_deleted   bit,
    url         text,
    redirect    text,
    date_create timestamp,
    date_update timestamp,
    user_id     text
);

alter table public.redirects
    owner to postgres;

create index redirects_url_index
    on public.redirects (url);

create index redirects_redirect_index
    on public.redirects (redirect);

create index redirects_id_index
    on public.redirects (id);

create index redirects_user_id_index
    on public.redirects (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.redirects;
-- +goose StatementEnd