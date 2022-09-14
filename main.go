package main

import (
	"encoding/json"
	"fmt"
	"kerbiot-scraper/weather"
	"log"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

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
			w, err := weather.FetchWeather(config.Token, location.Lat, location.Long)
			if err == nil {
				publish(client, toTopic(location.Name, "Rain in mm"), w.Precip1Hour)
				publish(client, toTopic(location.Name, "Pressure in hPa"), w.PressureAltimeter)
				publish(client, toTopic(location.Name, "Humidity in %"), w.RelativeHumidity)
				publish(client, toTopic(location.Name, "Snow in cm"), w.Snow1Hour) // not really sure if this is really cm 🤡
				publish(client, toTopic(location.Name, "Temperature in °C"), w.Temperature)
				publish(client, toTopic(location.Name, "Wind speed in kmh"), w.WindSpeed)
			} else {
				log.Printf("Error fetching weather data: %s", err)
			}

			airQuality, err := weather.FetchAirQuality(config.Token, location.Lat, location.Long)
			if err == nil {
				publish(client, toTopic(location.Name, "Air quality index"), float64(airQuality.AirQualityIndex))
				for _, pollutant := range airQuality.Pollutants {
					publish(client, toTopic(location.Name, "Pollutant: "+pollutant.Name+" in μgm3"), pollutant.Amount)
					publish(client, toTopic(location.Name, "Pollutant: "+pollutant.Name+" index"), float64(pollutant.CategoryIndex))
				}
			} else {
				log.Printf("Error fetching air quality data: %s", err)
			}
		}

		time.Sleep(time.Duration(config.Delay) * time.Second)
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

func toTopic(location string, key string) string {
	return location + "/" + key
}
