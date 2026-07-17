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
		if r.Method != http.MethodPost || r.URL.Path != "/emby/Items/playback-item/PlaybackInfo" {
			t.Error("PlaybackInfo request used an unexpected route")
		}
		var payload struct {
			DeviceProfile struct {
				SubtitleProfiles []SubtitleProfile `json:"SubtitleProfiles"`
			} `json:"DeviceProfile"`
			MediaSourceID       string `json:"MediaSourceId"`
			SubtitleStreamIndex *int   `json:"SubtitleStreamIndex"`
			IsPlayback          bool   `json:"IsPlayback"`
			AutoOpenLiveStream  bool   `json:"AutoOpenLiveStream"`
			EnableDirectPlay    bool   `json:"EnableDirectPlay"`
			EnableDirectStream  bool   `json:"EnableDirectStream"`
			EnableTranscoding   bool   `json:"EnableTranscoding"`
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
		if payload.MediaSourceID != "source-real" || payload.SubtitleStreamIndex == nil || *payload.SubtitleStreamIndex != 0 {
			t.Error("selected source/subtitle were not serialized with zero presence")
		}
		if !payload.EnableDirectPlay || !payload.EnableDirectStream {
			t.Error("direct delivery capabilities were not explicitly enabled")
		}
		if payload.IsPlayback || payload.AutoOpenLiveStream || payload.EnableTranscoding {
			t.Error("non-playback request opened playback or transcoding semantics")
		}
		requestChecked <- struct{}{}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"PlaySessionId":"session","MediaSources":[]}`))
	}))
	defer server.Close()

	client := NewClient(server.URL, WithUserID("playback-user"), WithKey("playback-token"))
	if _, err := client.UserPlaybackInfo("playback-item", WithMediaSourceID("source-real"), WithSubtitleStreamIndex(0), WithNonPlayback()); err != nil {
		t.Fatalf("UserPlaybackInfo failed: %v", err)
	}
	select {
	case <-requestChecked:
	default:
		t.Fatal("PlaybackInfo request was not inspected")
	}
}
