package database

type Quality string
type Format string

const (
	HighQuality = Quality("high")
	LowQuality  = Quality("low")
	AudioFormat = Format("audio")
	VideoFormat = Format("video")
)

type Feed struct {
	Id       int64
	HashId   string
	UserId   string
	URL      string
	PageSize int
	Quality  Quality
	Format   Format
}

// Query helpers

type WhereFunc func() (string, interface{})

func WithId(id int) WhereFunc {
	return func() (string, interface{}) {
		return "id", id
	}
}

func WithHashId(hashId string) WhereFunc {
	return func() (string, interface{}) {
		return "hash_id", hashId
	}
}

func WithUserId(userId string) WhereFunc {
	return func() (string, interface{}) {
		return "user_id", userId
	}
}
