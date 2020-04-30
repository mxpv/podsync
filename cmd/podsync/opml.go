package main

import (
	"encoding/xml"

	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/config"
)

func getOPML(feeds map[string]*config.Feed, server *config.Server) (doc *string, err error) {
	outlines := []Outline{}
	for name := range feeds {
		outlines = append(outlines, Outline{
			XMLURL: server.Hostname + "/" + name + ".xml",
			Text:   name,
			Type:   "rss",
		})
	}
	body := Body{
		Outlines: outlines,
	}
	root := OPML{
		Version: "1.0",
		Body:    body,
		Head: Head{
			Title: "Podsync Feeds",
		},
	}
	out, err := xml.MarshalIndent(root, "", "  ")
	if err != nil {
		log.Error("Error creating OPML file failed")
		return nil, err
	}
	xmlDoc := xml.Header + string(out) + "\n"
	return &xmlDoc, nil
}

type OPML struct {
	XMLName xml.Name `xml:"opml"`
	Version string   `xml:"version,attr"`
	Head    Head     `xml:"head"`
	Body    Body     `xml:"body"`
}

type Head struct {
	Title string `xml:"title"`
}

type Body struct {
	Outlines []Outline `xml:"outline"`
}

type Outline struct {
	XMLURL string `xml:"xmlUrl,attr"`
	Text   string `xml:"title,attr"`
	Type   string `xml:"type,attr"`
}
