# gatorcli

**gatorcli** is a command-line RSS feed aggregator built in Go. It allows users to follow RSS feeds, fetch the latest items, and print them to the terminal. Data is persisted with PostgreSQL.

---

## Features

- Manage user accounts  
- Add, follow, unfollow RSS feeds  
- Aggregate and display newest feed items
- Persistent storage (PostgreSQL)  

---

## 📚 Prerequisites

Before you can run **gatorcli**, ensure you have:

- Go (version 1.22 or newer recommended)  
- PostgreSQL (running and accessible)  

---

## Installation

1. Clone the repo  
  ```bash
  git clone https://github.com/louiehdev/gatorcli.git
  cd gatorcli
  ```

2. Build or install with Go

  ```bash
  go install
  ```

---

## Configuration

gatorcli expects a config file to know how to connect to PostgreSQL.

Create a file named ~/.gatorconfig.json (or similar, based on what your code reads) with content:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gatordb?sslmode=disable"
}
```
Adjust:

- username
- password
- localhost:5432 (host and port)
- gatordb (database name)
- sslmode if needed

Be sure the database exists and your credentials are correct.

---

## Commands Overview

Here are a few key commands you can run:
|Command | Description|
|:---|:--------------------------------------:|
|register <user_name> | Create a new user account|
|login <user_name> | Log into your account|
|addfeed <rss_name> <rss_url> | Add a new RSS feed to follow|
|feeds | List all available feeds|
|follow <feed_url> | Follow a feed|
|unfollow <feed_url> | Unfollow a feed|
|agg | Fetch and display newest feed items|
