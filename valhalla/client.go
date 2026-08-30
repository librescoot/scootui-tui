package valhalla

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"
)

type Client struct {
	endpoint string
	client   *http.Client
}

func NewClient() *Client {
	endpoint := os.Getenv("VALHALLA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8002"
	}

	return &Client{
		endpoint: endpoint,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) SetEndpoint(url string) {
	if url != "" && url != c.endpoint {
		c.endpoint = url
	}
}

func (c *Client) CalculateRoute(startLat, startLon, endLat, endLon float64) (*Route, error) {
	req := RouteRequest{
		Locations: []Location{
			{Lat: startLat, Lon: startLon},
			{Lat: endLat, Lon: endLon},
		},
		Costing: "auto",
		DirectionsOptions: DirectionsOptions{
			Units: "kilometers",
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/route", c.endpoint)
	resp, err := c.client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to call Valhalla API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Valhalla API returned status %d: %s", resp.StatusCode, string(body))
	}

	var routeResp RouteResponse
	if err := json.NewDecoder(resp.Body).Decode(&routeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(routeResp.Trip.Legs) == 0 {
		return nil, fmt.Errorf("no route legs found")
	}

	leg := routeResp.Trip.Legs[0]

	route := &Route{
		StartLat:        startLat,
		StartLon:        startLon,
		EndLat:          endLat,
		EndLon:          endLon,
		TotalLength:     routeResp.Trip.Summary.Length,
		TotalTime:       routeResp.Trip.Summary.Time,
		Maneuvers:       leg.Maneuvers,
		CurrentManeuver: 0,
		RemainingDist:   0,
	}

	if len(route.Maneuvers) > 0 {
		route.RemainingDist = route.Maneuvers[0].Length
	}

	return route, nil
}

func (r *Route) UpdateRouteProgress(currentLat, currentLon float64) {
	if r.IsComplete() {
		return
	}

	currentManeuver := r.GetCurrentManeuver()
	if currentManeuver == nil {
		return
	}

	_ = calculateDistance(currentLat, currentLon, r.EndLat, r.EndLon)

	var totalRemaining float64
	for i := r.CurrentManeuver; i < len(r.Maneuvers); i++ {
		totalRemaining += r.Maneuvers[i].Length
	}

	r.RemainingDist = totalRemaining

	if r.CurrentManeuver < len(r.Maneuvers)-1 {
		if currentManeuver.Length > 0 && r.RemainingDist < totalRemaining-currentManeuver.Length {
			r.CurrentManeuver++
			if r.CurrentManeuver < len(r.Maneuvers) {
				r.RemainingDist = r.Maneuvers[r.CurrentManeuver].Length
			}
		}
	}
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180.0
}
