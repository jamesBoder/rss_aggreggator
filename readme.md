# RSS Aggregator

A simple CLI tool to aggregate and read RSS feeds from the command line.

## Prerequisites

- **Go** (1.20 or higher)
- **PostgreSQL** database

## Installation

Install the CLI tool using Go:

```bash
go install github.com/jamesBoder/rss_aggreggator@latest
```

## Setup

### 1. Create the config file

Create a file named `.gatorconfig.json` in your home directory:

```json
{
  "db_url": "postgres://username:password@localhost:5432/rss_aggregator?sslmode=disable"
}
```

Replace `username` and `password` with your PostgreSQL credentials.

### 2. Create the database

```bash
createdb rss_aggregator
```

## Usage

### Register a new user
```bash
rss_aggreggator register <your_username>
```

### Add an RSS feed
```bash
rss_aggreggator addfeed "Blog Name" https://example.com/feed.xml
```

### Follow a feed
```bash
rss_aggreggator follow https://example.com/feed.xml
```

### Start aggregating feeds
```bash
rss_aggreggator agg 1m
```
This will fetch new posts every minute. Press `Ctrl+C` to stop.

### Browse your posts
```bash
rss_aggreggator browse 10
```
Shows the 10 most recent posts from feeds you follow.

### Other useful commands
- `rss_aggreggator users` - List all users
- `rss_aggreggator feeds` - List all feeds
- `rss_aggreggator following` - See feeds you're following
- `rss_aggreggator unfollow <feed_url>` - Unfollow a feed

## Quick Start Example

```bash
# Register yourself
rss_aggreggator register john

# Add and follow a feed
rss_aggreggator addfeed "TechCrunch" https://techcrunch.com/feed/

# Start aggregating (in a separate terminal)
rss_aggreggator agg 5m

# Browse posts
rss_aggreggator browse 5
```

Enjoy reading your RSS feeds!
