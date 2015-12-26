package gpx

import (
	"encoding/xml"
	"time"
)

// Document represents a GPX document.
type Document struct {
	Version  string
	Metadata Metadata
	Tracks   []Track
}

// Metadata provides additional information about a GPX document.
type Metadata struct {
	Time time.Time
}

// Track represents a track.
type Track struct {
	Segments []Segment
}

// Segments represents a track segment.
type Segment struct {
	Points []Point
}

// Point represents a track point. Extensions contains the raw XML tokens
// of the point’s extensions if it has any (excluding the <extensions>
// start and end tag).
type Point struct {
	Latitude   float64
	Longitude  float64
	Elevation  float64
	Time       time.Time
	Extensions []xml.Token
}
