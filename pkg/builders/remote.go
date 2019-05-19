package builders

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/mxpv/podsync/pkg/model"
)

type Remote struct {
	url string
}

func NewRemote(url string) Remote {
	return Remote{url: url}
}

func (r Remote) Build(feed *model.Feed) error {
	addr, err := r.makeURL(feed)
	if err != nil {
		return err
	}

	client := http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Get(addr)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrapf(err, "failed to read error body (status: %d)", resp.StatusCode)
		}

		return errors.Errorf("unexpected response (%d) from updater: %q", resp.StatusCode, body)
	}

	var out responsePayload
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}

	feed.LastID = out.LastID
	feed.Episodes = append(out.Episodes, feed.Episodes...)
	feed.UpdatedAt = time.Now().UTC()

	return nil
}

func (r Remote) makeURL(feed *model.Feed) (string, error) {
	qs := url.Values{}
	qs.Add("url", feed.ItemURL)
	qs.Add("start", "1")
	qs.Add("count", strconv.Itoa(feed.PageSize))
	qs.Add("last_id", feed.LastID)
	qs.Add("format", string(feed.Format))
	qs.Add("quality", string(feed.Quality))

	parsed, err := url.Parse(r.url)
	if err != nil {
		return "", err
	}

	parsed.RawQuery = qs.Encode()

	return parsed.String(), nil
}
