package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Weather struct {
	Precip1Hour       float64 `json:"precip1Hour"`
	PressureAltimeter float64 `json:"pressureAltimeter"`
	RelativeHumidity  float64 `json:"relativeHumidity"`
	Snow1Hour         float64 `json:"snow1Hour"`
	Temperature       float64 `json:"temperature"`
	WindSpeed         float64 `json:"windSpeed"`
}

type Location struct {
	Name string  `json:"Name"`
	Lat  float64 `json:"Lat"`
	Long float64 `json:"Long"`
}

type Config struct {
	Token string `json:"Token"`
	Delay int    `json:"Delay"`
	MQTT  struct {
		Broker   string `json:"Broker"`
		Port     int    `json:"Port"`
		Username string `json:"Username"`
		Password string `json:"Password"`
	} `json:"MQTT"`
	Locations []Location `json:"Locations"`
}

func main() {
	log.Println("Starting kerbiot-scraper ...")

	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Error loading scraper.json: %s", err)
	}

	client, err := connectMQTT(config)
	if err != nil {
		log.Fatalf("Error connecting to MQTT broker: %s", err)
	}

	for {
		for _, location := range config.Locations {
			weather, err := fetchWeather(config.Token, &location)
			if err == nil {
				publish(client, toTopic(location.Name, "Rain"), weather.Precip1Hour)
				publish(client, toTopic(location.Name, "Pressure"), weather.PressureAltimeter)
				publish(client, toTopic(location.Name, "Humiditiy"), weather.RelativeHumidity)
				publish(client, toTopic(location.Name, "Snow"), weather.Snow1Hour)
				publish(client, toTopic(location.Name, "Temperature"), weather.Temperature)
				publish(client, toTopic(location.Name, "Wind speed"), weather.WindSpeed)
			} else {
				log.Printf("Error fetching weather data: %s", err)
			}
		}

		time.Sleep(time.Second * 1000000000)
	}
}

func loadConfig() (*Config, error) {
	raw, err := os.ReadFile("./scraper.json")
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func connectMQTT(config *Config) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.MQTT.Broker, config.MQTT.Port))
	opts.SetClientID("kerbiot-scraper")
	opts.SetUsername(config.MQTT.Username)
	opts.SetPassword(config.MQTT.Password)
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return client, nil
}

func publish(client mqtt.Client, topic string, value float64) {
	result := client.Publish(topic, 0, false, fmt.Sprintf("%f", value))
	result.Wait()

	if result.Error() != nil {
		log.Fatalf("WARNING: Failed to publish value to topic '%s': %s", topic, result.Error())
	}
}

func fetchWeather(Token string, location *Location) (*Weather, error) {
	url := fmt.Sprintf("https://api.weather.com/v3/wx/observations/current?apiKey=%s&geocode=%.3f%%2C%.3f&language=en-US&units=m&format=json", Token, location.Lat, location.Long)
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

func toTopic(location string, key string) string {
	return location + "/" + key
}
