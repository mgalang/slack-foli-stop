package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nlopes/slack"
	middleware "github.com/s12i/gin-throttle"
)

// Maximum items to show in list.
const maxItems = 3

// API endpoint.
const endpoint = "https://data.foli.fi/siri/sm/"

// APIResponse interface.
type APIResponse struct {
	Status     string
	Servertime int64
	Result     []struct {
		Destinationdisplay    string
		Aimedarrivaltime      int64
		Expecteddeparturetime int64
		Lineref               string
	}
}

// SlackResponse interface
type SlackResponse struct {
	Text string `json:"text"`
}

// Get JSON from URL helper method.
func getJSON(path string, result interface{}) error {
	resp, err := http.Get(path)

	if resp.StatusCode != 200 {
		panic("Server error")
	}

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(result)
}

func validateStopCode(code string) bool {
	// Regex validation pattern.
	r, _ := regexp.Compile("(^[0-9]+)|(T[0-9]+)|(PT[0-9]+)|(L[0-9]+)")

	return r.MatchString(code)
}

// Slack request handler
func handleSlack(stopcode string) (string, bool) {
	if !validateStopCode(stopcode) {
		return "", false
	}

	j := APIResponse{}

	err := getJSON(endpoint+stopcode, &j)

	if err != nil {
		return "", false
	}

	var count = len(j.Result)

	if count > maxItems {
		count = maxItems
	}

	var responseText = ""

	loc, _ := time.LoadLocation("Europe/Helsinki")

	for index := 0; index < count; index++ {
		t := time.Unix(j.Result[index].Expecteddeparturetime, 0)

		responseText += j.Result[index].Destinationdisplay + " (" +
			j.Result[index].Lineref + "), Departure at: " +
			t.In(loc).Format("15:04") + "\n"
	}

	return responseText, true
}

// Main handler
func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	maxEventsPerSec := 2
	maxBurstSize := 2
	router.Use(middleware.Throttle(maxEventsPerSec, maxBurstSize))

	router.POST("/slack", func(c *gin.Context) {
		s, err := slack.SlashCommandParse(c.Request)

		if err != nil {
			c.JSON(http.StatusOK, SlackResponse{Text: "Command Error"})
			return
		}

		if !s.ValidateToken(os.Getenv("SECRET")) {
			c.JSON(http.StatusOK, SlackResponse{Text: "Validation Error"})
			return
		}

		responseJSON, ok := handleSlack(s.Text)

		if ok {
			c.JSON(http.StatusOK, SlackResponse{Text: responseJSON})
		} else {
			c.JSON(http.StatusOK, SlackResponse{Text: "Error"})
		}
	})

	router.Run(":" + port)
}
