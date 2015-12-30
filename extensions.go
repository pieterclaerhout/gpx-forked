package gpx

import (
	"encoding/xml"
	"errors"
	"strconv"
)

var (
	ErrNoSuchExtension = errors.New("gpx: no such extension")
)

// GarminTrackPointExtension is Garmin’s TrackPoint extension defined by
// https://www8.garmin.com/xmlschemas/TrackPointExtensionv1.xsd
type GarminTrackPointExtension struct {
	ATemp     float64 // Air temperature (Celsius)
	WTemp     float64 // Water temperature (Celsius)
	Depth     float64 // Diving depth (meters)
	HeartRate uint    // Heart rate (beats per minute)
	Cadence   uint    // Cadence (revs per minute)
}

const GarminTrackPointExtensionNS = "http://www.garmin.com/xmlschemas/TrackPointExtension/v1"

// ParseGarminTrackPointExtension tries to parse Garmin’s TrackPoint extension
// from a point’s extensions tokens.
func ParseGarminTrackPointExtension(tokens []xml.Token) (e GarminTrackPointExtension, err error) {
	ts := tokenStream{&sliceTokener{tokens: tokens}}

	if !findExtension(ts, GarminTrackPointExtensionNS, "TrackPointExtension") {
		return e, ErrNoSuchExtension
	}

	for {
		tok, err := ts.Token()
		if err != nil {
			return e, err
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if se.Name.Space != GarminTrackPointExtensionNS {
				ts.skipTag()
				continue
			}
			switch se.Name.Local {
			case "hr":
				s, err := ts.consumeString()
				if err != nil {
					return e, err
				}
				n, _ := strconv.Atoi(s)
				e.HeartRate = uint(n)
			case "cad":
				s, err := ts.consumeString()
				if err != nil {
					return e, err
				}
				n, _ := strconv.Atoi(s)
				e.Cadence = uint(n)
			case "atemp":
				s, err := ts.consumeString()
				if err != nil {
					return e, err
				}
				n, _ := strconv.ParseFloat(s, 64)
				e.ATemp = n
			case "wtemp":
				s, err := ts.consumeString()
				if err != nil {
					return e, err
				}
				n, _ := strconv.ParseFloat(s, 64)
				e.WTemp = n
			case "depth":
				s, err := ts.consumeString()
				if err != nil {
					return e, err
				}
				n, _ := strconv.ParseFloat(s, 64)
				e.Depth = n
			default:
				ts.skipTag()
			}
		case xml.EndElement:
			return e, nil
		}
	}
}

func findExtension(ts tokenStream, space, local string) bool {
	for {
		tok, err := ts.Token()
		if err != nil {
			return false
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if se.Name.Space == space && se.Name.Local == local {
				return true
			}
			ts.skipTag()
		case xml.EndElement:
			return false
		}
	}
}
