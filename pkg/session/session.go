package session

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/mxpv/podsync/pkg/api"
)

const (
	identitySessionKey = "identity"
	stateKey           = "state"
)

var errBrokenSession = errors.New("broken session, try to login again")

func Clear(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	s.Save()
}

func GetIdentity(c *gin.Context) (*api.Identity, error) {
	s := sessions.Default(c)
	i := &api.Identity{}

	buf, ok := s.Get(identitySessionKey).(string)
	if ok {
		// Deserialize string to Identity{}
		if err := json.Unmarshal([]byte(buf), i); err != nil {
			s.Clear()
			s.Save()

			return nil, errBrokenSession
		}
	}

	return i, nil
}

func SetIdentity(c *gin.Context, identity *api.Identity) error {
	buf, err := json.Marshal(identity)
	if err != nil {
		return err
	}

	s := sessions.Default(c)
	s.Clear()
	s.Set(identitySessionKey, string(buf))
	return s.Save()
}

func SetState(c *gin.Context) (string, error) {
	s := sessions.Default(c)
	state := randToken()
	s.Set(stateKey, state)
	return state, s.Save()
}

func GetSetate(c *gin.Context) interface{} {
	s := sessions.Default(c)
	return s.Get(stateKey)
}

func randToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
