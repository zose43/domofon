insert into apps (name, secret)
VALUES ('test', 'test-secret')
on conflict do nothing