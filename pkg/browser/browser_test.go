package browser

import (
	"testing"
	"time"
)

func TestBrowser_SaveCookies(t *testing.T) {
	urls := []string{
		"https://www.ixigua.com/7082647768023433758",
		"https://www.douyin.com/video/7116100914905107719",
	}
	for _, u := range urls {
		t.Run(u, func(t *testing.T) {
			b, err := NewBrowser("ws://127.0.0.1:7317", time.Second*90)
			if err != nil {
				t.Errorf("NewBrowser error = %v", err)
			}
			if filePath, err := b.SaveCookies(u, false); err != nil {
				t.Errorf("Browser.SaveCookies() error = %v", err)
			} else {
				t.Logf("%v Cookies Saved to %v", u, filePath)
			}
		})
	}
}
