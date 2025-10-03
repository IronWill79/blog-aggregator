# Gator - the blog aggregator

## Requirements

- Postgres
- Go v1.25.0+

## Installation

```bash
go install github.com/IronWill79/gator
```

## Configuration

- Create a `~/.gatorconfig.json` file with the following format

```json
{
  "db_url": "postgres://user:pass@domain/db"
}
```

## Usage

Commands are as follows :-

- `register <username>` - Registers a user in the DB and adds a config value in the `~/.gatorconfig.json` file
- `login <username>` - Changes the current user in the `~/.gatorconfig.json` file. Requires `register`ing the user.
- `addfeed <title> <url>` - Adds a feed and follows it by the current user.
- `agg <time_between_reqs>` - Aggregates feed posts every `time_between_reqs` period
- `browse [limit]` - Browse posts for current user. Default limit is 2.
- `feeds` - List all RSS feeds.
- `follow <url>` - Follows the `url` feed for the current user.
- `following` - Displays all feeds being followed by the current user.
- `reset` - Resets the user table, clearing all users.
- `unfollow <url>` - Unfollows the `url` feed for the current user.
- `users` - Lists all gator users.
