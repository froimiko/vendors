package emby

import (
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestMediaStreamInfoDescriptor(t *testing.T) {
	message := (&MediaStreamInfo{}).ProtoReflect()
	fields := message.Descriptor().Fields()
	if got, want := fields.Len(), 16; got != want {
		t.Fatalf("Fields().Len() = %d, want %d", got, want)
	}
	for number := protoreflect.FieldNumber(1); number <= 9; number++ {
		if field := fields.ByNumber(number); field == nil {
			t.Fatalf("legacy field %d not found", number)
		}
	}

	tests := []struct {
		name     protoreflect.Name
		jsonName string
		number   protoreflect.FieldNumber
		kind     protoreflect.Kind
	}{
		{"deliveryUrl", "deliveryUrl", 10, protoreflect.StringKind},
		{"deliveryMethod", "deliveryMethod", 11, protoreflect.StringKind},
		{"isTextSubtitleStream", "isTextSubtitleStream", 12, protoreflect.BoolKind},
		{"isExternal", "isExternal", 13, protoreflect.BoolKind},
		{"supportsExternalStream", "supportsExternalStream", 14, protoreflect.BoolKind},
		{"subtitleLocationType", "subtitleLocationType", 15, protoreflect.StringKind},
		{"mimeType", "mimeType", 16, protoreflect.StringKind},
	}

	for _, tt := range tests {
		t.Run(string(tt.name), func(t *testing.T) {
			field := fields.ByName(tt.name)
			if field == nil {
				t.Fatalf("field %q not found", tt.name)
			}
			if got := field.JSONName(); got != tt.jsonName {
				t.Errorf("JSONName() = %q, want %q", got, tt.jsonName)
			}
			if got := field.Number(); got != tt.number {
				t.Errorf("Number() = %d, want %d", got, tt.number)
			}
			if got := field.Kind(); got != tt.kind {
				t.Errorf("Kind() = %v, want %v", got, tt.kind)
			}
			if got := fields.ByNumber(tt.number); got != field {
				t.Errorf("Fields().ByNumber(%d) did not return field %q", tt.number, tt.name)
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
		t.Fatalf("binary round trip = %v, want %v", &fromWire, want)
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
		t.Fatalf("JSON round trip = %v, want %v", &fromJSON, want)
	}
}
