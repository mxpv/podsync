# Episode Filters

Podsync supports filtering episodes by title, description, duration, and age. Filters are configured per feed under `[feeds.<id>.filters]`.

All filters use **AND logic** — an episode must satisfy every configured filter to be downloaded.

## Available Filters

| Field             | Type   | Description                                          |
| ----------------- | ------ | ---------------------------------------------------- |
| `title`           | string | Include only episodes whose title matches the regex  |
| `not_title`       | string | Exclude episodes whose title matches the regex       |
| `description`     | string | Include only episodes whose description matches      |
| `not_description` | string | Exclude episodes whose description matches           |
| `min_duration`    | int    | Exclude episodes shorter than N seconds              |
| `max_duration`    | int    | Exclude episodes longer than N seconds               |
| `min_age`         | int    | Skip episodes newer than N days                      |
| `max_age`         | int    | Skip episodes older than N days                      |

Regex patterns use [Go regular expression syntax](https://pkg.go.dev/regexp/syntax).

---

## Common Examples

### Exclude episodes by keyword

Use `not_title` with regex alternation (`|`) to skip episodes matching any of several keywords:

```toml
[feeds.my_feed.filters]
# Skip live streams, Q&A sessions, and shorts (case-insensitive)
not_title = "(?i)(live|q&a|#shorts)"
```

### Include only episodes matching a keyword

Use `title` to download only episodes that match a pattern:

```toml
[feeds.my_feed.filters]
# Download only tutorials and guides
title = "(?i)(tutorial|how.to|guide)"
```

### Filter by duration

Use `min_duration` and `max_duration` (in seconds) to skip episodes that are too short or too long:

```toml
[feeds.my_feed.filters]
# Only download full episodes (between 10 minutes and 3 hours)
min_duration = 600
max_duration = 10800
```

### Skip short clips and trailers

Combine a title filter with a minimum duration to exclude both by name and by length:

```toml
[feeds.my_feed.filters]
# Exclude anything labelled as a clip/preview AND skip anything under 5 minutes
not_title = "(?i)(clip|preview|trailer|teaser)"
min_duration = 300
```

### Only recent episodes

Use `max_age` to skip episodes older than a given number of days:

```toml
[feeds.my_feed.filters]
# Only keep episodes from the last 90 days
max_age = 90
```

### Skip very new episodes (wait for edits/corrections)

Use `min_age` to delay downloading until an episode is at least N days old:

```toml
[feeds.my_feed.filters]
# Wait at least 2 days before downloading (lets the creator fix mistakes)
min_age = 2
```

### Filter by description keyword

Use `description` to include only episodes whose description mentions a topic:

```toml
[feeds.my_feed.filters]
# Only download episodes that mention "interview" in the description
description = "(?i)interview"
```

### Combine title and description filters

All filters apply together — the episode must match every one:

```toml
[feeds.my_feed.filters]
# Include episodes about Python that are not beginner content
title       = "(?i)python"
not_title   = "(?i)(beginner|intro|101)"
min_duration = 600
```

### Match an exact phrase

Use `\b` word boundaries or anchors to be more precise:

```toml
[feeds.my_feed.filters]
# Only episodes containing the exact phrase "full episode"
title = "(?i)\\bfull episode\\b"
```

### Exclude several channels or series by title prefix

```toml
[feeds.my_feed.filters]
# Skip "Shorts:" and "Clip:" prefixed titles
not_title = "(?i)^(shorts?:|clip:)"
```

---

## Notes

- `title` and `description` are **include** filters: the episode is downloaded only if it matches.
- `not_title` and `not_description` are **exclude** filters: the episode is skipped if it matches.
- Duration and age filters always exclude episodes outside the specified range.
- When multiple filters are set, **all** must be satisfied (AND logic). Use regex alternation (`a|b`) for OR logic within a single filter field.
