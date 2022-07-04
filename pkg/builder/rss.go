package builder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/mxpv/podsync/pkg/feed"
	"github.com/pkg/errors"

	"net/http"
	"net/url"

	"github.com/mxpv/podsync/pkg/model"

	"github.com/PuerkitoBio/goquery"
)

type RssBuilder struct{}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func iconForUrl(link string) (icon string, err error) {
	res, err := http.Get(link)
	if err != nil {
		return
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		err = errors.New(fmt.Sprintf("page status %d", res.StatusCode))
		return
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return
	}

	interestedItemProps := []string{"image", "thumbnailUrl"}

	interestedMetaProperties := []string{"og:image"}

	doc.Find("head meta").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if prop, ok := s.Attr("itemprop"); ok && contains(interestedItemProps, prop) {
			if content, ok := s.Attr("content"); ok && strings.HasPrefix(content, "http") {
				icon = content
				return false
			}
		}
		if prop, ok := s.Attr("property"); ok && contains(interestedMetaProperties, prop) {
			if content, ok := s.Attr("content"); ok && strings.HasPrefix(content, "http") {
				icon = content
				return false
			}
		}
		return true
	})
	if icon != "" {
		return
	}
	interestedRels := []string{"apple-touch-icon"}
	doc.Find("head link").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if rel, ok := s.Attr("rel"); ok && contains(interestedRels, rel) {
			if href, ok := s.Attr("href"); ok {
				icon = href
				return false
			}
		}
		return true
	})
	if icon != "" {
		if strings.HasPrefix(icon, "//") {
			u, e := url.Parse(link)
			if e == nil {
				icon = u.Scheme + ":" + icon
			}
		}
	}
	if icon == "" {
		err = errors.New("icon not found")
	}
	return
}

func (s *RssBuilder) Build(_ctx context.Context, cfg *feed.Config) (*model.Feed, error) {
	info, err := ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	feed := &model.Feed{
		ItemID:    info.ItemID,
		Provider:  info.Provider,
		LinkType:  info.LinkType,
		Format:    cfg.Format,
		Quality:   cfg.Quality,
		PageSize:  cfg.PageSize,
		UpdatedAt: time.Now().UTC(),
	}

	fp := gofeed.NewParser()
	rss, err := fp.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}

	feed.Title = rss.Title
	feed.Description = rss.Description
	feed.ItemURL = rss.Link
	if rss.PublishedParsed != nil {
		feed.PubDate = *rss.PublishedParsed
	}
	if rss.Author != nil {
		feed.Author = rss.Author.Name
	}
	if rss.Image != nil {
		feed.CoverArt = rss.Image.URL
	}
	if feed.CoverArt == "" {
		icon, err := iconForUrl(rss.Link)
		if err == nil {
			feed.CoverArt = icon
		}
	}

	added := 0

	for _, item := range rss.Items {
		link, err := url.Parse(item.Link)
		if err != nil {
			continue
		}
		_, videoID, err := parseOtherUrl(link)
		if err != nil {
			continue
		}
		episode := model.Episode{
			ID:          videoID,
			Title:       item.Title,
			Description: item.Description,
			VideoURL:    item.Link,
			PubDate:     *item.PublishedParsed,
			Status:      model.EpisodeNew,
		}
		if item.Image != nil {
			episode.Thumbnail = item.Image.URL
		} else {
			icon, _ := iconForUrl(item.Link)
			episode.Thumbnail = icon
		}
		feed.Episodes = append(feed.Episodes, &episode)
		added++
		if added >= feed.PageSize {
			return feed, nil
		}
	}
	if len(feed.Episodes) > 0 {
		return feed, nil
	} else {
		return nil, errors.New(fmt.Sprintf("unsupported rss feed type %v", err))
	}
}

func NewRssBuilder() (*RssBuilder, error) {
	return &RssBuilder{}, nil
}
