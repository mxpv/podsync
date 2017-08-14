package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	itunes "github.com/mxpv/podcast"
	"github.com/mxpv/podsync/web/pkg/api"
	"github.com/pkg/errors"
)

type feed interface {
	CreateFeed(ctx context.Context, req *api.CreateFeedRequest) (string, error)
	GetFeed(hashId string) (*itunes.Podcast, error)
	GetMetadata(hashId string) (*api.Feed, error)
}

func MakeHandlers(feed feed) http.Handler {
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.POST("/create", func(c *gin.Context) {
		req := &api.CreateFeedRequest{}

		if err := c.BindJSON(req); err != nil {
			c.JSON(badRequest(err))
			return
		}

		if err := binding.Validator.ValidateStruct(req); err != nil {
			c.JSON(badRequest(err))
			return
		}

		hashId, err := feed.CreateFeed(c.Request.Context(), req)
		if err != nil {
			c.JSON(internalError(err))
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": hashId})
	})

	r.GET("/:hashId", func(c *gin.Context) {
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

	r.GET("/metadata/:hashId", func(c *gin.Context) {
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
