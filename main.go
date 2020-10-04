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
	"github.com/mvandergrift/energy-sdk/healthmate/measure"
	"github.com/mvandergrift/energy-sdk/healthmate/workout"

	"github.com/mvandergrift/energy-sdk/model"
	"github.com/mvandergrift/energy-sdk/repo"
)

var debugFlag *bool

func main() {
	var (
		token *oauth2.Token
		err   error
	)

	getNewAuth := flag.Bool("auth", false, "Reauthenticate access to appliaction")
	debugFlag = flag.Bool("debug", false, "Debug database access")
	helpFlag := flag.Bool("help", false, "Show help")
	startDate := flag.String("start", "", "Start date in yyyy-mm-dd format")
	endDate := flag.String("end", "", "End date in yyyy-mm-dd format")
	typeFilter := flag.String("type", "", "Integration type to retrieve (workout, measure, etc)")
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

	var result []model.Export
	if *typeFilter == "" || *typeFilter == "measure" {
		result = append(result, GetMeasure(*startDate, *endDate, hc)...)
	}
	if *typeFilter == "" || *typeFilter == "workout" {
		result = append(result, GetWorkouts(*startDate, *endDate, hc)...)
	}

	for _, v := range result {
		k, err := v.Export()
		check("Export", err)
		log.Println("Saving", v)

		switch modelExport := k.(type) {
		case model.Workout:
			workoutRepo := repo.NewWorkoutRepo(cn)
			check("workoutRepo.save", workoutRepo.Save(&modelExport))
		case model.Measure:
			measureRepo := repo.NewMeasureRepo(cn)
			check("measureRepo.save", measureRepo.Save(&modelExport))
		default:
		}
	}
}

// todo #1 Factory pattern supports multiple data vendors @mvandergrift
func GetWorkouts(startDate string, endDate string, hc healthmate.Client) []model.Export {
	payload := url.Values{}
	payload.Set("action", "getworkouts")
	payload.Set("data_fields", "calories,effduration,intensity,manual_distance,manual_calories,hr_average,hr_min,hr_max,steps,distance,elevation,pause,hr_zone_0,hr_zone_1,hr_zone_2,hr_zone_3")

	if startDate != "" && endDate != "" {
		payload.Set("startdateymd", startDate)
		payload.Set("enddateymd", endDate)
	} else {
		last := strconv.FormatInt(time.Now().Add(48*time.Hour*-1).Unix(), 10)
		log.Println("last", last)
		payload.Set("lastupdate", last)
	}

	var result workout.Result
	check("ProcessHealthmateRequest", ProcessHealthmateRequest(hc, payload, &result))

	retval := make([]model.Export, len(result.Body.Series))
	for k, v := range result.Body.Series {
		retval[k] = v
	}

	return retval
}

// todo #1 Factory pattern supports multiple data vendors @mvandergrift
func GetMeasure(startDate string, endDate string, hc healthmate.Client) []model.Export {
	payload := url.Values{}
	payload.Set("action", "getmeas")
	payload.Set("meastypes", "1,6,4,11")

	if startDate != "" && endDate != "" {
		unixStart, _ := time.Parse("2006-01-02", startDate)
		unixEnd, _ := time.Parse("2006-01-02", endDate)
		payload.Set("startdate", strconv.FormatInt(unixStart.Unix(), 10))
		payload.Set("enddate", strconv.FormatInt(unixEnd.Unix(), 10))
	} else {
		last := strconv.FormatInt(time.Now().Add(168*time.Hour*-1).Unix(), 10)
		log.Println("last", last)
		payload.Set("lastupdate", last)
	}

	var result measure.Result
	check("ProcessHealthmateRequest", ProcessHealthmateRequest(hc, payload, &result))

	var retval []model.Export
	for _, group := range result.Body.Measuregrps {
		for _, measure := range group.Measures {
			measure.Date = group.Date
			measure.Grpid = group.Grpid
			retval = append(retval, measure)
		}
	}

	return retval
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

func ProcessHealthmateRequest(hc healthmate.Client, payload url.Values, v interface{}) error {
	client, err := NewHTTPClient(hc, "withing.json")
	if err != nil {
		return fmt.Errorf("NewHTTPClient %w", err)
	}

	resp, err := client.PostForm("https://wbsapi.withings.net/v2/measure", payload)
	if err != nil {
		return fmt.Errorf("PostForm %w", err)
	}

	defer resp.Request.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ioutil.ReadAll %w", err)
	}

	if *debugFlag {
		err = ioutil.WriteFile("debug.json", body, 0644)
		if err != nil {
			return fmt.Errorf("WriteFile (debug) %w", err)
		}
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("Unmarshal %w", err)
	}

	return nil
}

func check(subject string, err error) {
	if err != nil {
		log.Fatalf("FATAL | %v | %v", subject, err)
	}
}
