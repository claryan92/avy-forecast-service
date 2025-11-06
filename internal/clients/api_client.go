package clients

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "net/url"

    "example.com/avalanche/internal/models"
)

type HTTPClient struct {
    Client *http.Client
}

func (c *HTTPClient) Get(u string) (*http.Response, error) {
    return c.Client.Get(u)
}

type AvalancheAPIClient struct {
    BaseURL    string
    HTTPClient *HTTPClient
}

func NewAvalancheAPIClient(baseURL string, client *HTTPClient) *AvalancheAPIClient {
    return &AvalancheAPIClient{BaseURL: baseURL, HTTPClient: client}
}

func (a *AvalancheAPIClient) FetchForecasts(centerID string) ([]models.Forecast, error) {
    u, err := url.Parse(a.BaseURL + "/products")
    if err != nil {
        return nil, err
    }
    q := u.Query()
    q.Set("avalanche_center_id", centerID)
    u.RawQuery = q.Encode()

    resp, err := a.HTTPClient.Get(u.String())
    if err != nil {
        return nil, fmt.Errorf("fetch failed for %s: %w", centerID, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
        return nil, fmt.Errorf("bad response: %s", string(body))
    }

    var data []models.Forecast
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, fmt.Errorf("decode error: %w", err)
    }
    return data, nil
}
