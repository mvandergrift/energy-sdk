package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload" // autoload configuration
	"golang.org/x/oauth2"

	"github.com/mvandergrift/energy-sdk/auth"
	"github.com/mvandergrift/energy-sdk/driver"
	"github.com/mvandergrift/energy-sdk/healthmate"
	"github.com/mvandergrift/energy-sdk/repo"
)

func main() {
	var (
		token *oauth2.Token
		err   error
	)

	getNewAuth := flag.Bool("auth", false, "Reauthenticate access to appliaction")
	debugFlag := flag.Bool("debug", false, "Debug database access")
	helpFlag := flag.Bool("help", false, "Show help")
	startDate := flag.String("start", "", "Start date in yyyy-mm-dd format")
	endDate := flag.String("end", "", "End date in yyyy-mm-dd format")
	// todo #3 Flag specifies list of types to retrieve @mvandergrift
	flag.Parse()

	if *helpFlag {
		flag.PrintDefaults()
		return
	}

	cn, err := driver.OpenCn(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_DATABASE"), *debugFlag)
	check("OpenCn", err)
	hc := healthmate.NewClient(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("CALLBACK_URL"))

	if *getNewAuth { // todo #4 Transfer api authentication to user facing application @mvandergrift
		fmt.Println("Visit link to authorize account: ", hc.GetAuthCodeURL())
		fmt.Print("Enter authorization code: ")
		var code string
		fmt.Scan(&code)
		token, err = hc.GetAccessToken(code)
		check("Get access token", err)
		err = auth.SaveToken(token, "withing.json")
		check("SaveToken", err)
	}

	workouts := GetWorkouts(*startDate, *endDate, "healthmate", hc)
	workoutRepo := repo.NewWorkoutRepo(cn)
	fmt.Printf("Found %v workouts\n", len(workouts))

	for _, v := range workouts {
		v.Display()
		w, err := v.ExportWorkout()
		check("ExportWorkout", err)
		check("WorkoutRepo.Save", workoutRepo.Save(&w))

	}
}

// todo #1 Factory pattern supports multiple data vendors @mvandergrift
func GetWorkouts(startDate string, endDate string, provider string, hc healthmate.Client) []healthmate.Export {
	payload := url.Values{}
	payload.Set("action", "getworkouts")

	if startDate != "" && endDate != "" {
		payload.Set("startdateymd", startDate)
		payload.Set("enddateymd", endDate)
	} else {
		last := strconv.FormatInt(time.Now().Add(48*time.Hour*-1).Unix(), 10)
		log.Println("last", last)
		payload.Set("lastupdate", last)
	}

	client, err := NewHTTPClient(hc, "withing.json")
	check("NewHTTPClient", err)
	resp, err := client.PostForm("https://wbsapi.withings.net/v2/measure", payload)
	check("Request workouts", err)
	defer resp.Request.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check("Read request body", err)

	var result healthmate.WorkoutResult
	err = json.Unmarshal(body, &result)
	check("Unmarshal request", err)

	retval := make([]healthmate.Export, len(result.Body.Series))
	for k, v := range result.Body.Series {
		retval[k] = v
	}

	return retval

	//return result.Body.Series
}

/*
Returns a new HTTPClient based on the Healthmate OAuth2 client & configurations. Uses
the cachedTokenPath to load and store the access token needed for api authentication
*/
func NewHTTPClient(c healthmate.Client, cachedTokenPath string) (*http.Client, error) {
	token, err := auth.LoadToken(cachedTokenPath)
	if err != nil {
		return nil, err
	}

	tokenSource := auth.RefreshToken(token, c.OAuth2Config, cachedTokenPath)
	client := oauth2.NewClient(context.Background(), *tokenSource)
	return client, nil
}

func check(subject string, err error) {
	if err != nil {
		log.Fatalf("FATAL | %v | %v", subject, err)
	}
}
