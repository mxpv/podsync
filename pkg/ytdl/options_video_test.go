package ytdl

import (
	"reflect"
	"testing"

	"github.com/mxpv/podsync/pkg/config"
	"github.com/mxpv/podsync/pkg/model"
)

func TestNewOptionsVideo(t *testing.T) {
	type args struct {
		feedConfig *config.Feed
	}
	tests := []struct {
		name string
		args args
		want *OptionsVideo
	}{
		{
			"Get OptionsVideo in low quality",
			args{
				feedConfig: &config.Feed{Quality: model.QualityLow},
			},
			&OptionsVideo{quality: model.QualityLow},
		},
		{
			"Get OptionsVideo in low quality with maxheight",
			args{
				feedConfig: &config.Feed{Quality: model.QualityLow, MaxHeight: 720},
			},
			&OptionsVideo{quality: model.QualityLow, maxHeight: 720},
		},
		{
			"Get OptionsVideo in high quality",
			args{
				feedConfig: &config.Feed{Quality: model.QualityHigh},
			},
			&OptionsVideo{quality: model.QualityHigh},
		},
		{
			"Get OptionsVideo in high quality with maxheight",
			args{
				feedConfig: &config.Feed{Quality: model.QualityHigh, MaxHeight: 720},
			},
			&OptionsVideo{quality: model.QualityHigh, maxHeight: 720},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewOptionsVideo(tt.args.feedConfig); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewOptionsVideo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOptionsVideo_GetConfig(t *testing.T) {
	type fields struct {
		quality   model.Quality
		maxHeight int
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			"OptionsVideo in unknown quality",
			fields{quality: model.Quality("unknown")},
			[]string{"--format", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best"},
		},
		{
			"OptionsVideo in unknown quality with maxheight",
			fields{quality: model.Quality("unknown"), maxHeight: 720},
			[]string{"--format", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best"},
		},
		{
			"OptionsVideo in low quality",
			fields{quality: model.QualityLow},
			[]string{"--format", "worstvideo[ext=mp4]+worstaudio[ext=m4a]/worst[ext=mp4]/worst"},
		},
		{
			"OptionsVideo in low quality with maxheight",
			fields{quality: model.QualityLow, maxHeight: 720},
			[]string{"--format", "worstvideo[ext=mp4]+worstaudio[ext=m4a]/worst[ext=mp4]/worst"},
		},
		{
			"OptionsVideo in high quality",
			fields{quality: model.QualityHigh},
			[]string{"--format", "bestvideo[ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best"},
		},
		{
			"OptionsVideo in high quality with maxheight",
			fields{quality: model.QualityHigh, maxHeight: 720},
			[]string{"--format", "bestvideo[height<=720][ext=mp4]+bestaudio[ext=m4a]/best[ext=mp4]/best"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := OptionsVideo{
				quality:   tt.fields.quality,
				maxHeight: tt.fields.maxHeight,
			}
			if got := options.GetConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
