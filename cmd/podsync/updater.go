package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	itunes "github.com/mxpv/podcast"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/db"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/mxpv/podsync/pkg/link"
	"github.com/mxpv/podsync/pkg/model"
	"github.com/mxpv/podsync/pkg/ytdl"
)

type Downloader interface {
	Download(ctx context.Context, feedConfig *config.Feed, episode *model.Episode) (io.ReadCloser, error)
}

type Updater struct {
	config     *config.Config
	downloader Downloader
	db         db.Storage
}

func NewUpdater(config *config.Config, downloader Downloader, db db.Storage) (*Updater, error) {
	return &Updater{
		config:     config,
		downloader: downloader,
		db:         db,
	}, nil
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

	if err := u.updateFeed(ctx, feedConfig); err != nil {
		return err
	}

	if err := u.downloadEpisodes(ctx, feedConfig, feedPath); err != nil {
		return err
	}

	if err := u.buildXML(ctx, feedConfig); err != nil {
		return err
	}

	elapsed := time.Since(started)
	nextUpdate := time.Now().Add(feedConfig.UpdatePeriod.Duration)
	log.Infof("successfully updated feed in %s, next update at %s", elapsed, nextUpdate.Format(time.Kitchen))
	return nil
}

// updateFeed pulls API for new episodes and saves them to database
func (u *Updater) updateFeed(ctx context.Context, feedConfig *config.Feed) error {
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

	if err := u.db.AddFeed(ctx, feedConfig.ID, result); err != nil {
		return err
	}

	log.Debug("successfully saved updates to storage")
	return nil
}

func (u *Updater) downloadEpisodes(ctx context.Context, feedConfig *config.Feed, targetDir string) error {
	var (
		feedID       = feedConfig.ID
		downloadList []*model.Episode
	)

	// Build the list of files to download
	if err := u.db.WalkEpisodes(ctx, feedID, func(episode *model.Episode) error {
		if episode.Status != model.EpisodeNew && episode.Status != model.EpisodeError {
			// File already downloaded
			return nil
		}

		if feedConfig.Filters.Title != "" {
			matched, err := regexp.MatchString(feedConfig.Filters.Title, episode.Title)
			if err != nil {
				log.Warnf("Pattern '%s' is not a valid filter for %s Title", feedConfig.Filters.Title, feedConfig.ID)
			} else {
				if !matched {
					log.Infof("Skipping '%s' due to lack of match with '%s'", episode.Title, feedConfig.Filters.Title)
					return nil
				}
			}
		}

		downloadList = append(downloadList, episode)
		return nil
	}); err != nil {
		return errors.Wrapf(err, "failed to build update list")
	}

	var (
		downloadCount = len(downloadList)
		downloaded    = 0
	)

	if downloadCount > 0 {
		log.Infof("download count: %d", downloadCount)
	} else {
		log.Info("no episodes to download")
		return nil
	}

	// Download pending episodes

	for idx, episode := range downloadList {
		logger := log.WithFields(log.Fields{
			"index":      idx,
			"episode_id": episode.ID,
		})

		// Check whether episode exists on disk

		episodePath := filepath.Join(targetDir, u.episodeName(feedConfig, episode))
		stat, err := os.Stat(episodePath)
		if err == nil {
			logger.Infof("episode %q already exists on disk (%s)", episode.ID, episodePath)

			// File already exists, update file status and disk size
			if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
				episode.Size = stat.Size()
				episode.Status = model.EpisodeDownloaded
				return nil
			}); err != nil {
				logger.WithError(err).Error("failed to update file info")
				return err
			}

			continue
		} else if os.IsNotExist(err) {
			// Will download, do nothing here
		} else {
			logger.WithError(err).Error("failed to stat file")
			return err
		}

		// Download episode to disk
		// We download the episode to a temp directory first to avoid downloading this file by clients
		// while still being processed by youtube-dl (e.g. a file is being downloaded from YT or encoding in progress)

		logger.Infof("! downloading episode %s", episode.VideoURL)
		tempFile, err := u.downloader.Download(ctx, feedConfig, episode)
		if err != nil {
			// YouTube might block host with HTTP Error 429: Too Many Requests
			// We still need to generate XML, so just stop sending download requests and
			// retry next time
			if err == ytdl.ErrTooManyRequests {
				break
			}

			if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
				episode.Status = model.EpisodeError
				return nil
			}); err != nil {
				return err
			}

			continue
		}

		logger.Debugf("copying file to: %s", episodePath)
		fileSize, err := copyFile(tempFile, episodePath)

		tempFile.Close()

		if err != nil {
			logger.WithError(err).Error("failed to copy file")
			return err
		}
		logger.Debugf("copied %d bytes", episode.Size)

		// Update file status in database

		logger.Infof("successfully downloaded file %q", episode.ID)
		if err := u.db.UpdateEpisode(feedID, episode.ID, func(episode *model.Episode) error {
			episode.Size = fileSize
			episode.Status = model.EpisodeDownloaded
			return nil
		}); err != nil {
			return err
		}

		downloaded++
	}

	log.Infof("downloaded %d episode(s)", downloaded)
	return nil
}

func copyFile(source io.Reader, destinationPath string) (int64, error) {
	dest, err := os.Create(destinationPath)
	if err != nil {
		return 0, errors.Wrap(err, "failed to create destination file")
	}

	defer dest.Close()

	written, err := io.Copy(dest, source)
	if err != nil {
		return 0, errors.Wrap(err, "failed to copy data")
	}

	return written, nil
}

func (u *Updater) buildXML(ctx context.Context, feedConfig *config.Feed) error {
	feed, err := u.db.GetFeed(ctx, feedConfig.ID)
	if err != nil {
		return err
	}

	// Build iTunes XML feed with data received from builder
	log.Debug("building iTunes podcast feed")
	podcast, err := u.buildPodcast(feed, feedConfig)
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
		if episode.Status != model.EpisodeDownloaded {
			// Skip episodes that are not yet downloaded
			continue
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
