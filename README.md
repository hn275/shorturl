# shorturl

[Live site](https://shorturl.haln.dev) (with TLS)

My friend made [HalFS](https://github.com/malcolmseyd/halfs), an API that uses this project as
a backend for object file storage!

This project is not a RestAPI project, though it does contain all the CRUD's! I wanted to know what
goes on when I run something like `fly launch` from [fly.io](https://fly.io), so it's deployed manually
with Nginx and auto-renewal TLS cert-bot. And the side quest is that I get to tinker around with
discrete math.

## URL Shortening Schema

Auto-incrementing key with SQLite database and a bijection function, this combination will map an
arbitrary length URL into a few characters.

The allowed characters are a-z, A-Z, and 0-9, which has 62 characters, are organized into a table.

```
abcdefghijklmnopqrstuvwxyz
ABCDEFGHIJKLMNOPQRSTUVWXYZ
0123456789
```

The auto-incrementing `id`'s from SQLite is then encoded into this base 62 version. The algorithm is
the same as [from decimal to binary](https://www.wikihow.com/Convert-from-Decimal-to-Binary), but
instead of base 2, we use base 62.

[implementation](./encode/encode.go).

## Deployment

```sh
docker compose up -d
```

TODO: write up this thing
