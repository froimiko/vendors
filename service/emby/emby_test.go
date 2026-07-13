package emby

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/synctv-org/vendors/api/emby"
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

func TestGetItemProvesVirtualRootReachability(t *testing.T) {
	const (
		userID = "user-1"
		rootID = "root-1"
		itemID = "item-1"
		token  = "token-1"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/emby/Items"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		query := r.URL.Query()
		for key, want := range map[string]string{
			"UserId": userID, "ParentId": rootID, "Recursive": "true", "Ids": itemID, "Limit": "2",
		} {
			if got := query.Get(key); got != want {
				t.Errorf("query %s = %q, want %q", key, got, want)
			}
		}
		if got := r.Header.Get("X-Emby-Token"); got != token {
			t.Errorf("X-Emby-Token = %q, want %q", got, token)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"Items\":[{\"Id\":\"item-1\",\"Name\":\"Episode\",\"ParentId\":\"physical-parent\"}],\"TotalRecordCount\":1}"))
	}))
	defer server.Close()

	item, err := (&EmbyService{}).GetItem(context.Background(), &pb.GetItemReq{
		Host: server.URL, Token: token, ItemId: itemID, UserId: userID, RootItemId: rootID,
	})
	if err != nil {
		t.Fatalf("GetItem() error = %v", err)
	}
	if got := item.GetId(); got != itemID {
		t.Fatalf("Id = %q, want %q", got, itemID)
	}
	if got := item.GetParentId(); got != rootID {
		t.Fatalf("ParentId proof = %q, want %q", got, rootID)
	}
}

func TestGetItemUsesDocumentedItemsQueryForRootReachability(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/emby/Items" {
			t.Fatalf("path = %q, want documented /emby/Items endpoint", r.URL.Path)
		}
		if r.URL.Query().Get("UserId") == "" {
			t.Fatal("UserId query is missing")
		}
		_, _ = w.Write([]byte("{\"Items\":[{\"Id\":\"item-1\"}],\"TotalRecordCount\":1}"))
	}))
	defer server.Close()

	_, err := (&EmbyService{}).GetItem(context.Background(), &pb.GetItemReq{
		Host: server.URL, Token: "token-1", ItemId: "item-1", UserId: "user-1", RootItemId: "root-1",
	})
	if err != nil {
		t.Fatalf("GetItem() error = %v", err)
	}
}

func TestGetItemRejectsAmbiguousVirtualRootResults(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "zero items", body: "{\"Items\":[],\"TotalRecordCount\":0}"},
		{name: "nil item", body: "{\"Items\":[null],\"TotalRecordCount\":1}"},
		{name: "wrong item ID", body: "{\"Items\":[{\"Id\":\"outside\"}],\"TotalRecordCount\":1}"},
		{name: "multiple items", body: "{\"Items\":[{\"Id\":\"item-1\"},{\"Id\":\"item-2\"}],\"TotalRecordCount\":2}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			item, err := (&EmbyService{}).GetItem(context.Background(), &pb.GetItemReq{
				Host: server.URL, Token: "token-1", ItemId: "item-1", UserId: "user-1", RootItemId: "root-1",
			})
			if err == nil {
				t.Fatal("GetItem() error = nil, want rejection")
			}
			if item != nil {
				t.Fatalf("GetItem() item = %#v, want nil", item)
			}
		})
	}
}
