package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Pollutant struct {
	Name          string  `json:"name"`
	Phrase        string  `json:"phrase"`
	Amount        float64 `json:"amount"`
	Unit          string  `json:"unit"`
	Category      string  `json:"category"`
	CategoryIndex int     `json:"categoryIndex"`
	Index         int     `json:"index"`
}

type Message struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type AirQuality struct {
	Latitude                     float64              `json:"latitude"`
	Longitude                    float64              `json:"longitude"`
	Source                       string               `json:"source"`
	Disclaimer                   string               `json:"disclaimer"`
	AirQualityIndex              int                  `json:"airQualityIndex"`
	AirQualityCategory           string               `json:"airQualityCategory"`
	AirQualityCategoryIndex      int                  `json:"airQualityCategoryIndex"`
	AirQualityCategoryIndexColor string               `json:"airQualityCategoryIndexColor"`
	PrimaryPollutant             string               `json:"primaryPollutant"`
	Pollutants                   map[string]Pollutant `json:"pollutants"`
	Messages                     map[string]Message   `json:"messages"`
	ExpireTimeGmt                int                  `json:"expireTimeGmt"`
}

type AirQualityResponse struct {
	AirQuality AirQuality `json:"globalairquality"`
}

func FetchAirQuality(token string, lat float64, long float64) (*AirQuality, error) {
	url := fmt.Sprintf("https://api.weather.com/v3/wx/globalAirQuality?apiKey=%s&geocode=%.3f%%2C%.3f&language=en-US&format=json&scale=UBA", token, lat, long)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var AirQualityResponse AirQualityResponse
	err = json.Unmarshal(content, &AirQualityResponse)
	if err != nil {
		return nil, err
	}

	return &AirQualityResponse.AirQuality, nil
}
