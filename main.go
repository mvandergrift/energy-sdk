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

	_ "github.com/joho/godotenv/autoload" // autoload configuration
	"golang.org/x/oauth2"

	"github.com/mvandergrift/energy-sdk/auth"
	"github.com/mvandergrift/energy-sdk/driver"
	"github.com/mvandergrift/energy-sdk/healthmate"
	"github.com/mvandergrift/energy-sdk/model"
	"github.com/mvandergrift/energy-sdk/repo"
)

func main() {
	var (
		token *oauth2.Token
		err   error
	)

	getNewAuth := flag.Bool("auth", false, "Reauthenticate access to appliaction")
	debugFlag := flag.Bool("debug", false, "Debug database access")
	startDate := flag.String("start", "", "Start date in yyyy-mm-dd format")
	endDate := flag.String("end", "", "End date in yyyy-mm-dd format")
	flag.Parse()

	if *startDate == "" || *endDate == "" {
		flag.PrintDefaults()
		return
	}

	cn, err := driver.OpenCn(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_DATABASE"), *debugFlag)
	check("OpenCn", err)
	hc := healthmate.NewClient(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("CALLBACK_URL"))

	if *getNewAuth {
		fmt.Println("Visit link to authorize account: ", hc.GetAuthCodeURL())
		fmt.Print("Enter authorization code: ")
		var code string
		fmt.Scan(&code)
		token, err = hc.GetAccessToken(code)
		check("Get access token", err)
		err = auth.SaveToken(token, "withing.json")
		check("SaveToken", err)
	}

	// token, err = auth.LoadToken("withing.json")
	// check("LoadToken", err)

	payload := url.Values{}
	payload.Set("action", "getworkouts")
	payload.Set("startdateymd", *startDate)
	payload.Set("enddateymd", *endDate)
	//payload.Set("lastupdate", "1538325667")

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

	workoutRepo := repo.NewWorkoutRepo(cn)

	fmt.Printf("Found %v workouts\n", len(result.Body.Series))

	series := result.Body.Series
	for k, v := range series {
		w, err := model.NewWorkoutFromHealthmate(v)
		check("NewWorkoutFromHealthMate", err)
		check("WorkoutRepo.Save", workoutRepo.Save(&w))
		log.Printf("%v:%v [%v]", k, w.Date, w.ID)
	}
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
