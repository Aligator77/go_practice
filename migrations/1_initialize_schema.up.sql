create table public.redirects
(
    id          text,
    is_active   bit,
    url         text,
    redirect    text,
    date_create timestamp,
    date_update timestamp
);

alter table public.redirects
    owner to postgres;

create index redirects_url_index
    on public.redirects (url);

create index redirects_redirect_index
    on public.redirects (redirect);

create index redirects_id_index
    on public.redirects (id);