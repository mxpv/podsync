package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"go/build"
	"log"
	"net/http"
	"path"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/mxpv/patreon-go"
	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/config"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	campaignId         = "278915"
	identitySessionKey = "identity"
)

type feed interface {
	CreateFeed(req *api.CreateFeedRequest, identity *api.Identity) (string, error)
	GetFeed(hashId string) (*itunes.Podcast, error)
	GetMetadata(hashId string) (*api.Feed, error)
}

func MakeHandlers(feed feed, cfg *config.AppConfig) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	store := sessions.NewCookieStore([]byte(cfg.CookieSecret))
	r.Use(sessions.Sessions("podsync", store))

	// Static files + HTML

	conf := &oauth2.Config{
		ClientID:     cfg.PatreonClientId,
		ClientSecret: cfg.PatreonSecret,
		RedirectURL:  "http://localhost:8080/patreon",
		Scopes:       []string{"users", "pledges-to-me", "my-campaign"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  patreon.AuthorizationURL,
			TokenURL: patreon.AccessTokenURL,
		},
	}

	rootDir := path.Join(build.Default.GOPATH, "src/github.com/mxpv/podsync")
	log.Printf("Using root directory: %s", rootDir)

	r.Static("/assets", path.Join(rootDir, "assets"))
	r.LoadHTMLGlob(path.Join(rootDir, "templates/*.html"))

	r.GET("/", func(c *gin.Context) {
		s := sessions.Default(c)

		identity := &api.Identity{
			FeatureLevel: api.DefaultFeatures,
		}

		buf, ok := s.Get(identitySessionKey).(string)
		if ok {
			// We are failed to deserialize Identity structure, do cleanup, force user to login again
			if err := json.Unmarshal([]byte(buf), identity); err != nil {
				s.Clear()
				s.Save()
			}
		}

		c.HTML(http.StatusOK, "index.html", identity)
	})

	r.GET("/login", func(c *gin.Context) {
		state := randToken()

		s := sessions.Default(c)
		s.Set("state", state)
		s.Save()

		authURL := conf.AuthCodeURL(state)
		c.Redirect(http.StatusFound, authURL)
	})

	r.GET("/logout", func(c *gin.Context) {
		s := sessions.Default(c)
		s.Clear()
		s.Save()

		c.Redirect(http.StatusFound, "/")
	})

	r.GET("/patreon", func(c *gin.Context) {
		// Validate session state
		s := sessions.Default(c)
		state := s.Get("state")
		if state != c.Query("state") {
			c.String(http.StatusUnauthorized, "invalid state")
			return
		}

		// Exchange code with tokens
		token, err := conf.Exchange(c.Request.Context(), c.Query("code"))
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		// Create Patreon client
		tc := conf.Client(c.Request.Context(), token)
		client := patreon.NewClient(tc)

		// Query user info from Patreon
		user, err := client.FetchUser()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		// Determine feature level
		level := api.DefaultFeatures
		amount := 0
		for _, item := range user.Included.Items {
			pledge, ok := item.(*patreon.Pledge)
			if ok {
				amount += pledge.Attributes.AmountCents
			}
		}

		if amount >= 100 {
			level = api.ExtendedFeatures
		}

		identity := &api.Identity{
			UserId:       user.Data.Id,
			FullName:     user.Data.Attributes.FullName,
			Email:        user.Data.Attributes.Email,
			ProfileURL:   user.Data.Attributes.URL,
			FeatureLevel: level,
		}

		// Serialize identity and return cookies
		buf, err := json.Marshal(identity)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		s.Clear()
		s.Set(identitySessionKey, string(buf))
		s.Save()

		c.Redirect(http.StatusFound, "/")
	})

	// REST API

	r.GET("/api/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.POST("/api/create", func(c *gin.Context) {
		req := &api.CreateFeedRequest{}

		if err := c.BindJSON(req); err != nil {
			c.JSON(badRequest(err))
			return
		}

		s := sessions.Default(c)

		identity := &api.Identity{
			FeatureLevel: api.DefaultFeatures,
		}

		buf, ok := s.Get(identitySessionKey).(string)
		if ok {
			// We are failed to deserialize Identity structure, do cleanup, force user to login again
			if err := json.Unmarshal([]byte(buf), identity); err != nil {
				s.Clear()
				s.Save()
			}
		}

		hashId, err := feed.CreateFeed(req, identity)
		if err != nil {
			c.JSON(internalError(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": hashId})
	})

	r.GET("/api/feed/:hashId", func(c *gin.Context) {
		hashId := c.Param("hashId")
		if hashId == "" || len(hashId) > 12 {
			c.JSON(badRequest(errors.New("invalid feed id")))
			return
		}

		podcast, err := feed.GetFeed(hashId)
		if err != nil {
			c.JSON(internalError(err))
			return
		}

		c.Data(http.StatusOK, "application/rss+xml", podcast.Bytes())
	})

	r.GET("/api/metadata/:hashId", func(c *gin.Context) {
		hashId := c.Param("hashId")
		if hashId == "" || len(hashId) > 12 {
			c.JSON(badRequest(errors.New("invalid feed id")))
			return
		}

		feed, err := feed.GetMetadata(hashId)
		if err != nil {
			c.JSON(internalError(err))
			return
		}

		c.JSON(http.StatusOK, feed)
	})

	return r
}

func badRequest(err error) (int, interface{}) {
	return http.StatusBadRequest, gin.H{"error": err.Error()}
}

func internalError(err error) (int, interface{}) {
	return http.StatusInternalServerError, gin.H{"error": err.Error()}
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
