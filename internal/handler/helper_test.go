package handler

import (
	"github.com/alexboor/lbx-telebot/internal/model"
	"reflect"
	"testing"
	"time"
)

func Test_findIndexName(t *testing.T) {
	tests := []struct {
		name string
		opts []string
		want int
	}{
		{
			name: "name exists index 0",
			opts: []string{"name"},
			want: 0,
		},
		{
			name: "name exists index 1",
			opts: []string{"1s", "name"},
			want: 1,
		},
		{
			name: "name not exists",
			opts: []string{"1s"},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := findIndexName(tt.opts); got != tt.want {
				t.Errorf("findIndexName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseInt(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want int
	}{
		{
			name: "correct",
			str:  "10",
			want: 10,
		},
		{
			name: "incorrect",
			str:  "saasdasd",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseInt(tt.str); got != tt.want {
				t.Errorf("parseInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseDuration(t *testing.T) {
	tests := []struct {
		name string
		str  string
		want time.Time
	}{
		{
			name: "yesterday",
			str:  "1d",
			want: time.Now().Add(-1 * time.Hour * 24).Truncate(time.Hour * 24),
		},
		{
			name: "7d",
			str:  "7d",
			want: time.Now().Add(-7 * time.Hour * 24).Truncate(time.Hour * 24),
		},
		{
			name: "today",
			str:  "1s",
			want: time.Now().Truncate(time.Hour * 24),
		},
		{
			name: "incorrect with d",
			str:  "asdasdas",
			want: time.Time{},
		},
		{
			name: "incorrect s1d",
			str:  "s1d",
			want: time.Time{},
		},
		{
			name: "incorrect 1dd",
			str:  "1dd",
			want: time.Time{},
		},
		{
			name: "incorrect without d",
			str:  "qweretre",
			want: time.Time{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseDuration(tt.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseSingleOpt(t *testing.T) {
	tests := []struct {
		name string
		opt  string
		want model.Option
		ok   bool
	}{
		{
			name: "string",
			opt:  "wrong",
			want: model.Option{},
			ok:   false,
		},
		{
			name: "incorrect duration 1ss",
			opt:  "1ss",
			want: model.Option{},
			ok:   false,
		},
		{
			name: "incorrect duration 1dd",
			opt:  "1dd",
			want: model.Option{},
			ok:   false,
		},
		{
			name: "correct int",
			opt:  "15",
			want: model.Option{Limit: 15},
			ok:   true,
		},
		{
			name: "correct duration",
			opt:  "1s",
			want: model.Option{Date: time.Now().Truncate(time.Hour * 24)},
			ok:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseSingleOpt(tt.opt)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSingleOpt() got = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("parseSingleOpt() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func Test_parseTwoOpts(t *testing.T) {
	tests := []struct {
		name string
		opts []string
		want model.Option
		ok   bool
	}{
		{
			name: "incorrect limit",
			opts: []string{"1s", "tess"},
			want: model.Option{},
			ok:   false,
		},
		{
			name: "incorrect duration 1st",
			opts: []string{"1ss", "12"},
			want: model.Option{},
			ok:   false,
		},
		{
			name: "incorrect duration 2nd",
			opts: []string{"12", "1ss"},
			want: model.Option{},
			ok:   false,
		},
		{
			name: "correct limit 1st",
			opts: []string{"12", "1s"},
			want: model.Option{Limit: 12, Date: time.Now().Truncate(time.Hour * 24)},
			ok:   true,
		},
		{
			name: "correct limit 2nd",
			opts: []string{"1s", "12"},
			want: model.Option{Limit: 12, Date: time.Now().Truncate(time.Hour * 24)},
			ok:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseTwoOpts(tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTwoOpts() got = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("parseTwoOpts() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func Test_parseTopAndBottomPayload(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    model.Option
		ok      bool
	}{
		{
			name:    "1 duration opt",
			payload: "1s",
			want:    model.Option{Date: time.Now().Truncate(time.Hour * 24)},
			ok:      true,
		},
		{
			name:    "1 limit opt",
			payload: "3",
			want:    model.Option{Limit: 3},
			ok:      true,
		},
		{
			name:    "2 opts",
			payload: "1s 12",
			want:    model.Option{Date: time.Now().Truncate(time.Hour * 24), Limit: 12},
			ok:      true,
		},
		{
			name:    "2 opts duration",
			payload: "1s 1s",
			want:    model.Option{},
			ok:      false,
		},
		{
			name:    "2 opts limit",
			payload: "1 2",
			want:    model.Option{},
			ok:      false,
		},
		{
			name:    "0 opt",
			payload: "",
			want:    model.Option{},
			ok:      false,
		},
		{
			name:    "3 opts",
			payload: "1s 12 profile",
			want:    model.Option{},
			ok:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseTopAndBottomPayload(tt.payload)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTopAndBottomPayload() got = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("parseTopAndBottomPayload() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}

func Test_parseProfilePayload(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    model.Option
		ok      bool
	}{
		{
			name:    "1 profile opt",
			payload: "profile",
			want:    model.Option{Profile: "profile"},
			ok:      true,
		},
		{
			name:    "1 duration opt",
			payload: "1s",
			want:    model.Option{Date: time.Now().Truncate(time.Hour * 24)},
			ok:      true,
		},
		{
			name:    "2 opts profile 1st",
			payload: "profile 1s",
			want:    model.Option{Date: time.Now().Truncate(time.Hour * 24), Profile: "profile"},
			ok:      true,
		},
		{
			name:    "2 opts profile 2nd",
			payload: "1s profile",
			want:    model.Option{Date: time.Now().Truncate(time.Hour * 24), Profile: "profile"},
			ok:      true,
		},
		{
			name:    "2 opts duration incorrect",
			payload: "profile 1ss",
			want:    model.Option{},
			ok:      false,
		},
		{
			name:    "2 opts profile incorrect",
			payload: "1s 1s",
			want:    model.Option{},
			ok:      false,
		},
		{
			name:    "2 opts profile incorrect",
			payload: "1s 10",
			want:    model.Option{},
			ok:      false,
		},
		{
			name:    "0 opt",
			payload: "",
			want:    model.Option{},
			ok:      false,
		},
		{
			name:    "3 opts",
			payload: "1s profile 12",
			want:    model.Option{},
			ok:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseProfilePayload(tt.payload)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseProfilePayload() got = %v, want %v", got, tt.want)
			}
			if ok != tt.ok {
				t.Errorf("parseProfilePayload() ok = %v, want %v", ok, tt.ok)
			}
		})
	}
}
