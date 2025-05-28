package builder

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/model"
)

// ParseURL is a top-level parser
func ParseURL(link string) (model.Info, error) {
	parsed, err := parseURL(link)
	if err != nil {
		return model.Info{}, err // err from parseURL is already contextualized
	}

	info := model.Info{}

	if strings.HasSuffix(parsed.Host, "youtube.com") {
		kind, id, errP := parseYoutubeURL(parsed)
		if errP != nil {
			return model.Info{}, errors.Wrapf(errP, "failed to parse YouTube URL: %s", link)
		}

		info.Provider = model.ProviderYoutube
		info.LinkType = kind
		info.ItemID = id

		return info, nil
	}

	if strings.HasSuffix(parsed.Host, "vimeo.com") {
		kind, id, errP := parseVimeoURL(parsed)
		if errP != nil {
			return model.Info{}, errors.Wrapf(errP, "failed to parse Vimeo URL: %s", link)
		}

		info.Provider = model.ProviderVimeo
		info.LinkType = kind
		info.ItemID = id

		return info, nil
	}

	if strings.HasSuffix(parsed.Host, "soundcloud.com") {
		kind, id, errP := parseSoundcloudURL(parsed)
		if errP != nil {
			return model.Info{}, errors.Wrapf(errP, "failed to parse SoundCloud URL: %s", link)
		}

		info.Provider = model.ProviderSoundcloud
		info.LinkType = kind
		info.ItemID = id

		return info, nil
	}

	return model.Info{}, errors.Errorf("unsupported URL host: %s", parsed.Host)
}

func parseURL(link string) (*url.URL, error) {
	// Ensure the link has a scheme for correct parsing by url.Parse
	if !strings.HasPrefix(link, "http://") && !strings.HasPrefix(link, "https://") {
		link = "https://" + link
	}

	parsed, err := url.Parse(link)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to perform URL parsing for link: %s", link)
	}

	return parsed, nil
}

