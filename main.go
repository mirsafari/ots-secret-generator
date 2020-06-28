package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

// ConfigurationFile is a struct that defines script config file structure
type ConfigurationFile struct {
	Endpoint       string `json:"endpoint"`
	Username       string `json:"username"`
	APIKey         string `json:"api-key"`
	SecretTTL      int    `json:"secret-ttl"`
	PasswordLength int    `json:"password-length"`
}

// Method for ConfigurationFile struct that cheks if struct fields are empty
func (ots ConfigurationFile) isInValid() bool {
	return ots.Endpoint == "" || ots.Username == "" || ots.APIKey == ""
}

// Method that creates HTTP requests to endpoint
func (ots ConfigurationFile) makeSecrets(ch chan<- string) {
	// Build request
	// Create POST request to URL
	req, error := http.NewRequest("POST", ots.Endpoint+"/api/v1/share", nil)

	if error != nil {
		fmt.Println("Error reading request.", error.Error())
	}

	// Set username and password for request
	req.SetBasicAuth(ots.Username, ots.APIKey)
	// Generate random password that will be sent to OTS
	var newPassword = generatePassword(ots.PasswordLength)
	// Setup query params for request
	q := req.URL.Query()
	q.Add("secret", newPassword)
	q.Add("ttl", string(ots.SecretTTL))
	req.URL.RawQuery = q.Encode()

	// Create HTTP client with 10sec timeout for responses
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, error := client.Do(req)
	if error != nil {
		log.Fatal("Error reading response. ", error.Error())
	}
	defer resp.Body.Close()

	// Try to read response
	body, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		fmt.Println("Error reading body. ", error.Error())
	}

	if resp.StatusCode != 200 {
		fmt.Println("Error occured during secret generation. OTS service responded with: " + string(resp.StatusCode) + " " + string(body))
		os.Exit(2)
	}

	// Save response to map as it is a JSON, so we can access fields later
	var jsonResponse map[string]interface{}
	json.Unmarshal([]byte(body), &jsonResponse)

	// Send generated secret to channel
	ch <- fmt.Sprintf("%s -> %s/secret/%s", newPassword, ots.Endpoint, jsonResponse["secret_key"].(string))

}

// Function that checks if OTS is reachable before generation starts
func endpointReachable(ots ConfigurationFile) bool {
	// Make GET request to healthcheck service
	resp, error := http.Get(ots.Endpoint + "/api/v1/status")
	if error != nil {
		fmt.Println(error.Error())
		return false
	}

	defer resp.Body.Close()

	// Try to read response
	body, error := ioutil.ReadAll(resp.Body)
	if error != nil {
		fmt.Println("Error reading body. ", error.Error())
		return false
	}
	// Print to user healthcheck status and return true if everything went well
	fmt.Println("OTS service reachable. Healthcheck response: " + string(body))
	return true
}

// Variables for password generation
var (
	lowerCharSet   = "abcdedfghijklmnopqrst"
	upperCharSet   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	specialCharSet = "!@#$%&*"
	numberSet      = "0123456789"
	allCharSet     = lowerCharSet + upperCharSet + specialCharSet + numberSet
)

//Function that generates random passwords
// https://golangbyexample.com/generate-random-password-golang/#:~:text='math%2Frand'%20package%20of,password%20from%20a%20character%20set.
func generatePassword(passwordLength int) string {

	var password strings.Builder

	for i := 0; i < passwordLength; i++ {
		random := rand.Intn(len(allCharSet))
		password.WriteString(string(allCharSet[random]))
	}
	inRune := []rune(password.String())
	rand.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})
	return string(inRune)
}

func main() {
	//Set variables for CLI flags
	var configPath string
	var numberOfPasswords int

	//Create and parse CLI flags
	flag.StringVar(&configPath, "config", "config.json", "Path to config file")
	flag.IntVar(&numberOfPasswords, "passwords", 1, "Number of passwords you wish to generate. Max: 150")
	flag.Parse()

	//Try to open file and exit program if error occus. Print error to user
	configfile, error := os.Open(configPath)
	defer configfile.Close()
	if error != nil {
		fmt.Println(error.Error())
		os.Exit(2)
	}

	// Create variable for saving decoded JSON from config file
	ots := ConfigurationFile{}
	// Decode JSON from file to struct
	configDecoder := json.NewDecoder(configfile)

	// Check if config file is valid JSON
	if error = configDecoder.Decode(&ots); error != nil {
		fmt.Println("Config file is not in valid JSON format:", error.Error())
		os.Exit(2)
	}

	// Call isInValid method to check if config file contains valid information
	if ots.isInValid() {
		fmt.Println("Config file does not contain valid information. Please check documentation for config file structure.")
		os.Exit(2)
	} else {
		// Let user know that config file is valid and check if user wanted to generate more then 150 passwords
		fmt.Println("Configuration successfully loaded.")
		if numberOfPasswords > 150 {
			fmt.Println("You can not generate more than 150 passwords. Please specify lower number.")
			os.Exit(2)
		} else {
			// Check for OTS endpoint reachability
			if endpointReachable(ots) {
				// Create seed for random number generator and call makeSecrets method
				rand.Seed(time.Now().Unix())

				// Start secret generation
				fmt.Println("Starting secret generation ...")
				// Start timer to get metrics for duration of generating
				start := time.Now()
				// Create new channel where makeSecrets method will write responses to
				ch := make(chan string)
				// Create new OTS API request for each password
				for i := 0; i < numberOfPasswords; i++ {
					// Use goroutine for each API request so requests will will be asynchronous. Pass channel arg where request will be writeen
					go ots.makeSecrets(ch)
				}

				// Read all URLs and passwords from channel where makeSecrets writes to
				for i := 0; i < numberOfPasswords; i++ {
					fmt.Println(<-ch)
				}
				// Print User info that generation is over
				fmt.Printf("Finished. %.2fs elapsed\n", time.Since(start).Seconds())
			} else {
				fmt.Println("Could not connect to OTS serevice.")
				os.Exit(2)
			}

		}

	}
}
