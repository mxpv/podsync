package feed

import (
	"context"
	"fmt"

	"encoding/xml"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/fs"
	"github.com/pkg/errors"
)

type opml struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    head
	Body    body
}

type head struct {
	XMLName xml.Name `xml:"head"`
	Title   string   `xml:"title"`
}

type body struct {
	XMLName  xml.Name  `xml:"body"`
	Outlines []outline `xml:"outline"`
}

type outline struct {
	Text   string `xml:"text,attr"`
	Title  string `xml:"title,attr"`
	Type   string `xml:"type,attr"`
	XMLURL string `xml:"xmlUrl,attr"`
}

func BuildOPML(ctx context.Context, config *config.Config, db db.Storage, fs fs.Storage) (string, error) {

	ou := make([]outline, 0)

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
			ou = append(ou, outline{Title: f.Title, Text: f.Title, Type: "rss", XMLURL: downloadURL})
		}
	}

	op := opml{Version: "1.0"}
	op.Head = head{Title: "PodSync feeds"}
	op.Body = body{Outlines: ou}

	out, _ := xml.MarshalIndent(op, " ", "  ")

	return xml.Header + string(out), nil

}
