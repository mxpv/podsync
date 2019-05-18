package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/mxpv/patreon-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/session"
)

const (
	maxHashIDLength = 16
)

type feedService interface {
	CreateFeed(req *api.CreateFeedRequest, identity *api.Identity) (string, error)
	BuildFeed(hashID string) ([]byte, error)
	GetMetadata(hashID string) (*api.Metadata, error)
	Downgrade(patronID string, featureLevel int) error
}

type patreonService interface {
	Hook(pledge *patreon.Pledge, event string) error
	GetFeatureLevelByID(patronID string) int
	GetFeatureLevelFromAmount(amount int) int
}

type Opts struct {
	CookieSecret          string
	PatreonClientID       string
	PatreonSecret         string
	PatreonRedirectURL    string
	PatreonWebhooksSecret string
}

type handler struct {
	feed                  feedService
	oauth2                oauth2.Config
	patreon               patreonService
	PatreonWebhooksSecret string
}

func New(feed feedService, support patreonService, opts Opts) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	cacheStore := persistence.NewRedisCache("redis:6379", "", time.Second)

	store := sessions.NewCookieStore([]byte(opts.CookieSecret))
	r.Use(sessions.Sessions("podsync", store))

	h := handler{
		feed:                  feed,
		patreon:               support,
		PatreonWebhooksSecret: opts.PatreonWebhooksSecret,
	}

	// OAuth 2 configuration

	h.oauth2 = oauth2.Config{
		ClientID:     opts.PatreonClientID,
		ClientSecret: opts.PatreonSecret,
		RedirectURL:  opts.PatreonRedirectURL,
		Scopes:       []string{"users", "pledges-to-me", "my-campaign"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  patreon.AuthorizationURL,
			TokenURL: patreon.AccessTokenURL,
		},
	}

	// Handlers

	r.GET("/user/login", h.login)
	r.GET("/user/logout", h.logout)
	r.GET("/user/patreon", h.patreonCallback)

	r.GET("/api/ping", h.ping)
	r.GET("/api/user", h.user)
	r.POST("/api/create", h.create)
	r.GET("/api/metadata/:hashId", h.metadata)
	r.POST("/api/webhooks", h.webhook)

	const feedTTL = 30 * time.Minute
	r.NoRoute(cache.CachePage(cacheStore, feedTTL, h.getFeed))

	return r
}

