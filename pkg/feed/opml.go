package feed

import (
	"context"
	"fmt"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/fs"
	"github.com/pkg/errors"
)

func BuildOPML(ctx context.Context, config *config.Config, db db.Storage, fs fs.Storage) (string, error) {

	xmlString := `<?xml version="1.0" encoding="utf-8" standalone="no"?>
<opml version="1.0">
<head>
<title>Podsync Feeds</title>
</head>
<body>
<outline text="feeds">
`
	for _, feed := range config.Feeds {

		f, err := db.GetFeed(ctx, feed.ID)
		if err != nil {
			return "", err
		}

		if feed.OPML {
			downloadURL, err := fs.URL(ctx, "", fmt.Sprintf("%s.xml", feed.ID))
			if err != nil {
				return "", errors.Wrapf(err, "failed to: obtain download URL for feed")
			}
			xmlString += fmt.Sprintf("<outline text=\"%s\" type=\"rss\" xmlUrl=\"%s\" />\n", f.Title, downloadURL)
		}
	}

	xmlString += `</outline>
</body>
</opml>
`
	return xmlString, nil
}
