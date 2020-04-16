package feed

import (
	"context"
	"fmt"

	"github.com/gilliek/go-opml/opml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

func BuildOPML(ctx context.Context, config *config.Config, db feedProvider, provider urlProvider) (string, error) {
	doc := opml.OPML{Version: "1.0"}
	doc.Head = opml.Head{Title: "Podsync feeds"}
	doc.Body = opml.Body{}

	for _, feed := range config.Feeds {
		f, err := db.GetFeed(ctx, feed.ID)
		if err == model.ErrNotFound {
			// As we update OPML on per-feed basis, some feeds may not yet be populated in database.
			log.Debugf("can't find configuration for feed %q, ignoring opml", feed.ID)
			continue
		} else if err != nil {
			return "", errors.Wrapf(err, "failed to query feed %q", feed.ID)
		}

		if !feed.OPML {
			continue
		}

		downloadURL, err := provider.URL(ctx, "", fmt.Sprintf("%s.xml", feed.ID))
		if err != nil {
			return "", errors.Wrapf(err, "failed to get feed URL for %q", feed.ID)
		}

		outline := opml.Outline{
			Title:  f.Title,
			Text:   f.Description,
			Type:   "rss",
			XMLURL: downloadURL,
		}

		doc.Body.Outlines = append(doc.Body.Outlines, outline)
	}

	out, err := doc.XML()
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal OPML")
	}

	return out, nil
}