func (h handler) login(c *gin.Context) {
	state, err := session.SetState(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	authURL := h.oauth2.AuthCodeURL(state)
	c.Redirect(http.StatusFound, authURL)
}

func (h handler) logout(c *gin.Context) {
	session.Clear(c)

	c.Redirect(http.StatusFound, "/")
}

func (h handler) patreonCallback(c *gin.Context) {
	// Validate session state
	if session.GetSetate(c) != c.Query("state") {
		c.String(http.StatusUnauthorized, "invalid state")
		return
	}

	// Exchange code with tokens
	token, err := h.oauth2.Exchange(c.Request.Context(), c.Query("code"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Create Patreon client
	tc := h.oauth2.Client(c.Request.Context(), token)
	client := patreon.NewClient(tc)

	// Query user info from Patreon
	user, err := client.FetchUser()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// Determine feature level
	level := h.patreon.GetFeatureLevelByID(user.Data.ID)

	identity := &api.Identity{
		UserID:       user.Data.ID,
		FullName:     user.Data.Attributes.FullName,
		Email:        user.Data.Attributes.Email,
		ProfileURL:   user.Data.Attributes.URL,
		FeatureLevel: level,
	}

	session.SetIdentity(c, identity)
	c.Redirect(http.StatusFound, "/")
}

func (h handler) ping(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h handler) user(c *gin.Context) {
	identity, err := session.GetIdentity(c)
	if err != nil {
		identity = &api.Identity{}
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":       identity.UserID,
		"feature_level": identity.FeatureLevel,
		"full_name":     identity.FullName,
	})
}

func (h handler) create(c *gin.Context) {
	req := &api.CreateFeedRequest{}

	if err := c.BindJSON(req); err != nil {
		c.JSON(badRequest(err))
		return
	}

	identity, err := session.GetIdentity(c)
	if err != nil {
		c.JSON(internalError(err))
		return
	}

	// Check feature level again if user deleted pledge by still logged in
	identity.FeatureLevel = h.patreon.GetFeatureLevelByID(identity.UserID)

	hashID, err := h.feed.CreateFeed(req, identity)
	if err != nil {
		c.JSON(internalError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": hashID})
}

func (h handler) getFeed(c *gin.Context) {
	hashID := c.Request.URL.Path[1:]
	if hashID == "" || len(hashID) > maxHashIDLength {
		c.String(http.StatusBadRequest, "invalid feed id")
		return
	}

	if strings.HasSuffix(hashID, ".xml") {
		hashID = strings.TrimSuffix(hashID, ".xml")
	}

	podcast, err := h.feed.BuildFeed(hashID)
	if err != nil {
		code := http.StatusInternalServerError
		if err == api.ErrNotFound {
			code = http.StatusNotFound
		} else if err == api.ErrQuotaExceeded {
			code = http.StatusTooManyRequests
		}

		c.String(code, err.Error())
		return
	}

	const feedContentType = "application/rss+xml; charset=UTF-8"
	c.Data(http.StatusOK, feedContentType, podcast)
}

func (h handler) metadata(c *gin.Context) {
	hashID := c.Param("hashId")
	if hashID == "" || len(hashID) > maxHashIDLength {
		c.String(http.StatusBadRequest, "invalid feed id")
		return
	}

	feed, err := h.feed.GetMetadata(hashID)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, feed)
}

func (h handler) webhook(c *gin.Context) {
	// Read body to byte array in order to verify signature first
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.WithError(err).Error("failed to read webhook request")
		c.Status(http.StatusBadRequest)
		return
	}

	// Verify signature
	signature := c.GetHeader(patreon.HeaderSignature)
	valid, err := patreon.VerifySignature(body, h.PatreonWebhooksSecret, signature)
	if err != nil {
		log.WithError(err).Error("failed to verify signature")
		c.Status(http.StatusBadRequest)
		return
	}

	if !valid {
		log.Errorf("webhooks signatures are not equal (header: %s)", signature)
		c.Status(http.StatusUnauthorized)
		return
	}

	// Get event name
	eventName := c.GetHeader(patreon.HeaderEventType)
	if eventName == "" {
		log.Error("event name header is empty")
		c.Status(http.StatusBadRequest)
		return
	}

	pledge := &patreon.WebhookPledge{}
	if err := json.Unmarshal(body, pledge); err != nil {
		log.WithError(err).Error("failed to unmarshal pledge")
		c.JSON(badRequest(err))
		return
	}

	if err := h.patreon.Hook(&pledge.Data, eventName); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user_id":      pledge.Data.Relationships.Patron.Data.ID,
			"pledge_id":    pledge.Data.ID,
			"pledge_event": eventName,
		}).Error("failed to process patreon event")

		// Don't return any errors to Patreon, otherwise subsequent notifications will be blocked.
		return
	}

	patronID := pledge.Data.Relationships.Patron.Data.ID

	if eventName == patreon.EventUpdatePledge {
		newLevel := h.patreon.GetFeatureLevelFromAmount(pledge.Data.Attributes.AmountCents)
		if err := h.feed.Downgrade(patronID, newLevel); err != nil {
			return
		}
	} else if eventName == patreon.EventDeletePledge {
		if err := h.feed.Downgrade(patronID, api.DefaultFeatures); err != nil {
			return
		}
	}

	log.Infof("sucessfully processed patreon event %s (%s)", pledge.Data.ID, eventName)
}

func badRequest(err error) (int, interface{}) {
	return http.StatusBadRequest, gin.H{"error": err.Error()}
}

func internalError(err error) (int, interface{}) {
	log.Printf("server error: %v", err)
	return http.StatusInternalServerError, gin.H{"error": err.Error()}
}
