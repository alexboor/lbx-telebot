package reaction

import (
	"reflect"
	"sort"
	"testing"

	tele "gopkg.in/telebot.v3"
)

func TestNewlyAdded(t *testing.T) {
	tests := []struct {
		name string
		old  []tele.Reaction
		new  []tele.Reaction
		want []string
	}{
		{
			name: "no reactions",
			old:  nil,
			new:  nil,
			want: nil,
		},
		{
			name: "first reaction added",
			old:  nil,
			new:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			want: []string{"👍"},
		},
		{
			name: "existing reaction is not newly added",
			old:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			new:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			want: nil,
		},
		{
			name: "second emoji added alongside an existing one",
			old:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			new:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}, {Type: "emoji", Emoji: "🔥"}},
			want: []string{"🔥"},
		},
		{
			name: "reaction removed, nothing newly added",
			old:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			new:  nil,
			want: nil,
		},
		{
			name: "custom emoji reactions are ignored",
			old:  nil,
			new:  []tele.Reaction{{Type: "custom_emoji", CustomEmoji: "12345"}},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newlyAdded(tt.old, tt.new)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newlyAdded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasNewPositive(t *testing.T) {
	tests := []struct {
		name string
		old  []tele.Reaction
		new  []tele.Reaction
		want bool
	}{
		{
			name: "new positive emoji",
			old:  nil,
			new:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			want: true,
		},
		{
			name: "new non-positive emoji",
			old:  nil,
			new:  []tele.Reaction{{Type: "emoji", Emoji: "💩"}},
			want: false,
		},
		{
			name: "positive emoji already present is not re-triggered",
			old:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			new:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			want: false,
		},
		{
			name: "multiple new emoji, one positive",
			old:  nil,
			new:  []tele.Reaction{{Type: "emoji", Emoji: "💩"}, {Type: "emoji", Emoji: "🔥"}},
			want: true,
		},
		{
			name: "reaction removed",
			old:  []tele.Reaction{{Type: "emoji", Emoji: "👍"}},
			new:  nil,
			want: false,
		},
		{
			name: "new custom emoji reaction is ignored, even if it were positive",
			old:  nil,
			new:  []tele.Reaction{{Type: "custom_emoji", CustomEmoji: "12345"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasNewPositive(tt.old, tt.new); got != tt.want {
				t.Errorf("hasNewPositive() = %v, want %v", got, tt.want)
			}
		})
	}
}
