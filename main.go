package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const appName = "gousdbrl"
const wiseCurrencyPage = "https://wise.com/gb/currency-converter/usd-to-brl-rate?amount=1"

type Config struct {
	Value float64
}

func getConfigFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user config directory: %w", err)
	}

	appDir := filepath.Join(configDir, "gousdbrl")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create app config directory: %w", err)
	}

	return filepath.Join(appDir, "data.json"), nil
}

func loadConfig() (*Config, error) {
	filePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	var config Config
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Value: 0.0}, nil
		}
		return nil, err
	}

	value := gjson.Get(string(data), "config.value")
	if !value.Exists() {
		return nil, fmt.Errorf("key 'config.value' not found in config file")
	}

	config.Value = value.Float()
	return &config, nil
}

func saveConfig(config *Config) error {
	filePath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	data := "{}"
	if _, err := os.Stat(filePath); err == nil {
		fileData, _ := os.ReadFile(filePath)
		data = string(fileData)
	}

	newData, err := sjson.Set(data, "config.value", config.Value)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(newData), 0644)
}

func fetchExchangeRate(url string, resultChan chan<- string, errChan chan<- error) {
	s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
	s.Start()

	resp, err := http.Get(url)
	if err != nil {
		errChan <- fmt.Errorf("failed to fetch the page: %w", err)
		s.Stop()
		return
	}
	defer resp.Body.Close()

	s.Stop()

	if resp.StatusCode != http.StatusOK {
		errChan <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		errChan <- fmt.Errorf("failed to parse HTML: %w", err)
		return
	}

	rate := doc.Find("span[dir='ltr'] span.text-success").Text()
	if rate == "" {
		errChan <- fmt.Errorf("exchange rate not found in HTML")
		return
	}

	resultChan <- rate
}

func handleRateComparison(rateStr string, configValue float64) (string, error) {
	rateFloat, err := strconv.ParseFloat(rateStr, 64)
	if err != nil {
		return "", fmt.Errorf("error parsing rate: %w", err)
	}

	arrowIcon := ""

	switch {
	case rateFloat < configValue:
		arrowIcon = color.New(color.Bold, color.FgHiRed).Sprint("▼ ")
	case rateFloat > configValue:
		arrowIcon = color.New(color.Bold, color.FgHiGreen).Sprint("▲ ")
	default:
		arrowIcon = color.New(color.Bold, color.FgHiWhite).Sprint("▶ ")
	}

	return arrowIcon + color.New(color.Bold, color.FgYellow).Sprint(rateStr), nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	loading := color.New(color.Bold, color.FgHiBlue).Sprint("Fetching USD to BRL exchange rate from ")
	highlight := color.New(color.Bold, color.FgGreen).Sprint("Wise")
	fmt.Println(loading + highlight)

	resultChan := make(chan string)
	errChan := make(chan error)

	go fetchExchangeRate(wiseCurrencyPage, resultChan, errChan)

	var config *Config
	var err error
	config, err = loadConfig()
	if err != nil {
		color.Red("Error loading config: %v\n", err)
		log.Fatalf("Error loading config: %v\n", err)
	}

	select {
	case rate := <-resultChan:
		arrowMessage, err := handleRateComparison(rate, config.Value)
		if err != nil {
			color.Red("Error handling rate comparison: %v\n", err)
			log.Fatalf("Error handling rate comparison: %v\n", err)
		}

		config.Value, err = strconv.ParseFloat(rate, 64)
		if err != nil {
			color.Red("Error updating config value: %v\n", err)
			log.Fatalf("Error updating config value: %v\n", err)
		}

		err = saveConfig(config)
		if err != nil {
			color.Red("Error saving config: %v\n", err)
			log.Fatalf("Error saving config: %v\n", err)
		}

		fmt.Println(arrowMessage)
	case err := <-errChan:
		color.Red("Error: %v\n", err)
		log.Fatalf("Error: %v\n", err)
	}
}
