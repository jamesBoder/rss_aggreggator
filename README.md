# Gator - RSS Aggregator

A simple CLI tool called **gator** to aggregate and read RSS feeds from the command line.

## Prerequisites

- **Go** (1.20 or higher)
- **PostgreSQL** database

## Installation

### Option 1: Install via go install
```bash
go install github.com/jamesBoder/rss_aggreggator/cmd/gator@latest
```

### Option 2: Build locally
```bash
git clone https://github.com/jamesBoder/rss_aggreggator.git
cd rss_aggreggator
go build -o gator ./cmd/gator
# Move to your PATH (optional)
sudo mv gator /usr/local/bin/
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
gator register <your_username>
```

### Add an RSS feed
```bash
gator addfeed "Blog Name" https://example.com/feed.xml
```

### Follow a feed
```bash
gator follow https://example.com/feed.xml
```

### Start aggregating feeds
```bash
gator agg 1m
```
This will fetch new posts every minute. Press `Ctrl+C` to stop.

### Browse your posts
```bash
gator browse 10
```
Shows the 10 most recent posts from feeds you follow.

### Other useful commands
- `gator users` - List all users
- `gator feeds` - List all feeds
- `gator following` - See feeds you're following
- `gator unfollow <feed_url>` - Unfollow a feed

## Quick Start Example

```bash
# Register yourself
gator register john

# Add and follow a feed
gator addfeed "TechCrunch" https://techcrunch.com/feed/

# Start aggregating (in a separate terminal)
gator agg 5m

# Browse posts
gator browse 5
```

Enjoy reading your RSS feeds!
