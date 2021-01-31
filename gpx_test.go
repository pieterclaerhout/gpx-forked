package gpx

import (
	"math"
	"os"
	"testing"
	"time"
)

func TestDecoder(t *testing.T) {
	f, err := os.Open("test/test.gpx")
	if err != nil {
		t.Fatal(err)
	}

	doc, err := NewDecoder(f).Decode()
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != "1.1" {
		t.Errorf("got wrong version %q", doc.Version)
	}
	if dist := doc.Distance(); math.Abs(dist-1362.370020) > 0.0000001 {
		t.Errorf("got %f distance; expected 1362.370020", dist)
	}
	expectedDuration, err := time.ParseDuration("39m19s")
	if err != nil {
		t.Fatal(err)
	}
	if dur := doc.Duration(); dur != expectedDuration {
		t.Errorf("got %s duration; expected %s", dur, expectedDuration)
	}
	expectedStart := time.Date(2015, 12, 13, 18, 35, 18, 0, time.UTC)
	if start := doc.Start(); !start.Equal(expectedStart) {
		t.Errorf("got %v start; expected %v", start, expectedStart)
	}
	expectedEnd := time.Date(2015, 12, 13, 19, 14, 37, 0, time.UTC)
	if end := doc.End(); !end.Equal(expectedEnd) {
		t.Errorf("got %v end; expected %v", end, expectedEnd)
	}

	if l := len(doc.Tracks); l != 1 {
		t.Errorf("got %d track(s); expected 1", l)
	}
	track := doc.Tracks[0]

	if track.Name != "Running" {
		t.Errorf("got %q name; expected %q", track.Name, "Running")
	}
	if track.Type != "running" {
		t.Errorf("got %q type; expected %q", track.Name, "running")
	}
	if l := len(track.Segments); l != 1 {
		t.Errorf("got %d segment(s); expected 1", l)
	}
	seg := track.Segments[0]

	if l := len(seg.Points); l != 9 {
		t.Errorf("got %d points(s); expected 9", l)
	}

	pointTestCases := []struct {
		point Point
		lat   float64
		lon   float64
		ele   float64
		t     time.Time
	}{
		{
			point: seg.Points[0],
			lat:   49.3973693847656250,
			lon:   11.1259574890136719,
			ele:   346.874267578125,
			t:     time.Date(2015, 12, 13, 18, 35, 18, 0, time.UTC),
		},
		{
			point: seg.Points[len(seg.Points)-1],
			lat:   49.3978729248046875,
			lon:   11.1260004043579102,
			ele:   346.11541748046875,
			t:     time.Date(2015, 12, 13, 19, 14, 37, 0, time.UTC),
		},
	}

	for i, testCase := range pointTestCases {
		if testCase.point.Latitude != testCase.lat {
			t.Errorf("point test case %d: got %v latitude; expected %v", i, testCase.point.Latitude, testCase.lat)
		}
		if testCase.point.Longitude != testCase.lon {
			t.Errorf("point test case %d: got %v longitude; expected %v", i, testCase.point.Longitude, testCase.lon)
		}
		if testCase.point.Elevation != testCase.ele {
			t.Errorf("point test case %d: got %v elevation; expected %v", i, testCase.point.Elevation, testCase.ele)
		}
		if !testCase.point.Time.Equal(testCase.t) {
			t.Errorf("point test case %d: got %v time; expected %v", i, testCase.point.Time, testCase.t)
		}
	}

	testMetadata(t, doc.Metadata)
}

func testMetadata(t *testing.T, metadata Metadata) {
	if expected := "Run"; metadata.Name != expected {
		t.Errorf("expected name %q; got %q", expected, metadata.Name)
	}
	if expected := "Running in the forest"; metadata.Description != expected {
		t.Errorf("expected description %q; got %q", expected, metadata.Description)
	}
	if expected := "run running forest sport sports"; metadata.Keywords != expected {
		t.Errorf("expected keywords %q; got %q", expected, metadata.Keywords)
	}
	if expected := time.Date(2015, 12, 13, 18, 35, 18, 0, time.UTC); !metadata.Time.Equal(expected) {
		t.Errorf("expected time %q; got %q", expected, metadata.Time)
	}

	testMetadataLink(t, metadata.Link)
	testMetadataCopyright(t, metadata.Copyright)
	testMetadataBounds(t, metadata.Bounds)
	testMetadataAuthor(t, metadata.Author)
}

func testMetadataLink(t *testing.T, link Link) {
	if expected := "http://www.runtastic.com"; link.Href != expected {
		t.Errorf("expected link href %q; got %q", expected, link.Href)
	}
	if expected := "runtastic"; link.Text != expected {
		t.Errorf("expected link text %q; got %q", expected, link.Text)
	}
	if expected := "text/html"; link.Type != expected {
		t.Errorf("expected link type %q; got %q", expected, link.Type)
	}
}

func testMetadataCopyright(t *testing.T, copyright Copyright) {
	if expected := "www.runtastic.com"; copyright.Author != expected {
		t.Errorf("expected copyright author %q; got %q", expected, copyright.Author)
	}
	if expected := 2015; copyright.Year != expected {
		t.Errorf("expected copyright year %d; got %d", expected, copyright.Year)
	}
	if expected := "http://www.runtastic.com"; copyright.License != expected {
		t.Errorf("expected copyright license %q; got %q", expected, copyright.License)
	}
}

func testMetadataBounds(t *testing.T, bounds Bounds) {
	if expected := 49.3956527709960938; bounds.MinLatitude != expected {
		t.Errorf("expected bounds min latitude %f; got %f", expected, bounds.MinLatitude)
	}
	if expected := 49.4017448425292969; bounds.MaxLatitude != expected {
		t.Errorf("expected bounds max latitude %f; got %f", expected, bounds.MaxLatitude)
	}
	if expected := 11.1253080368041992; bounds.MinLongitude != expected {
		t.Errorf("expected bounds min longitude %f; got %f", expected, bounds.MinLongitude)
	}
	if expected := 11.1280641555786133; bounds.MaxLongitude != expected {
		t.Errorf("expected bounds max longitude %f; got %f", expected, bounds.MaxLongitude)
	}
}

func testMetadataAuthor(t *testing.T, person Person) {
	if expected := "Runtastic"; person.Name != expected {
		t.Errorf("expected person name %q; got %q", expected, person.Name)
	}

	testMetadataAuthorEmail(t, person.Email)
}

func testMetadataAuthorEmail(t *testing.T, email Email) {
	if expected := "runtastic"; email.ID != expected {
		t.Errorf("expected email ID %q; got %q", expected, email.ID)
	}
	if expected := "example.com"; email.Domain != expected {
		t.Errorf("expected email domain %q; got %q", expected, email.Domain)
	}
}

func TestDecoderNoGPXTag(t *testing.T) {
	f, err := os.Open("test/no_gpx.gpx")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewDecoder(f).Decode()
	if err != ErrBadRootTag {
		t.Fatal("decoding should fail due to bad root tag")
	}
}

func TestDecoderGPX10(t *testing.T) {
	f, err := os.Open("test/gpx10.gpx")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewDecoder(f).Decode()
	if err != ErrGPX11Only {
		t.Fatal("decoding should fail for GPX 1.0 documents")
	}
}
