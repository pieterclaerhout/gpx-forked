package gpx

import (
	"encoding/xml"
	"errors"
	"io"
)

// A tokener provides tokens.
type tokener interface {
	Token() (xml.Token, error)
}

// A sliceTokener provides tokens from a slice.
type sliceTokener struct {
	tokens []xml.Token
}

func (t *sliceTokener) Token() (xml.Token, error) {
	tok := t.tokens[0]
	if tok == nil {
		return nil, io.EOF
	}
	t.tokens = t.tokens[1:]
	return tok, nil
}

// A tokenStream operates on a stream of tokens from a tokener.
type tokenStream struct {
	tokener
}

func (ts *tokenStream) consumeString() (string, error) {
	var s string
	for {
		tok, err := ts.Token()
		if err != nil {
			return "", err
		}
		switch tok.(type) {
		case xml.CharData:
			s += string(tok.(xml.CharData))
		case xml.EndElement:
			return s, nil
		default:
			return "", errors.New("gpx: unexpected element while reading string")
		}
	}
}

func (ts *tokenStream) skipTag() error {
	lvl := 0
	for {
		tok, err := ts.Token()
		if err != nil {
			return err
		}
		switch tok.(type) {
		case xml.StartElement:
			lvl++
		case xml.EndElement:
			if lvl == 0 {
				return nil
			}
			lvl--
		}
	}
}
