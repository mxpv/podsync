package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"time"

	itunes "github.com/mxpv/podcast"
	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/builder"
	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/link"
	"github.com/mxpv/podsync/pkg/model"
)

type Updater struct {
	config *config.Config
}

func NewUpdater(config *config.Config) (*Updater, error) {
	return &Updater{config: config}, nil
}

func (u *Updater) Update(ctx context.Context, feed *config.Feed) error {
	// Create an updater for this feed type
	provider, err := u.makeBuilder(ctx, feed)
	if err != nil {
		return err
	}

	// Query API to get episodes
	result, err := provider.Build(feed)
	if err != nil {
		return err
	}

	// Build iTunes XML feed with data received from builder
	podcast, err := u.buildPodcast(result)
	if err != nil {
		return err
	}

	// Save XML to disk
	xmlName := fmt.Sprintf("%s.xml", result.ItemID)
	xmlPath := filepath.Join(u.config.Server.DataDir, xmlName)
	if err := ioutil.WriteFile(xmlPath, []byte(podcast.String()), 600); err != nil {
		return errors.Wrapf(err, "failed to write XML feed to disk")
	}

	return nil
}

func (u *Updater) buildPodcast(feed *model.Feed) (*itunes.Podcast, error) {
	const (
		podsyncGenerator = "Podsync generator"
		defaultCategory  = "TV & Film"
	)

	now := time.Now().UTC()

	p := itunes.New(feed.Title, feed.ItemURL, feed.Description, &feed.PubDate, &now)
	p.Generator = podsyncGenerator
	p.AddSubTitle(feed.Title)
	p.AddCategory(defaultCategory, nil)
	p.AddImage(feed.CoverArt)
	p.IAuthor = feed.Title
	p.AddSummary(feed.Description)

	if feed.Explicit {
		p.IExplicit = "yes"
	} else {
		p.IExplicit = "no"
	}

	if feed.Language != "" {
		p.Language = feed.Language
	}

	for i, episode := range feed.Episodes {
		item := itunes.Item{
			GUID:        episode.ID,
			Link:        episode.VideoURL,
			Title:       episode.Title,
			Description: episode.Description,
			ISubtitle:   episode.Title,
			IOrder:      strconv.Itoa(i),
		}

		pubDate := episode.PubDate
		if pubDate.IsZero() {
			pubDate = now
		}

		item.AddPubDate(&pubDate)

		item.AddSummary(episode.Description)
		item.AddImage(episode.Thumbnail)
		item.AddDuration(episode.Duration)
		item.AddEnclosure(u.makeEnclosure(feed, episode))

		// p.AddItem requires description to be not empty, use workaround
		if item.Description == "" {
			item.Description = " "
		}

		if feed.Explicit {
			item.IExplicit = "yes"
		} else {
			item.IExplicit = "no"
		}

		_, err := p.AddItem(item)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to add item to podcast (id %q)", episode.ID)
		}
	}

	return &p, nil
}

func (u *Updater) makeEnclosure(feed *model.Feed, episode *model.Episode) (string, itunes.EnclosureType, int64) {
	ext := "mp4"
	contentType := itunes.MP4
	if feed.Format == model.FormatAudio {
		ext = "m4a"
		contentType = itunes.M4A
	}

	url := fmt.Sprintf("%s/%s/%s.%s", u.config.Server.Hostname, feed.ItemID, episode.ID, ext)
	return url, contentType, episode.Size
}

func (u *Updater) makeBuilder(ctx context.Context, feed *config.Feed) (builder.Builder, error) {
	var (
		provider builder.Builder
		err      error
	)

	info, err := link.Parse(feed.URL)
	if err != nil {
		return nil, err
	}

	switch info.Provider {
	case link.ProviderYoutube:
		provider, err = builder.NewYouTubeBuilder(u.config.Tokens.YouTube)
	case link.ProviderVimeo:
		provider, err = builder.NewVimeoBuilder(ctx, u.config.Tokens.Vimeo)
	default:
		return nil, errors.Errorf("unsupported provider %q", info.Provider)
	}

	return provider, err
}
