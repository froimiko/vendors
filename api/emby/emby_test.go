package emby

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestMediaIdentityProofDescriptors(t *testing.T) {
	tests := []struct {
		name       string
		message    protoreflect.Message
		legacyLen  int
		presentNum protoreflect.FieldNumber
		matchesNum protoreflect.FieldNumber
	}{
		{"media stream", (&MediaStreamInfo{}).ProtoReflect(), 16, 17, 18},
		{"media source", (&MediaSourceInfo{}).ProtoReflect(), 10, 11, 12},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := tt.message.Descriptor().Fields()
			if got, want := fields.Len(), tt.legacyLen+2; got != want {
				t.Fatalf("safe identity field count = %d, want %d", got, want)
			}
			for number := protoreflect.FieldNumber(1); number <= protoreflect.FieldNumber(tt.legacyLen); number++ {
				if field := fields.ByNumber(number); field == nil {
					t.Fatalf("legacy field %d not found", number)
				}
			}
			if fields.ByName("itemId") != nil {
				t.Fatal("raw item ID must not cross the protobuf boundary")
			}
			for _, expected := range []struct {
				name     protoreflect.Name
				jsonName string
				number   protoreflect.FieldNumber
			}{
				{"itemIdPresent", "itemIdPresent", tt.presentNum},
				{"itemIdMatchesRequested", "itemIdMatchesRequested", tt.matchesNum},
			} {
				field := fields.ByName(expected.name)
				if field == nil {
					t.Fatalf("safe identity field %q not found", expected.name)
				}
				if field.JSONName() != expected.jsonName || field.Number() != expected.number || field.Kind() != protoreflect.BoolKind {
					t.Errorf("safe identity field %q has an unexpected descriptor", expected.name)
				}
			}
		})
	}
}

func TestMediaStreamInfoRoundTrip(t *testing.T) {
	want := &MediaStreamInfo{
		Codec:                  "srt",
		Language:               "eng",
		Type:                   "Subtitle",
		Title:                  "English",
		DisplayTitle:           "English SRT",
		DisplayLanguage:        "English",
		IsDefault:              true,
		Index:                  7,
		Protocol:               "File",
		DeliveryUrl:            "/Videos/item/Subtitles/7/Stream.srt",
		DeliveryMethod:         "External",
		IsTextSubtitleStream:   true,
		IsExternal:             true,
		SupportsExternalStream: true,
		SubtitleLocationType:   "External",
		MimeType:               "application/x-subrip",
	}

	wire, err := proto.Marshal(want)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	var fromWire MediaStreamInfo
	if err := proto.Unmarshal(wire, &fromWire); err != nil {
		t.Fatalf("proto.Unmarshal() error = %v", err)
	}
	if !proto.Equal(want, &fromWire) {
		t.Fatal("binary round trip changed media stream data")
	}

	jsonData, err := protojson.Marshal(want)
	if err != nil {
		t.Fatalf("protojson.Marshal() error = %v", err)
	}
	var fromJSON MediaStreamInfo
	if err := protojson.Unmarshal(jsonData, &fromJSON); err != nil {
		t.Fatalf("protojson.Unmarshal() error = %v", err)
	}
	if !proto.Equal(want, &fromJSON) {
		t.Fatal("JSON round trip changed media stream data")
	}
}

func TestMediaIdentityProofRoundTrip(t *testing.T) {
	want := &MediaSourceInfo{MediaStreamInfo: []*MediaStreamInfo{{}}}
	setProof := func(message protoreflect.Message, name protoreflect.Name, value bool) {
		t.Helper()
		field := message.Descriptor().Fields().ByName(name)
		if field == nil {
			t.Fatalf("safe identity field %q not found", name)
		}
		message.Set(field, protoreflect.ValueOfBool(value))
	}
	setProof(want.ProtoReflect(), "itemIdPresent", true)
	setProof(want.ProtoReflect(), "itemIdMatchesRequested", true)
	setProof(want.MediaStreamInfo[0].ProtoReflect(), "itemIdPresent", true)
	setProof(want.MediaStreamInfo[0].ProtoReflect(), "itemIdMatchesRequested", false)

	assertProofs := func(message *MediaSourceInfo) {
		t.Helper()
		for _, expected := range []struct {
			message protoreflect.Message
			name    protoreflect.Name
			value   bool
		}{
			{message.ProtoReflect(), "itemIdPresent", true},
			{message.ProtoReflect(), "itemIdMatchesRequested", true},
			{message.MediaStreamInfo[0].ProtoReflect(), "itemIdPresent", true},
			{message.MediaStreamInfo[0].ProtoReflect(), "itemIdMatchesRequested", false},
		} {
			field := expected.message.Descriptor().Fields().ByName(expected.name)
			if field == nil || expected.message.Get(field).Bool() != expected.value {
				t.Errorf("safe identity proof %q did not survive round trip", expected.name)
			}
		}
	}

	wire, err := proto.Marshal(want)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}
	var fromWire MediaSourceInfo
	if err := proto.Unmarshal(wire, &fromWire); err != nil {
		t.Fatalf("proto.Unmarshal() error = %v", err)
	}
	assertProofs(&fromWire)

	jsonData, err := protojson.Marshal(want)
	if err != nil {
		t.Fatalf("protojson.Marshal() error = %v", err)
	}
	var fromJSON MediaSourceInfo
	if err := protojson.Unmarshal(jsonData, &fromJSON); err != nil {
		t.Fatalf("protojson.Unmarshal() error = %v", err)
	}
	assertProofs(&fromJSON)
}
