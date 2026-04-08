package client

import (
	"context"
	"strings"
)

// Region represents a bunny.net CDN region/point of presence.
type Region struct {
	Id                  int64   `json:"Id"`
	Name                string  `json:"Name"`
	PricePerGigabyte    float64 `json:"PricePerGigabyte"`
	RegionCode          string  `json:"RegionCode"`
	ContinentCode       string  `json:"ContinentCode"`
	CountryCode         string  `json:"CountryCode"`
	Latitude            float64 `json:"Latitude"`
	Longitude           float64 `json:"Longitude"`
	AllowLatencyRouting bool    `json:"AllowLatencyRouting"`
}

// Country represents a country in the bunny.net system.
type Country struct {
	Name      string   `json:"Name"`
	IsoCode   string   `json:"IsoCode"`
	IsEU      bool     `json:"IsEU"`
	TaxRate   float64  `json:"TaxRate"`
	TaxPrefix string   `json:"TaxPrefix"`
	FlagUrl   string   `json:"FlagUrl"`
	PopList   []string `json:"PopList"`
}

// FormatPopList formats a PopList slice for display.
func FormatPopList(pops []string) string {
	return strings.Join(pops, ", ")
}

// ListRegions returns all CDN regions.
func (c *Client) ListRegions(ctx context.Context) ([]*Region, error) {
	var regions []*Region
	if err := c.Get(ctx, "/region", &regions); err != nil {
		return nil, err
	}
	return regions, nil
}

// ListCountries returns all countries.
func (c *Client) ListCountries(ctx context.Context) ([]*Country, error) {
	var countries []*Country
	if err := c.Get(ctx, "/country", &countries); err != nil {
		return nil, err
	}
	return countries, nil
}
