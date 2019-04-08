package builders

import (
	"net/http"
	"strconv"
	"time"

	itunes "github.com/mxpv/podcast"
	"github.com/pkg/errors"
	vimeo "github.com/silentsokolov/go-vimeo/vimeo"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/model"
)

const (
	vimeoDefaultPageSize = 50
)

type VimeoBuilder struct {
	client *vimeo.Client
}

func (v *VimeoBuilder) selectImage(p *vimeo.Pictures, q api.Quality) string {
	if p == nil || len(p.Sizes) == 0 {
		return ""
	}

	if q == api.QualityLow {
		return p.Sizes[0].Link
	} else {
		return p.Sizes[len(p.Sizes)-1].Link
	}
}

func (v *VimeoBuilder) queryChannel(feed *model.Feed) (*itunes.Podcast, error) {
	channelID := feed.ItemID

	ch, resp, err := v.client.Channels.Get(channelID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, api.ErrNotFound
		}

		return nil, errors.Wrapf(err, "failed to query channel with id %q", channelID)
	}

	podcast := itunes.New(ch.Name, ch.Link, ch.Description, &ch.CreatedTime, nil)
	podcast.Generator = podsyncGenerator
	podcast.AddSubTitle(ch.Name)
	podcast.AddImage(v.selectImage(ch.Pictures, feed.Quality))
	podcast.AddCategory(defaultCategory, nil)
	podcast.IAuthor = ch.User.Name

	feed.PubDate = ch.CreatedTime

	return &podcast, nil
}

func (v *VimeoBuilder) queryGroup(feed *model.Feed) (*itunes.Podcast, error) {
	groupID := feed.ItemID

	gr, resp, err := v.client.Groups.Get(groupID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, api.ErrNotFound
		}

		return nil, errors.Wrapf(err, "failed to query group with id %q", groupID)
	}

	podcast := itunes.New(gr.Name, gr.Link, gr.Description, &gr.CreatedTime, nil)
	podcast.Generator = podsyncGenerator
	podcast.AddSubTitle(gr.Name)
	podcast.AddImage(v.selectImage(gr.Pictures, feed.Quality))
	podcast.AddCategory(defaultCategory, nil)
	podcast.IAuthor = gr.User.Name

	feed.PubDate = gr.CreatedTime

	return &podcast, nil
}

func (v *VimeoBuilder) queryUser(feed *model.Feed) (*itunes.Podcast, error) {
	userID := feed.ItemID

	user, resp, err := v.client.Users.Get(userID)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, api.ErrNotFound
		}

		return nil, errors.Wrapf(err, "failed to query user with id %q", userID)
	}

	podcast := itunes.New(user.Name, user.Link, user.Bio, &user.CreatedTime, nil)
	podcast.Generator = podsyncGenerator
	podcast.AddSubTitle(user.Name)
	podcast.AddImage(v.selectImage(user.Pictures, feed.Quality))
	podcast.AddCategory(defaultCategory, nil)
	podcast.IAuthor = user.Name

	feed.PubDate = user.CreatedTime

	return &podcast, nil
}

func (v *VimeoBuilder) getVideoSize(video *vimeo.Video) int64 {
	// Very approximate video file size
	return int64(float64(video.Duration*video.Width*video.Height) * 0.38848958333)
}

type getVideosFunc func(string, ...vimeo.CallOption) ([]*vimeo.Video, *vimeo.Response, error)

func (v *VimeoBuilder) queryVideos(getVideos getVideosFunc, podcast *itunes.Podcast, feed *model.Feed) error {
	var (
		page  = 1
		added = 0
	)

	for {
		videos, response, err := getVideos(feed.ItemID, vimeo.OptPage(page), vimeo.OptPerPage(vimeoDefaultPageSize))
		if err != nil {
			if response != nil {
				return errors.Wrapf(err, "failed to query videos (error %d %s)", response.StatusCode, response.Status)
			}

			return err
		}

		for _, video := range videos {

			var (
				videoID  = strconv.Itoa(video.GetID())
				videoURL = video.Link
				duration = int64(video.Duration)
				size     = v.getVideoSize(video)
				image    = v.selectImage(video.Pictures, feed.Quality)
			)

			item := itunes.Item{}

			item.GUID = videoID
			item.Link = videoURL
			item.Title = video.Name
			item.Description = video.Description
			if item.Description == "" {
				item.Description = " " // Videos can be without description, workaround for AddItem
			}

			item.AddDuration(duration)
			item.AddPubDate(&video.CreatedTime)
			item.AddImage(image)

			item.AddEnclosure(makeEnclosure(feed, item.GUID, size))

			_, err = podcast.AddItem(item)
			if err != nil {
				return errors.Wrapf(err, "failed to add episode %s (%s)", item.GUID, item.Title)
			}

			feed.Episodes = append(feed.Episodes, &model.Item{
				ID:          videoID,
				Title:       item.Title,
				Description: item.Description,
				Duration:    duration,
				Size:        size,
				PubDate:     model.Timestamp(video.CreatedTime),
				Thumbnail:   image,
				VideoURL:    videoURL,
			})

			added++
		}

		if added >= feed.PageSize || response.NextPage == "" {
			return nil
		}

		page++
	}
}

func (v *VimeoBuilder) Build(feed *model.Feed) error {
	feed.Episodes = []*model.Item{}

	if feed.LinkType == api.LinkTypeChannel {
		podcast, err := v.queryChannel(feed)
		if err != nil {
			return err
		}

		if err := v.queryVideos(v.client.Channels.ListVideo, podcast, feed); err != nil {
			return err
		}

		v.updateFeed(feed, podcast)
		return nil
	}

	if feed.LinkType == api.LinkTypeGroup {
		podcast, err := v.queryGroup(feed)
		if err != nil {
			return err
		}

		if err := v.queryVideos(v.client.Groups.ListVideo, podcast, feed); err != nil {
			return err
		}

		v.updateFeed(feed, podcast)
		return nil
	}

	if feed.LinkType == api.LinkTypeUser {
		podcast, err := v.queryUser(feed)
		if err != nil {
			return err
		}

		if err := v.queryVideos(v.client.Users.ListVideo, podcast, feed); err != nil {
			return err
		}

		v.updateFeed(feed, podcast)
		return nil
	}

	return errors.New("unsupported feed type")
}

func (v *VimeoBuilder) updateFeed(feed *model.Feed, podcast *itunes.Podcast) {
	feed.Title = podcast.Title
	feed.Description = podcast.Description
	feed.Author = podcast.IAuthor
	feed.ItemURL = podcast.Link
	feed.UpdatedAt = time.Now().UTC()

	if podcast.IImage != nil {
		feed.CoverArt = podcast.IImage.HREF
	}

	if len(feed.Episodes) > 0 {
		feed.LastID = feed.Episodes[0].ID
	} else {
		feed.LastID = ""
	}
}

func (v *VimeoBuilder) GetVideoCount(feed *model.Feed) (uint64, error) {
	return 0, errors.New("not supported")
}

func NewVimeoBuilder(ctx context.Context, token string) (*VimeoBuilder, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	client := vimeo.NewClient(tc, nil)
	return &VimeoBuilder{client}, nil
}
