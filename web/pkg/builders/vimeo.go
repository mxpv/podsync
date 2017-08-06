package builders

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/web/pkg/storage"
	"github.com/pkg/errors"
	"github.com/silentsokolov/go-vimeo"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	vimeoDefaultPageSize = 50
)

type VimeoBuilder struct {
	client *vimeo.Client
}

func (v *VimeoBuilder) parseUrl(link string) (kind linkType, id string, err error) {
	parsed, err := url.Parse(link)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse url: %s", link)
		return
	}

	if !strings.HasSuffix(parsed.Host, "vimeo.com") {
		err = errors.New("invalid vimeo host")
		return
	}

	parts := strings.Split(parsed.EscapedPath(), "/")

	if len(parts) <= 1 {
		err = errors.New("invalid vimeo link path")
		return
	}

	if parts[1] == "groups" {
		kind = linkTypeGroup
	} else if parts[1] == "channels" {
		kind = linkTypeChannel
	} else {
		kind = linkTypeUser
	}

	if kind == linkTypeGroup || kind == linkTypeChannel {
		if len(parts) <= 2 {
			err = errors.New("invalid channel link")
			return
		}

		id = parts[2]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	if kind == linkTypeUser {
		id = parts[1]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	err = errors.New("unsupported link format")
	return
}

func (v *VimeoBuilder) selectImage(p *vimeo.Pictures, q storage.Quality) string {
	if p == nil || len(p.Sizes) < 1 {
		return ""
	}

	if q == storage.LowQuality {
		return p.Sizes[0].Link
	} else {
		return p.Sizes[len(p.Sizes)-1].Link
	}
}

func (v *VimeoBuilder) queryChannel(channelId string, feed *storage.Feed) (*itunes.Podcast, error) {
	ch, resp, err := v.client.Channels.Get(channelId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query channel with channelId %s", channelId)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid http response from video server")
	}

	podcast := itunes.New(ch.Name, ch.Link, ch.Description, &ch.CreatedTime, nil)
	podcast.Generator = podsyncGenerator
	podcast.AddSubTitle(ch.Name)
	podcast.AddImage(v.selectImage(ch.Pictures, feed.Quality))
	podcast.AddCategory(defaultCategory, nil)
	podcast.IAuthor = ch.User.Name

	return &podcast, nil
}

func (v *VimeoBuilder) queryGroup(groupId string, feed *storage.Feed) (*itunes.Podcast, error) {
	gr, resp, err := v.client.Groups.Get(groupId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query group with id %s", groupId)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid http response from video server")
	}

	podcast := itunes.New(gr.Name, gr.Link, gr.Description, &gr.CreatedTime, nil)
	podcast.Generator = podsyncGenerator
	podcast.AddSubTitle(gr.Name)
	podcast.AddImage(v.selectImage(gr.Pictures, feed.Quality))
	podcast.AddCategory(defaultCategory, nil)
	podcast.IAuthor = gr.User.Name

	return &podcast, nil
}

func (v *VimeoBuilder) queryUser(userId string, feed *storage.Feed) (*itunes.Podcast, error) {
	user, resp, err := v.client.Users.Get(userId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to query user with id %s", userId)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("invalid http response from video server")
	}

	podcast := itunes.New(user.Name, user.Link, user.Bio, &user.CreatedTime, nil)
	podcast.Generator = podsyncGenerator
	podcast.AddSubTitle(user.Name)
	podcast.AddImage(v.selectImage(user.Pictures, feed.Quality))
	podcast.AddCategory(defaultCategory, nil)
	podcast.IAuthor = user.Name

	return &podcast, nil
}

func (v *VimeoBuilder) getVideoSize(video *vimeo.Video) int64 {
	// Very approximate video file size
	return int64(float64(video.Duration*video.Width*video.Height) * 0.38848958333)
}

type queryVideosFunc func(id string, opt *vimeo.ListVideoOptions) ([]*vimeo.Video, *vimeo.Response, error)

func (v *VimeoBuilder) queryVideos(queryVideos queryVideosFunc, id string, podcast *itunes.Podcast, feed *storage.Feed) error {
	opt := vimeo.ListVideoOptions{}
	opt.Page = 1
	opt.PerPage = vimeoDefaultPageSize

	added := 0

	for {
		videos, response, err := queryVideos(id, &opt)
		if err != nil {
			return errors.Wrap(err, "failed to query videos")
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("invalid http response %d from vimeo: %s", response.StatusCode, response.Status)
		}

		for _, video := range videos {
			item := itunes.Item{}

			item.GUID = strconv.Itoa(video.GetID())
			item.Link = video.Link
			item.Title = video.Name
			item.Description = video.Description
			if item.Description == "" {
				item.Description = " " // Videos can be without description, workaround for AddItem
			}

			item.AddDuration(int64(video.Duration))
			item.AddPubDate(&video.CreatedTime)
			item.AddImage(v.selectImage(video.Pictures, feed.Quality))

			size := v.getVideoSize(video)
			item.AddEnclosure(makeEnclosure(feed, item.GUID, size))

			_, err = podcast.AddItem(item)
			if err != nil {
				return errors.Wrapf(err, "failed to add episode")
			}

			added++
		}

		if added >= feed.PageSize || response.NextPage == "" {
			return nil
		}

		opt.Page++
	}
}

func (v *VimeoBuilder) Build(feed *storage.Feed) (podcast *itunes.Podcast, err error) {
	kind, id, err := v.parseUrl(feed.URL)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse link: %s", feed.URL)
	}

	if kind == linkTypeChannel {
		if podcast, err = v.queryChannel(id, feed); err == nil {
			err = v.queryVideos(v.client.Channels.ListVideo, id, podcast, feed)
		}
	} else if kind == linkTypeGroup {
		if podcast, err = v.queryGroup(id, feed); err == nil {
			err = v.queryVideos(v.client.Groups.ListVideo, id, podcast, feed)
		}
	} else if kind == linkTypeUser {
		if podcast, err = v.queryUser(id, feed); err == nil {
			err = v.queryVideos(v.client.Users.ListVideo, id, podcast, feed)
		}
	} else {
		err = errors.New("unsupported feed type")
	}

	return
}

func NewVimeoBuilder(ctx context.Context, token string) (*VimeoBuilder, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client := vimeo.NewClient(tc)
	return &VimeoBuilder{client}, nil
}
