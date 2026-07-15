package emby

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	pb "github.com/synctv-org/vendors/api/emby"
	vendoremby "github.com/synctv-org/vendors/vendors/emby"
)

func TestGetItemUsesUserItemEndpoint(t *testing.T) {
	const (
		userID   = "user-1"
		itemID   = "item-1"
		token    = "token-1"
		parentID = "parent-from-user-view"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/emby/Users/"+userID+"/Items/"+itemID; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got := r.Header.Get("X-Emby-Token"); got != token {
			t.Errorf("X-Emby-Token = %q, want %q", got, token)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"Id\":\"item-1\",\"Name\":\"Episode\",\"ParentId\":\"parent-from-user-view\"}"))
	}))
	defer server.Close()

	item, err := (&EmbyService{}).GetItem(context.Background(), &pb.GetItemReq{
		Host: server.URL, Token: token, ItemId: itemID, UserId: userID,
	})
	if err != nil {
		t.Fatalf("GetItem() error = %v", err)
	}
	if got := item.GetParentId(); got != parentID {
		t.Fatalf("ParentId = %q, want %q", got, parentID)
	}
}

func TestGetItemIgnoresLegacyRootItemIDAndUsesUserItemEndpoint(t *testing.T) {
	const (
		userID   = "user-1"
		itemID   = "item-1"
		token    = "token-1"
		parentID = "physical-parent"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/emby/Users/"+userID+"/Items/"+itemID; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if r.URL.RawQuery != "" {
			t.Errorf("legacy root proof query is still active: %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"Id\":\"item-1\",\"ParentId\":\"physical-parent\"}"))
	}))
	defer server.Close()

	item, err := (&EmbyService{}).GetItem(context.Background(), &pb.GetItemReq{
		Host: server.URL, Token: token, ItemId: itemID, UserId: userID, RootItemId: "legacy-root",
	})
	if err != nil {
		t.Fatalf("GetItem() error = %v", err)
	}
	if item.GetParentId() != parentID {
		t.Fatalf("ParentId = %q, want physical value %q", item.GetParentId(), parentID)
	}
}

func TestMediaStreamInfo2PB(t *testing.T) {
	tests := []struct {
		name string
		in   vendoremby.MediaStreams
		want *pb.MediaStreamInfo
	}{
		{
			name: "complete stream",
			in: vendoremby.MediaStreams{
				Codec:                  "ass",
				Language:               "eng",
				Type:                   "Subtitle",
				Title:                  "English",
				DisplayTitle:           "English ASS",
				DisplayLanguage:        "English",
				IsDefault:              true,
				Index:                  4,
				Protocol:               "File",
				DeliveryURL:            "/Videos/item/Subtitles/4/Stream.ass?api_key=redacted",
				DeliveryMethod:         "External",
				IsTextSubtitleStream:   true,
				IsExternal:             true,
				SupportsExternalStream: true,
				SubtitleLocationType:   "External",
				MimeType:               "text/x-ssa",
			},
			want: &pb.MediaStreamInfo{
				Codec:                  "ass",
				Language:               "eng",
				Type:                   "Subtitle",
				Title:                  "English",
				DisplayTitle:           "English ASS",
				DisplayLanguage:        "English",
				IsDefault:              true,
				Index:                  4,
				Protocol:               "File",
				DeliveryUrl:            "/Videos/item/Subtitles/4/Stream.ass?api_key=redacted",
				DeliveryMethod:         "External",
				IsTextSubtitleStream:   true,
				IsExternal:             true,
				SupportsExternalStream: true,
				SubtitleLocationType:   "External",
				MimeType:               "text/x-ssa",
			},
		},
		{
			name: "false and empty values",
			in:   vendoremby.MediaStreams{},
			want: &pb.MediaStreamInfo{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mediaStreamInfo2pb([]vendoremby.MediaStreams{tt.in})
			if len(got) != 1 {
				t.Fatalf("mediaStreamInfo2pb() length = %d, want 1", len(got))
			}
			if !reflect.DeepEqual(got[0], tt.want) {
				t.Errorf("mediaStreamInfo2pb() = %#v, want %#v", got[0], tt.want)
			}
		})
	}
}
