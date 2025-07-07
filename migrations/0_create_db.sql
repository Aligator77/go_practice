-- +goose Up
create database yapr
    with owner postgres;

-- +goose Down
DROP DATABASE IF EXISTS yapr;
