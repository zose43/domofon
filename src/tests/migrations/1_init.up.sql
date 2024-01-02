begin;

create table if not exists users
(
    id        int primary key generated always as identity,
    email     text  not null unique,
    pass_hash bytea not null
);

create index if not exists idx_email on users (email);

create table if not exists apps
(
    id     int primary key generated always as identity,
    name   text not null unique,
    secret text not null unique
);

commit