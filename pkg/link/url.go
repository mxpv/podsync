package link

import (
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func Parse(link string) (Info, error) {
	parsed, err := parseURL(link)
	if err != nil {
		return Info{}, err
	}

	info := Info{}

	if strings.HasSuffix(parsed.Host, "youtube.com") {
		kind, id, err := parseYoutubeURL(parsed)
		if err != nil {
			return Info{}, err
		}

		info.Provider = ProviderYoutube
		info.LinkType = kind
		info.ItemID = id

		return info, nil
	}

	if strings.HasSuffix(parsed.Host, "vimeo.com") {
		kind, id, err := parseVimeoURL(parsed)
		if err != nil {
			return Info{}, err
		}

		info.Provider = ProviderVimeo
		info.LinkType = kind
		info.ItemID = id

		return info, nil
	}

	return Info{}, errors.New("unsupported URL host")
}

func parseURL(link string) (*url.URL, error) {
	if !strings.HasPrefix(link, "http") {
		link = "https://" + link
	}

	parsed, err := url.Parse(link)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse url: %s", link)
	}

	return parsed, nil
}

func parseYoutubeURL(parsed *url.URL) (Type, string, error) {
	path := parsed.EscapedPath()

	// https://www.youtube.com/playlist?list=PLCB9F975ECF01953C
	// https://www.youtube.com/watch?v=rbCbho7aLYw&list=PLMpEfaKcGjpWEgNtdnsvLX6LzQL0UC0EM
	if strings.HasPrefix(path, "/playlist") || strings.HasPrefix(path, "/watch") {
		kind := TypePlaylist

		id := parsed.Query().Get("list")
		if id != "" {
			return kind, id, nil
		}

		return "", "", errors.New("invalid playlist link")
	}

	// - https://www.youtube.com/channel/UC5XPnUk8Vvv_pWslhwom6Og
	// - https://www.youtube.com/channel/UCrlakW-ewUT8sOod6Wmzyow/videos
	if strings.HasPrefix(path, "/channel") {
		kind := TypeChannel
		parts := strings.Split(parsed.EscapedPath(), "/")
		if len(parts) <= 2 {
			return "", "", errors.New("invalid youtube channel link")
		}

		id := parts[2]
		if id == "" {
			return "", "", errors.New("invalid id")
		}

		return kind, id, nil
	}

	// - https://www.youtube.com/user/fxigr1
	if strings.HasPrefix(path, "/user") {
		kind := TypeUser

		parts := strings.Split(parsed.EscapedPath(), "/")
		if len(parts) <= 2 {
			return "", "", errors.New("invalid user link")
		}

		id := parts[2]
		if id == "" {
			return "", "", errors.New("invalid id")
		}

		return kind, id, nil
	}

	return "", "", errors.New("unsupported link format")
}

func parseVimeoURL(parsed *url.URL) (Type, string, error) {
	parts := strings.Split(parsed.EscapedPath(), "/")
	if len(parts) <= 1 {
		return "", "", errors.New("invalid vimeo link path")
	}

	var kind Type
	switch parts[1] {
	case "groups":
		kind = TypeGroup
	case "channels":
		kind = TypeChannel
	default:
		kind = TypeUser
	}

	if kind == TypeGroup || kind == TypeChannel {
		if len(parts) <= 2 {
			return "", "", errors.New("invalid channel link")
		}

		id := parts[2]
		if id == "" {
			return "", "", errors.New("invalid id")
		}

		return kind, id, nil
	}

	if kind == TypeUser {
		id := parts[1]
		if id == "" {
			return "", "", errors.New("invalid id")
		}

		return kind, id, nil
	}

	return "", "", errors.New("unsupported link format")
}
