package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	itunes "github.com/mxpv/podcast"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/link"
	"github.com/mxpv/podsync/pkg/model"
)

type Downloader interface {
	Download(ctx context.Context, feedConfig *config.Feed, episode *model.Episode, feedPath string) (string, error)
}

type Updater struct {
	config     *config.Config
	downloader Downloader
}

func NewUpdater(config *config.Config, downloader Downloader) (*Updater, error) {
	return &Updater{config: config, downloader: downloader}, nil
}

func (u *Updater) Update(ctx context.Context, feedConfig *config.Feed) error {
	log.WithFields(log.Fields{
		"feed_id": feedConfig.ID,
		"format":  feedConfig.Format,
		"quality": feedConfig.Quality,
	}).Infof("-> updating %s", feedConfig.URL)
	started := time.Now()

	// Make sure feed directory exists
	feedPath := filepath.Join(u.config.Server.DataDir, feedConfig.ID)
	log.Debugf("creating directory for feed %q", feedPath)
	if err := os.MkdirAll(feedPath, 0755); err != nil {
		return errors.Wrapf(err, "failed to create directory for feed %q", feedConfig.ID)
	}

	// Create an updater for this feed type
	provider, err := u.makeBuilder(ctx, feedConfig)
	if err != nil {
		return err
	}

	// Query API to get episodes
	log.Debug("building feed")
	result, err := provider.Build(ctx, feedConfig)
	if err != nil {
		return err
	}

	log.Debugf("received %d episode(s) for %q", len(result.Episodes), result.Title)

	// Since there is no way to detect the size of an episode after download and encoding via API,
	// we'll patch XML feed with values from this map
	sizes := map[string]int64{}

	// The number of episodes downloaded during this update
	downloaded := 0

	// Download and encode episodes
	for idx, episode := range result.Episodes {
		logger := log.WithFields(log.Fields{
			"index":      idx,
			"episode_id": episode.ID,
		})

		episodePath := filepath.Join(feedPath, u.episodeName(feedConfig, episode))
		_, err := os.Stat(episodePath)
		if err != nil && !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to check whether episode exists")
		}

		if os.IsNotExist(err) {
			// There is no file on disk, download episode
			logger.Infof("! downloading episode %s", episode.VideoURL)
			if output, err := u.downloader.Download(ctx, feedConfig, episode, feedPath); err == nil {
				downloaded++
			} else {
				// YouTube might block host with HTTP Error 429: Too Many Requests
				// We still need to generate XML, so just stop sending download requests and
				// retry next time
				if strings.Contains(output, "HTTP Error 429") {
					logger.WithError(err).Warnf("got too many requests error, will retry download next time")
					break
				}

				logger.WithError(err).Errorf("youtube-dl error: %s", output)
			}
		} else {
			// Episode already downloaded
			logger.Debug("skipping download of episode")
		}

		// Record file size
		if size, err := u.fileSize(episodePath); err != nil {
			// Don't return on error, use estimated file size provided by builders
			logger.WithError(err).Error("failed to get episode file size")
		} else { //nolint
			logger.Debugf("file size %d", size)
			sizes[episode.ID] = size
		}
	}

	// Build iTunes XML feed with data received from builder
	log.Debug("building iTunes podcast feed")
	podcast, err := u.buildPodcast(result, feedConfig, sizes)
	if err != nil {
		return err
	}

	// Save XML to disk
	xmlName := fmt.Sprintf("%s.xml", feedConfig.ID)
	xmlPath := filepath.Join(u.config.Server.DataDir, xmlName)
	log.Debugf("saving feed XML file to %s", xmlPath)
	if err := ioutil.WriteFile(xmlPath, []byte(podcast.String()), 0600); err != nil {
		return errors.Wrapf(err, "failed to write XML feed to disk")
	}

	elapsed := time.Since(started)
	nextUpdate := time.Now().Add(feedConfig.UpdatePeriod.Duration)
	log.Infof(
		"successfully updated feed in %s, downloaded: %d episode(s), next update at %s",
		elapsed,
		downloaded,
		nextUpdate.Format(time.Kitchen),
	)
	return nil
}

func (u *Updater) buildPodcast(feed *model.Feed, cfg *config.Feed, sizes map[string]int64) (*itunes.Podcast, error) {
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
		// Fixup episode size after downloading and encoding
		if size, ok := sizes[episode.ID]; ok {
			episode.Size = size
		}

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

func (u *Updater) makeEnclosure(
	feed *model.Feed,
	episode *model.Episode,
	cfg *config.Feed,
) (string, itunes.EnclosureType, int64) {
	ext := "mp4"
	contentType := itunes.MP4
	if feed.Format == model.FormatAudio {
		ext = "mp3"
		contentType = itunes.MP3
	}

	url := fmt.Sprintf(
		"%s/%s/%s.%s",
		u.hostname(),
		cfg.ID,
		episode.ID,
		ext,
	)

	return url, contentType, episode.Size
}

func (u *Updater) hostname() string {
	hostname := strings.TrimSuffix(u.config.Server.Hostname, "/")
	if !strings.HasPrefix(hostname, "http") {
		hostname = fmt.Sprintf("http://%s", hostname)
	}

	return hostname
}

func (u *Updater) episodeName(feedConfig *config.Feed, episode *model.Episode) string {
	ext := "mp4"
	if feedConfig.Format == model.FormatAudio {
		ext = "mp3"
	}

	return fmt.Sprintf("%s.%s", episode.ID, ext)
}

func (u *Updater) fileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

func (u *Updater) makeBuilder(ctx context.Context, cfg *config.Feed) (feed.Builder, error) {
	var (
		provider feed.Builder
		err      error
	)

	info, err := link.Parse(cfg.URL)
	if err != nil {
		return nil, err
	}

	switch info.Provider {
	case link.ProviderYoutube:
		provider, err = feed.NewYouTubeBuilder(u.config.Tokens.YouTube)
	case link.ProviderVimeo:
		provider, err = feed.NewVimeoBuilder(ctx, u.config.Tokens.Vimeo)
	default:
		return nil, errors.Errorf("unsupported provider %q", info.Provider)
	}

	return provider, err
}
