package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Weather struct {
	Precip1Hour       float64 `json:"precip1Hour"`
	PressureAltimeter float64 `json:"pressureAltimeter"`
	RelativeHumidity  float64 `json:"relativeHumidity"`
	Snow1Hour         float64 `json:"snow1Hour"`
	Temperature       float64 `json:"temperature"`
	WindSpeed         float64 `json:"windSpeed"`
}

func FetchWeather(Token string, lat float64, long float64) (*Weather, error) {
	url := fmt.Sprintf("https://api.weather.com/v3/wx/observations/current?apiKey=%s&geocode=%.3f%%2C%.3f&language=en-US&units=m&format=json", Token, lat, long)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var weather Weather
	err = json.Unmarshal(content, &weather)
	if err != nil {
		return nil, err
	}

	return &weather, nil
}
