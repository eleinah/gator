# gator - A Blog Aggregator
This is for the guided [blog aggregator project](https://www.boot.dev/courses/build-blog-aggregator-golang) on [Boot.dev](https://boot.dev)

## Prerequisites
- PostgreSQL (check your distribution's official repositories, i.e. `sudo apt search postgres`)
- Go ([install with Webi](https://webinstall.dev/golang/) or through the [official Go website](https://go.dev/doc/install))
- A `.gatorconfig.json` file in your home directory (`$HOME`) -- See below on how to set this up.

## Installation
You can install Gator CLI with the following:

```
go install github.com/eleinah/gator@latest
```

## Configuration File
Your configuration file should look something like this to start off with:

```
{"db_url":"postgress://<USERNAME>:<PASSWORD>@localhost:5432/gator?sslmode=disable","current_user_name":""}
```

Replace `<USERNAME>` and `<PASSWORD>` with the username and password of the system user running Postgres, i.e. `postgres:postgres`

<details>

<summary>Commands</summary>

The usage for gator is `gator <command> [args...]`

### login [user]
Logs into a user in the database

### register [user]
Register a user into the database

### reset
Resets the database

### users
Lists all users in the database, and which one is currently logged in

### agg [wait time between requests]
Start aggregating posts from feeds and populating the database, refreshing based on the given duration

### addfeed [URL]
Adds a feed by URL to the database

### feeds
Shows all feeds in the database

### follow [URL]
Follows a feed by URL for the logged in database user

### following
Displays the followed feeds for the logged in database user

### unfollow [URL]
Unfollows a feed by URL for the logged in database user

### browse [URL]
Browse all posts from followed feeds for the logged in database user

</details>
