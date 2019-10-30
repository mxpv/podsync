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
	log "github.com/sirupsen/logrus"

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

func (u *Updater) Update(ctx context.Context, cfg *config.Feed) error {
	log.WithFields(log.Fields{
		"id":      cfg.ID,
		"format":  cfg.Format,
		"quality": cfg.Quality,
	}).Infof("-> updating %s", cfg.URL)
	started := time.Now()

	// Create an updater for this feed type
	provider, err := u.makeBuilder(ctx, cfg)
	if err != nil {
		return err
	}

	// Query API to get episodes
	log.Debug("building feed")
	result, err := provider.Build(ctx, cfg)
	if err != nil {
		return err
	}

	log.Debugf("received %d episode(s) for %q", len(result.Episodes), result.Title)

	// Build iTunes XML feed with data received from builder
	log.Debug("building iTunes podcast feed")
	podcast, err := u.buildPodcast(result, cfg)
	if err != nil {
		return err
	}

	// Save XML to disk
	xmlName := fmt.Sprintf("%s.xml", cfg.ID)
	xmlPath := filepath.Join(u.config.Server.DataDir, xmlName)
	log.Debugf("saving feed XML file to %s", xmlPath)
	if err := ioutil.WriteFile(xmlPath, []byte(podcast.String()), 0600); err != nil {
		return errors.Wrapf(err, "failed to write XML feed to disk")
	}

	elapsed := time.Since(started)
	nextUpdate := time.Now().Add(cfg.UpdatePeriod.Duration)
	log.Infof("successfully updated feed in %s, next update at %s", elapsed, nextUpdate.Format(time.Kitchen))
	return nil
}

func (u *Updater) buildPodcast(feed *model.Feed, cfg *config.Feed) (*itunes.Podcast, error) {
	const (
		podsyncGenerator = "Podsync generator (support us at https://github.com/mxpv/podsync)"
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
		item.AddEnclosure(u.makeEnclosure(feed, episode, cfg))

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

func (u *Updater) makeEnclosure(feed *model.Feed, episode *model.Episode, cfg *config.Feed) (string, itunes.EnclosureType, int64) {
	ext := "mp4"
	contentType := itunes.MP4
	if feed.Format == model.FormatAudio {
		ext = "m4a"
		contentType = itunes.M4A
	}

	url := fmt.Sprintf("%s/%s/%s.%s", u.config.Server.Hostname, cfg.ID, episode.ID, ext)
	return url, contentType, episode.Size
}

func (u *Updater) makeBuilder(ctx context.Context, cfg *config.Feed) (builder.Builder, error) {
	var (
		provider builder.Builder
		err      error
	)

	info, err := link.Parse(cfg.URL)
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
