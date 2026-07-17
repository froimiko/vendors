package emby

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserPlaybackInfoAdvertisesExternalTextSubtitleProfiles(t *testing.T) {
	requestChecked := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Error("PlaybackInfo request did not use POST")
		}
		if r.URL.Path != "/emby/Items/playback-item/PlaybackInfo" {
			t.Error("PlaybackInfo request used an unexpected endpoint")
		}

		var payload struct {
			DeviceProfile struct {
				SubtitleProfiles []SubtitleProfile `json:"SubtitleProfiles"`
			} `json:"DeviceProfile"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode PlaybackInfo request: %v", err)
		}

		profiles := make(map[string]string, len(payload.DeviceProfile.SubtitleProfiles))
		for _, profile := range payload.DeviceProfile.SubtitleProfiles {
			profiles[profile.Format] = profile.Method
		}
		for _, format := range []string{"vtt", "ass", "ssa", "srt"} {
			if profiles[format] != "External" {
				t.Errorf("missing External subtitle profile for format %q", format)
			}
		}
		requestChecked <- struct{}{}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"PlaySessionId":"session","MediaSources":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, WithUserID("playback-user"), WithKey("playback-token"))
	if _, err := client.UserPlaybackInfo("playback-item"); err != nil {
		t.Fatalf("UserPlaybackInfo failed: %v", err)
	}
	select {
	case <-requestChecked:
	default:
		t.Fatal("PlaybackInfo request was not inspected")
	}
}
