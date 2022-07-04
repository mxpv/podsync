package browser

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
	log "github.com/sirupsen/logrus"
	netscapecookiejar "github.com/vanym/golang-netscape-cookiejar"
)

type Browser struct {
	rod     *rod.Browser
	timeout time.Duration
	dir     string
}

func NewBrowser(serviceUrl string, timeout time.Duration) (browser *Browser, err error) {
	l, err := launcher.NewManaged(serviceUrl)
	if err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("rod connect error %v", err)
		}
	}()
	r := rod.New().Client(l.MustClient()).Timeout(timeout)
	err = r.Connect()
	if err != nil {
		return
	}
	b := &Browser{
		rod:     r,
		timeout: timeout,
	}
	tmpDir, err := ioutil.TempDir("", "podsync-cookies-")
	if err != nil {
		return
	}
	log.Infof("rod connected to %s", serviceUrl)
	b.dir = tmpDir
	return b, nil
}

func (b *Browser) Close() {
	b.rod.Close()
	os.RemoveAll(b.dir)
}

func (b *Browser) SaveCookies(pageUrl string, noPlayer bool) (filePath string, err error) {
	parsedUrl, err := url.Parse(pageUrl)
	if err != nil {
		return
	}
	filePath = b.dir + "/" + parsedUrl.Hostname() + "_cookies.txt"
	jar, err := netscapecookiejar.New(&netscapecookiejar.Options{
		WriteHeader: true,
	})
	if err != nil {
		return
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return
	}
	defer file.Close()
	_, err = jar.ReadFrom(file)
	if err == nil {
		expired := false
		cookies := jar.Cookies(parsedUrl)
		if len(cookies) > 0 {
			for _, c := range cookies {
				if c.Expires.After(time.Unix(0, 0)) && c.Expires.Before(time.Now()) {
					expired = true
					break
				}
			}
			if !expired {
				return
			}
		}
	}
	page, err := stealth.Page(b.rod)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), b.timeout)
	defer cancel()
	page = page.Context(ctx)
	defer page.Close()
	err = page.Navigate(pageUrl)
	if err != nil {
		return
	}
	if !noPlayer {
		wait := page.WaitEvent(proto.MediaPlayerEventsAdded{})
		wait()
	} else {
		err = page.WaitIdle(time.Second * 5)
		if err != nil {
			return
		}
	}
	cookies, err := page.Cookies([]string{})
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	cookiesTarget := []*http.Cookie{}

	for _, cookie := range cookies {
		c := &http.Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			HttpOnly: cookie.HTTPOnly,
			Domain:   cookie.Domain,
			Secure:   cookie.Secure,
		}
		if cookie.Expires > 0 {
			c.Expires = cookie.Expires.Time()
		} else {
			c.Expires = time.Unix(0, 0)
		}
		cookiesTarget = append(cookiesTarget, c)
	}
	jar.SetCookies(parsedUrl, cookiesTarget)
	_, err = jar.WriteTo(file)
	log.Infof("cookies saved to %s", filePath)
	return
}
