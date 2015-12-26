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
	err    error
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

	return d.consumeGPX(se), d.err
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

func (d *Decoder) consumeGPX(se xml.StartElement) (doc Document) {
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
			d.err = err
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "trk" {
				track := d.consumeTrack(se)
				if d.err != nil {
					return
				}
				doc.Tracks = append(doc.Tracks, track)
			} else if lvl == 0 && se.Name.Local == "metadata" {
				metadata := d.consumeMetadata(se)
				if d.err != nil {
					return
				}
				doc.Metadata = metadata
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "gpx" {
				return
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeMetadata(se xml.StartElement) (metadata Metadata) {
	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			d.err = err
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "time" {
				t := d.consumeTime(se)
				if d.err != nil {
					return
				}
				metadata.Time = t
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "metadata" {
				return
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeTrack(se xml.StartElement) (track Track) {
	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			d.err = err
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "trkseg" {
				seg := d.consumeSegment(se)
				if d.err != nil {
					return
				}
				track.Segments = append(track.Segments, seg)
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "trk" {
				return
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeSegment(se xml.StartElement) (seg Segment) {
	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			d.err = err
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "trkpt" {
				point := d.consumePoint(se)
				if d.err != nil {
					return
				}
				seg.Points = append(seg.Points, point)
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "trkseg" {
				return
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumePoint(se xml.StartElement) (point Point) {
	for _, a := range se.Attr {
		switch a.Name.Local {
		case "lat":
			lat, err := strconv.ParseFloat(a.Value, 64)
			if err == nil {
				point.Latitude = lat
			} else if d.Strict {
				d.err = fmt.Errorf("gpx: invalid <trkpt> lat: %s", err)
				return
			}
		case "lon":
			lon, err := strconv.ParseFloat(a.Value, 64)
			if err == nil {
				point.Longitude = lon
			} else if d.Strict {
				d.err = fmt.Errorf("gpx: invalid <trkpt> lon: %s", err)
				return
			}
		}
	}

	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			d.err = err
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			se := tok.(xml.StartElement)
			if lvl == 0 && se.Name.Local == "ele" {
				ele := d.consumeEle(se)
				if d.err != nil {
					return
				}
				point.Elevation = ele
			} else if lvl == 0 && se.Name.Local == "time" {
				t := d.consumeTime(se)
				if d.err != nil {
					return
				}
				point.Time = t
			} else if lvl == 0 && se.Name.Local == "extensions" {
				exts := d.consumeExtensions(se)
				if d.err != nil {
					return
				}
				point.Extensions = exts
			} else {
				lvl++
			}
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "trkpt" {
				return
			}
			lvl--
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeEle(se xml.StartElement) (ele float64) {
	for {
		tok, err := d.xd.Token()
		if err != nil {
			d.err = err
			return
		}
		switch tok.(type) {
		case xml.CharData:
			cd := tok.(xml.CharData)
			ele, err = strconv.ParseFloat(string(cd), 64)
			if err != nil && d.Strict {
				d.err = fmt.Errorf("gpx: invalid <ele>: %s", err)
				return
			}
		case xml.StartElement:
			d.err = errors.New("gpx: invalid <ele>")
			return
		case xml.EndElement:
			return
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeTime(se xml.StartElement) (t time.Time) {
	for {
		tok, err := d.xd.Token()
		if err != nil {
			d.err = err
			return
		}
		switch tok.(type) {
		case xml.CharData:
			cd := tok.(xml.CharData)
			t, err = time.Parse(time.RFC3339Nano, string(cd))
			if err != nil && d.Strict {
				d.err = fmt.Errorf("gpx: invalid <time>: %s", err)
				return
			}
		case xml.StartElement:
			d.err = errors.New("gpx: invalid <time>")
			return
		case xml.EndElement:
			return
		}
	}

	panic("gpx: internal error")
}

func (d *Decoder) consumeExtensions(se xml.StartElement) (tokens []xml.Token) {
	lvl := 0

	for {
		tok, err := d.xd.Token()
		if err != nil {
			d.err = err
			return
		}
		switch tok.(type) {
		case xml.StartElement:
			lvl++
		case xml.EndElement:
			ee := tok.(xml.EndElement)
			if lvl == 0 && ee.Name.Local == "extensions" {
				return
			}
			lvl--
		}
		tokens = append(tokens, xml.CopyToken(tok))
	}

	panic("gpx: internal error")
}