func parseYoutubeURL(parsed *url.URL) (model.Type, string, error) {
	path := parsed.EscapedPath()
	trimmedPath := strings.Trim(path, "/")
	parts := strings.Split(trimmedPath, "/")

	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		// This handles cases like "youtube.com/" or "youtube.com"
		return "", "", errors.New("youtube URL path is empty or just a slash")
	}

	firstPart := parts[0]

	// Order of checks: Most specific query param based, then path based.
	// https://www.youtube.com/playlist?list=ID
	// https://www.youtube.com/watch?v=VIDEOID&list=PLAYLISTID
	if firstPart == "playlist" || firstPart == "watch" {
		playlistID := parsed.Query().Get("list")
		if playlistID != "" {
			return model.TypePlaylist, playlistID, nil
		}
		if firstPart == "watch" {
			return "", "", errors.New("watch URL without a 'list' query parameter is not a supported feed type")
		}
		// If it's "playlist" but no list ID, it's an invalid playlist URL for feed purposes
		return "", "", errors.New("playlist URL is missing 'list' query parameter")
	}

	// https://www.youtube.com/channel/CHANNEL_ID or /channel/CHANNEL_ID/videos etc.
	if firstPart == "channel" {
		if len(parts) < 2 || parts[1] == "" {
			return "", "", errors.New("invalid youtube channel link: missing channel ID part")
		}
		channelID := parts[1]
		return model.TypeChannel, channelID, nil
	}

	// https://www.youtube.com/user/USERNAME or /user/USERNAME/videos etc.
	if firstPart == "user" {
		if len(parts) < 2 || parts[1] == "" {
			return "", "", errors.New("invalid youtube user link: missing username part")
		}
		userName := parts[1]
		return model.TypeUser, userName, nil // TypeUser will be resolved to a Channel ID by the API caller
	}

	// New @handle style URLs:
	// https://www.youtube.com/@HANDLE
	// https://www.youtube.com/@HANDLE/videos
	// https://www.youtube.com/@HANDLE/playlists
	// https://www.youtube.com/@HANDLE/shorts  (consider if this is a feed type we'd want)
	// https://www.youtube.com/@HANDLE/streams (could be a feed type)
	// https://www.youtube.com/@HANDLE/podcasts (could be a feed type)
	if strings.HasPrefix(firstPart, "@") {
		handle := firstPart // e.g., "@AndrejKarpathy"
		if len(parts) == 1 {
			// Just youtube.com/@handle, treat as a channel's main feed (uploads/videos)
			return model.TypeChannel, handle, nil
		}

		// We have a suffix like /videos, /playlists, /shorts, /streams, /podcasts
		suffix := strings.ToLower(parts[1])
		switch suffix {
		case "videos":
			return model.TypeChannel, handle, nil // Effectively the channel's video feed
		case "playlists":
			// This implies we want to list available playlists for the user to choose,
			// or it could mean a special "all public playlists merged" feed (complex).
			// For now, per user request, this should point to a feed.
			// Let's assume it means the "uploads" playlist of the channel represented by the handle.
			// So, we'll treat it as TypeChannel, and the youtube.go logic will fetch the uploads playlist.
			// Or, if the intention is to allow *any* playlist from that channel, this Type should be Playlist,
			// and the ItemID should be the handle. The youtube.go logic would then need to figure out
			// WHICH playlist. Given the example `/@AndrejKarpathy/playlists` from the issue,
			// it's likely intended to be the collection of his playlists, or his primary one.
			// Let's return TypePlaylist and the handle. The next step will clarify how youtube.go consumes this.
			return model.TypePlaylist, handle, nil
		case "shorts", "streams", "podcasts":
			// These are specific content types. Whether they map to a single feed or need special handling
			// depends on how YouTube structures their data via API for these.
			// For now, let's treat them as a channel type, implying the main feed of that content.
			// This might need refinement in youtube.go.
			// Example: treat @handle/shorts as the "shorts" feed for that channel.
			// This would be a TypeChannel, and ItemID handle. The builder would find the "Shorts" playlist.
			// However, the original issue only mentioned /videos and /playlists.
			// Let's stick to those and consider others unsupported for now to limit scope.
			return "", "", fmt.Errorf("unsupported youtube handle link suffix: /%s for handle %s (shorts, streams, podcasts not yet supported)", suffix, handle)
		default:
			// Suffixes like /community, /channels, /about are not media feeds.
			return "", "", fmt.Errorf("unsupported youtube handle link suffix: /%s for handle %s", suffix, handle)
		}
	}

	return "", "", errors.New("unsupported youtube link format")
}

func parseVimeoURL(parsed *url.URL) (model.Type, string, error) {
	trimmedPath := strings.Trim(parsed.EscapedPath(), "/")
	parts := strings.Split(trimmedPath, "/")

	if len(parts) == 0 || parts[0] == "" {
		return "", "", errors.New("invalid vimeo link path: path is empty")
	}

	firstPart := strings.ToLower(parts[0])
	var kind model.Type
	var id string

	switch firstPart {
	case "groups":
		if len(parts) < 2 || parts[1] == "" {
			return "", "", errors.New("invalid vimeo group link: missing group ID")
		}
		kind = model.TypeGroup
		id = parts[1]
	case "channels":
		if len(parts) < 2 || parts[1] == "" {
			return "", "", errors.New("invalid vimeo channel link: missing channel ID")
		}
		kind = model.TypeChannel
		id = parts[1]
	default:
		// Assumes a user URL like vimeo.com/username or vimeo.com/username/videos
		// The first part is the user identifier.
		kind = model.TypeUser
		id = parts[0]
	}
	return kind, id, nil
}

func parseSoundcloudURL(parsed *url.URL) (model.Type, string, error) {
	trimmedPath := strings.Trim(parsed.EscapedPath(), "/")
	parts := strings.Split(trimmedPath, "/")

	// Expected format: soundcloud.com/USERNAME/sets/PLAYLIST_NAME
	// parts[0] = USERNAME, parts[1] = "sets", parts[2] = PLAYLIST_NAME
	if len(parts) < 3 || strings.ToLower(parts[1]) != "sets" || parts[2] == "" {
		return "", "", errors.New("invalid soundcloud link path, expected format like /username/sets/playlist_name")
	}
	// The playlist identifier that podsync uses is the playlist_name (parts[2])
	return model.TypePlaylist, parts[2], nil
}
