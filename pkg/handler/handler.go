package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	patreon "github.com/mxpv/patreon-go"
	itunes "github.com/mxpv/podcast"
	"golang.org/x/oauth2"

	"github.com/mxpv/podsync/pkg/api"
	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/session"

	log "github.com/sirupsen/logrus"
)

const (
	maxHashIDLength = 16
)

type feedService interface {
	CreateFeed(req *api.CreateFeedRequest, identity *api.Identity) (string, error)
	BuildFeed(hashID string) (*itunes.Podcast, error)
	GetMetadata(hashID string) (*api.Metadata, error)
	Downgrade(patronID string, featureLevel int) error
}

type patreonService interface {
	Hook(pledge *patreon.Pledge, event string) error
	GetFeatureLevelByID(patronID string) int
	GetFeatureLevelFromAmount(amount int) int
}

type cacheService interface {
	Set(key, value string, ttl time.Duration) error
	Get(key string) (string, error)
}

type handler struct {
	feed    feedService
	cfg     *config.AppConfig
	oauth2  oauth2.Config
	patreon patreonService
	cache   cacheService
}

func New(feed feedService, support patreonService, cache cacheService, cfg *config.AppConfig) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	store := sessions.NewCookieStore([]byte(cfg.CookieSecret))
	r.Use(sessions.Sessions("podsync", store))

	h := handler{
		feed:    feed,
		patreon: support,
		cache:   cache,
		cfg:     cfg,
	}

	// OAuth 2 configuration

	h.oauth2 = oauth2.Config{
		ClientID:     cfg.PatreonClientID,
		ClientSecret: cfg.PatreonSecret,
		RedirectURL:  cfg.PatreonRedirectURL,
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

	r.NoRoute(h.getFeed)

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

	createEventLog := log.WithField("event", "create_feed")

	if err := c.BindJSON(req); err != nil {
		createEventLog.WithError(err).Error("invalid request")
		c.JSON(badRequest(err))
		return
	}

	identity, err := session.GetIdentity(c)
	if err != nil {
		createEventLog.WithError(err).Error("invalid identity")
		c.JSON(internalError(err))
		return
	}

	// Check feature level again if user deleted pledge by still logged in
	identity.FeatureLevel = h.patreon.GetFeatureLevelByID(identity.UserID)

	createEventLog = createEventLog.WithFields(log.Fields{
		"user_id":       identity.UserID,
		"feature_level": identity.FeatureLevel,
	})

	createEventLog.Info("creating feed")

	hashID, err := h.feed.CreateFeed(req, identity)
	if err != nil {
		createEventLog.WithError(err).Error("failed to create new feed")
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

	log.WithFields(log.Fields{
		"event":   "get_feed",
		"hash_id": hashID,
	}).Infof("getting feed %s", hashID)

	if strings.HasSuffix(hashID, ".xml") {
		hashID = strings.TrimSuffix(hashID, ".xml")
	}

	const feedContentType = "application/rss+xml; charset=UTF-8"

	cached, err := h.cache.Get(hashID)
	if err == nil {
		c.Data(http.StatusOK, feedContentType, []byte(cached))
		return
	}

	podcast, err := h.feed.BuildFeed(hashID)
	if err != nil {
		code := http.StatusInternalServerError
		if err == api.ErrNotFound {
			code = http.StatusNotFound
		} else if err == api.ErrQuotaExceeded {
			code = http.StatusTooManyRequests
		}

		log.WithFields(log.Fields{
			"event":     "get_feed",
			"hash_id":   hashID,
			"http_code": code,
		}).WithError(err).Error("failed to get feed")
		c.String(code, err.Error())
		return
	}

	data := podcast.String()

	if err := h.cache.Set(hashID, data, 10*time.Minute); err != nil {
		log.WithError(err).Warnf("failed to cache feed %q", hashID)
	}

	c.Data(http.StatusOK, feedContentType, []byte(data))
}

func (h handler) metadata(c *gin.Context) {
	hashID := c.Param("hashId")
	if hashID == "" || len(hashID) > maxHashIDLength {
		c.String(http.StatusBadRequest, "invalid feed id")
		return
	}

	log.WithFields(log.Fields{
		"event":   "get_metadata",
		"hash_id": hashID,
	}).Infof("getting metadata for '%s'", hashID)
	feed, err := h.feed.GetMetadata(hashID)
	if err != nil {
		log.WithFields(log.Fields{
			"event":   "get_metadata",
			"hash_id": hashID,
		}).WithError(err).Error("failed to query metadata")
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
	valid, err := patreon.VerifySignature(body, h.cfg.PatreonWebhooksSecret, signature)
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
