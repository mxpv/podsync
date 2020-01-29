package ytdl

import (
	"reflect"
	"testing"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

func TestNewOptionsAudio(t *testing.T) {
	type args struct {
		feedConfig *config.Feed
	}
	tests := []struct {
		name string
		args args
		want *OptionsAudio
	}{
		{
			"Get OptionsAudio in low quality",
			args{
				feedConfig: &config.Feed{Quality: model.QualityLow},
			},
			&OptionsAudio{quality: model.QualityLow},
		},
		{
			"Get OptionsAudio in high quality",
			args{
				feedConfig: &config.Feed{Quality: model.QualityHigh},
			},
			&OptionsAudio{quality: model.QualityHigh},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOptionsAudio(tt.args.feedConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOptionsAudio() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptionsAudio_GetConfig(t *testing.T) {
	type fields struct {
		quality model.Quality
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{"OptionsAudio in unknown quality", fields{quality: model.Quality("unknown")}, []string{"--extract-audio", "--audio-format", "mp3", "--format", "bestaudio"}},
		{"OptionsAudio in low quality", fields{quality: model.Quality("low")}, []string{"--extract-audio", "--audio-format", "mp3", "--format", "worstaudio"}},
		{"OptionsAudio in high quality", fields{quality: model.Quality("high")}, []string{"--extract-audio", "--audio-format", "mp3", "--format", "bestaudio"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := OptionsAudio{
				quality: tt.fields.quality,
			}
			if got := options.GetConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
