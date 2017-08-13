package parsers

import (
	"net/url"
	"strings"

	"github.com/mxpv/podsync/web/pkg/api"
	"github.com/pkg/errors"
)

func ParseURL(link string) (*api.Feed, error) {
	parsed, err := url.Parse(link)
	if err != nil {
		err = errors.Wrapf(err, "failed to parse url: %s", link)
		return nil, err
	}

	feed := &api.Feed{}

	if strings.HasSuffix(parsed.Host, "youtube.com") {
		kind, id, err := parseYoutubeURL(parsed)
		if err != nil {
			return nil, err
		}

		feed.Provider = api.Youtube
		feed.LinkType = kind
		feed.ItemId = id

		return feed, nil
	}

	if strings.HasSuffix(parsed.Host, "vimeo.com") {
		kind, id, err := parseVimeoURL(parsed)
		if err != nil {
			return nil, err
		}

		feed.Provider = api.Vimeo
		feed.LinkType = kind
		feed.ItemId = id

		return feed, nil
	}

	return nil, errors.New("unsupported URL host")
}

func parseYoutubeURL(parsed *url.URL) (kind api.LinkType, id string, err error) {
	path := parsed.EscapedPath()

	// https://www.youtube.com/playlist?list=PLCB9F975ECF01953C
	if strings.HasPrefix(path, "/playlist") {
		kind = api.Playlist

		id = parsed.Query().Get("list")
		if id != "" {
			return
		}

		err = errors.New("invalid playlist link")
		return
	}

	// - https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og
	// - https://www.youtube.com/channel/UCrlakW-ewUT8sOod6Wmzyow/videos
	if strings.HasPrefix(path, "/channel") {
		kind = api.Channel
		parts := strings.Split(parsed.EscapedPath(), "/")
		if len(parts) <= 2 {
			err = errors.New("invalid youtube channel link")
			return
		}

		id = parts[2]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	// - https://www.youtube.com/user/fxigr1
	if strings.HasPrefix(path, "/user") {
		kind = api.User

		parts := strings.Split(parsed.EscapedPath(), "/")
		if len(parts) <= 2 {
			err = errors.New("invalid user link")
			return
		}

		id = parts[2]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	err = errors.New("unsupported link format")
	return
}

func parseVimeoURL(parsed *url.URL) (kind api.LinkType, id string, err error) {
	parts := strings.Split(parsed.EscapedPath(), "/")

	if len(parts) <= 1 {
		err = errors.New("invalid vimeo link path")
		return
	}

	if parts[1] == "groups" {
		kind = api.Group
	} else if parts[1] == "channels" {
		kind = api.Channel
	} else {
		kind = api.User
	}

	if kind == api.Group || kind == api.Channel {
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

	if kind == api.User {
		id = parts[1]
		if id == "" {
			err = errors.New("invalid id")
		}

		return
	}

	err = errors.New("unsupported link format")
	return
}
