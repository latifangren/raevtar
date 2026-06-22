package model

import "encoding/json"

type SEOData struct {
	Description  string
	CanonicalURL string
	ImageURL     string
	JSONLD       string
	SiteDomain   string
}

func MustJSONLD(value any) string {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(b)
}
