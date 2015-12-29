package gpx

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"
)

const nsGPX11 = "http://www.topografix.com/GPX/1/1"

// Decoder decodes a GPX document from an input stream.
type Decoder struct {
	Strict bool
	r      io.Reader
	xd     *xml.Decoder
}

// NewDecoder creates a new decoder reading from r. The decoder
// operates in strict mode.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		Strict: true,
		r:      r,
	}
}

// Decode decodes a document.
func (d *Decoder) Decode() (doc Document, err error) {
	d.xd = xml.NewDecoder(d.r)

	se, err := d.findGPX()
	if err != nil {
		return doc, err
	}

	return d.consumeGPX(se)
}

func (d *Decoder) findGPX() (se xml.StartElement, err error) {
	for {
		tok, err := d.xd.Token()
		if err != nil {
			return se, err
		}
		if se, ok := tok.(xml.StartElement); ok {
			if se.Name.Local != "gpx" {
				return se, errors.New("gpx: root element must be <gpx>")
			}
			if se.Name.Space != nsGPX11 {
				return se, errors.New("gpx: can only parse GPX 1.1 documents")
			}
			return se, nil
		}
	}

	return se, errors.New("gpx: no start <gpx> found")
}

func (d *Decoder) consumeGPX(se xml.StartElement) (doc Document, err error) {
	for _, a := range se.Attr {
		switch a.Name.Local {
		case "version":
			doc.Version = a.Value
		}
	}

	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			return doc, err
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "trk" {
				track, err := d.consumeTrack(se)
				if err != nil {
					return doc, err
				}
				doc.Tracks = append(doc.Tracks, track)
			} else if lvl == 0 && se.Name.Local == "metadata" {
				metadata, err := d.consumeMetadata(se)
				if err != nil {
					return doc, err
				}
				doc.Metadata = metadata
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "gpx" {
				return doc, nil
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeMetadata(se xml.StartElement) (metadata Metadata, err error) {
	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			return metadata, err
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "time" {
				t, err := d.consumeTime(se)
				if err != nil {
					return metadata, err
				}
				metadata.Time = t
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "metadata" {
				return metadata, nil
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeTrack(se xml.StartElement) (track Track, err error) {
	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			return track, err
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "trkseg" {
				seg, err := d.consumeSegment(se)
				if err != nil {
					return track, err
				}
				track.Segments = append(track.Segments, seg)
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "trk" {
				return track, nil
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeSegment(se xml.StartElement) (seg Segment, err error) {
	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			return seg, err
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "trkpt" {
				point, err := d.consumePoint(se)
				if err != nil {
					return seg, err
				}
				seg.Points = append(seg.Points, point)
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "trkseg" {
				return seg, nil
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumePoint(se xml.StartElement) (point Point, err error) {
	for _, a := range se.Attr {
		switch a.Name.Local {
		case "lat":
			lat, err := strconv.ParseFloat(a.Value, 64)
			if err == nil {
				point.Latitude = lat
			} else if d.Strict {
				return point, fmt.Errorf("gpx: invalid <trkpt> lat: %s", err)
			}
		case "lon":
			lon, err := strconv.ParseFloat(a.Value, 64)
			if err == nil {
				point.Longitude = lon
			} else if d.Strict {
				return point, fmt.Errorf("gpx: invalid <trkpt> lon: %s", err)
			}
		}
	}

	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			return point, err
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "ele" {
				ele, err := d.consumeEle(se)
				if err != nil {
					return point, err
				}
				point.Elevation = ele
			} else if lvl == 0 && se.Name.Local == "time" {
				t, err := d.consumeTime(se)
				if err != nil {
					return point, err
				}
				point.Time = t
			} else if lvl == 0 && se.Name.Local == "extensions" {
				exts, err := d.consumeExtensions(se)
				if err != nil {
					return point, err
				}
				point.Extensions = exts
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "trkpt" {
				return point, nil
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeEle(se xml.StartElement) (ele float64, err error) {
	for {
		tok, err := d.xd.Token()
		if err != nil {
			return ele, err
		}
		switch tok.(type) {
		case xml.CharData:
			cd := tok.(xml.CharData)
			ele, err = strconv.ParseFloat(string(cd), 64)
			if err != nil && d.Strict {
				return ele, fmt.Errorf("gpx: invalid <ele>: %s", err)
			}
		case xml.StartElement:
			return ele, errors.New("gpx: invalid <ele>")
		case xml.EndElement:
			return ele, nil
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeTime(se xml.StartElement) (t time.Time, err error) {
	for {
		tok, err := d.xd.Token()
		if err != nil {
			return t, err
		}
		switch tok.(type) {
		case xml.CharData:
			cd := tok.(xml.CharData)
			t, err = time.Parse(time.RFC3339Nano, string(cd))
			if err != nil && d.Strict {
				return t, fmt.Errorf("gpx: invalid <time>: %s", err)
			}
		case xml.StartElement:
			return t, errors.New("gpx: invalid <time>")
		case xml.EndElement:
			return t, nil
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeExtensions(se xml.StartElement) (tokens []xml.Token, err error) {
	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			return tokens, err
		}
		switch tok.(type) {
		case xml.StartElement:
			lvl++
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "extensions" {
				return tokens, nil
			}
			lvl--
		}
		tokens = append(tokens, xml.CopyToken(tok))
	}

	panic("gpx: internal error")
}
