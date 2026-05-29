// Package maplestoryio is a thin client for the maplestory.io mob API.
package maplestoryio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Meta is the subset of mob metadata this project uses.
type Meta struct {
	ID     int
	Name   string
	Level  int
	IsBoss bool
}

// apiMob mirrors the real /mob/{id} response shape.
type apiMob struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Meta struct {
		Level  int  `json:"level"`
		IsBoss bool `json:"isBoss"`
	} `json:"meta"`
}

// Client targets a specific region+version of the API.
type Client struct {
	Region  string
	Version string
	HTTP    *http.Client
}

func New(region, version string) *Client {
	return &Client{Region: region, Version: version, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) base() string {
	return fmt.Sprintf("https://maplestory.io/api/%s/%s/mob", c.Region, c.Version)
}

func (c *Client) MetaURL(id int) string   { return fmt.Sprintf("%s/%d", c.base(), id) }
func (c *Client) RenderURL(id int) string { return fmt.Sprintf("%s/%d/render/stand", c.base(), id) }

func parseMeta(body []byte) (Meta, error) {
	var a apiMob
	if err := json.Unmarshal(body, &a); err != nil {
		return Meta{}, err
	}
	return Meta{ID: a.ID, Name: a.Name, Level: a.Meta.Level, IsBoss: a.Meta.IsBoss}, nil
}

// FetchMeta returns the mob's metadata.
func (c *Client) FetchMeta(id int) (Meta, error) {
	body, err := c.get(c.MetaURL(id))
	if err != nil {
		return Meta{}, err
	}
	return parseMeta(body)
}

// FetchRender returns the raw bytes of the mob's standing render (a GIF).
func (c *Client) FetchRender(id int) ([]byte, error) {
	return c.get(c.RenderURL(id))
}

func (c *Client) get(url string) ([]byte, error) {
	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
