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
)

// RecordItem inferface.
type RecordItem struct {
	Destinationdisplay    string
	Aimedarrivaltime      int64
	Expecteddeparturetime int64
}

// SiriJSON interface.
type SiriJSON struct {
	Status     string
	Servertime int64
	Result     []RecordItem
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

// Handle the slack request.
func handleSlack(stopcode string) (string, bool) {
	if !validateStopCode(stopcode) {
		return "", false
	}

	j := SiriJSON{}

	err := getJSON("https://data.foli.fi/siri/sm/"+stopcode, &j)

	if err != nil {
		return "", false
	}

	var count = len(j.Result)

	if count > 3 {
		count = 3
	}

	var responseText = ""

	loc, _ := time.LoadLocation("Europe/Helsinki")

	for index := 0; index < count; index++ {
		t := time.Unix(j.Result[index].Expecteddeparturetime, 0)

		responseText += "Destination: " + j.Result[index].Destinationdisplay + ", Leaving at: " + t.In(loc).Format("15:04") + "\n"
	}

	return responseText, true
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	router := gin.New()
	router.Use(gin.Logger())

	router.POST("/slack", func(c *gin.Context) {
		text := c.DefaultQuery("text", "")
		s, err := slack.SlashCommandParse(c.Request)

		if err != nil {
			c.JSON(http.StatusOK, SlackResponse{Text: "Command Error"})
			return
		}

		token := os.Getenv("SECRET")

		if !s.ValidateToken(token) {
			c.JSON(http.StatusOK, SlackResponse{Text: "Validation Error"})
			return
		}

		responseJSON, ok := handleSlack(text)

		if ok {
			c.JSON(http.StatusOK, SlackResponse{Text: responseJSON})
		} else {
			c.JSON(http.StatusOK, SlackResponse{Text: "Error"})
		}
	})

	router.Run(":" + port)
}
