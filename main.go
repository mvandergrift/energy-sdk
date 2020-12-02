package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/joho/godotenv/autoload" // autoload configuration
	"golang.org/x/oauth2"

	"github.com/mvandergrift/energy-sdk/api"
	"github.com/mvandergrift/energy-sdk/api/healthmate"
	tokenAuth "github.com/mvandergrift/energy-sdk/auth"
	"github.com/mvandergrift/energy-sdk/driver"

	"github.com/mvandergrift/energy-sdk/model"
	"github.com/mvandergrift/energy-sdk/repo"
)

const defaultHoursBack = 144

func Handler(ctx context.Context, detail interface{}) error {
	log.Println("Executing AWS handler")
	var (
		err error
	)

	debugFlag := false
	startDate := ""
	endDate := ""

	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		log.Println("Debug mode activated")
		debugFlag = true
	}

	db, err := driver.OpenCn(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_DATABASE"), debugFlag)
	check("OpenCn", err)
	hc := apiFactory("healthmate", os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("CALLBACK_URL"))

	var result []model.Export
	measure, err := getMeasure(startDate, endDate, hc)
	check("GetMeasure", err)
	result = append(measure, result...)

	workouts, err := getWorkouts(startDate, endDate, hc)
	check("GetWorkouts", err)
	result = append(workouts, result...)

	for _, v := range result {
		k, err := v.Export()
		check("Export", err)

		if debugFlag {
			log.Println(v)
		}

		switch modelExport := k.(type) {
		case model.Workout:
			workoutRepo := repo.NewWorkoutRepo(db)
			check("workoutRepo.save", workoutRepo.Save(&modelExport))
		case model.Measure:
			measureRepo := repo.NewMeasureRepo(db)
			check("measureRepo.save", measureRepo.Save(&modelExport))
		default:
			panic(fmt.Sprintf("No handler for model.Export type %v", modelExport))
		}
	}

	return err
}

// todo #1 Factory pattern supports multiple data vendors @mvandergrift
func getWorkouts(startDate string, endDate string, hc api.ApiClient) ([]model.Export, error) {
	payload := url.Values{}
	payload.Set("action", "getworkouts")
	payload.Set("data_fields", "calories,effduration,intensity,manual_distance,manual_calories,hr_average,hr_min,hr_max,steps,distance,elevation,pause,hr_zone_0,hr_zone_1,hr_zone_2,hr_zone_3")

	if startDate != "" && endDate != "" {
		payload.Set("startdateymd", startDate)
		payload.Set("enddateymd", endDate)
	} else {
		last := strconv.FormatInt(time.Now().Add(defaultHoursBack*time.Hour*-1).Unix(), 10)
		payload.Set("lastupdate", last)
	}

	var result healthmate.WorkoutResult
	err := hc.ProcessRequest(payload, &result)
	if err != nil {
		return nil, err
	}

	retval := make([]model.Export, len(result.Body.Series))
	for k, v := range result.Body.Series {
		retval[k] = v
	}

	return retval, nil
}

// todo #1 Factory pattern supports multiple data vendors @mvandergrift
func getMeasure(startDate string, endDate string, hc api.ApiClient) ([]model.Export, error) {
	payload := url.Values{}
	payload.Set("action", "getmeas")
	payload.Set("meastypes", "1,6,4,11")

	if startDate != "" && endDate != "" {
		// todo #10 Verify date formats for data provider @mvandergrift
		unixStart, _ := time.Parse("2006-01-02", startDate)
		unixEnd, _ := time.Parse("2006-01-02", endDate)
		payload.Set("startdate", strconv.FormatInt(unixStart.Unix(), 10))
		payload.Set("enddate", strconv.FormatInt(unixEnd.Unix(), 10))
	} else {
		last := strconv.FormatInt(time.Now().Add(defaultHoursBack*time.Hour*-1).Unix(), 10)
		payload.Set("lastupdate", last)
	}

	var result healthmate.MeasureResult
	// todo #9 Capture and return error to caller @mvandergrift
	check("ProcessHealthmateRequest", hc.ProcessRequest(payload, &result))

	var retval []model.Export
	for _, group := range result.Body.Measuregrps {
		for _, measure := range group.Measures {
			measure.Date = group.Date
			measure.Grpid = group.Grpid
			retval = append(retval, measure)
		}
	}

	return retval, nil
}

func main() {
	if os.Getenv("_LAMBDA_SERVER_PORT") != "" {
		lambda.Start(Handler)
	} else {
		start()
	}
}

func start() {
	var (
		token *oauth2.Token
		err   error
	)

	getNewAuth := flag.Bool("auth", false, "Reauthenticate access to appliaction")
	debugFlag := flag.Bool("debug", false, "Debug database access")
	helpFlag := flag.Bool("help", false, "Show help")
	startDate := flag.String("start", "", "Start date in yyyy-mm-dd format")
	endDate := flag.String("end", "", "End date in yyyy-mm-dd format")
	typeFilter := flag.String("type", "", "Integration type to retrieve (workout, measure, etc)")
	flag.Parse()

	if *helpFlag {
		flag.PrintDefaults()
		return
	}

	db, err := driver.OpenCn(os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PWD"), os.Getenv("DB_DATABASE"), *debugFlag)
	check("OpenCn", err)

	hc := apiFactory("healthmate", os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("CALLBACK_URL"))

	if *getNewAuth { // todo #4 Transfer api authentication to user facing application @mvandergrift
		fmt.Println("Visit link to authorize account: ", hc.GetAuthCodeURL())
		fmt.Print("Enter authorization code: ")
		var code string
		fmt.Scan(&code)
		token, err = hc.GetAccessToken(code)
		check("Get access token", err)
		err = tokenAuth.SaveToken(token, db)
		check("SaveToken", err)
	}

	var result []model.Export
	if *typeFilter == "" || *typeFilter == "measure" {
		measure, err := getMeasure(*startDate, *endDate, hc)
		check("GetMeasure", err)
		result = append(measure, result...)
	}

	if *typeFilter == "" || *typeFilter == "workout" {
		workouts, err := getWorkouts(*startDate, *endDate, hc)
		check("GetWorkouts", err)
		result = append(workouts, result...)
	}

	for _, v := range result {
		k, err := v.Export()
		check("Export", err)
		log.Println(v)

		switch modelExport := k.(type) {
		case model.Workout:
			workoutRepo := repo.NewWorkoutRepo(db)
			check("workoutRepo.save", workoutRepo.Save(&modelExport))
		case model.Measure:
			measureRepo := repo.NewMeasureRepo(db)
			check("measureRepo.save", measureRepo.Save(&modelExport))
		default:
			panic(fmt.Sprintf("No handler for model.Export type %v", modelExport))
		}
	}
}

func apiFactory(provider string, clientID string, clientSecret string, callbackUrl string) api.ApiClient {
	return healthmate.NewClient(clientID, clientSecret, callbackUrl)
}

func check(subject string, err error) {
	if err != nil {
		log.Fatalf("FATAL | %v | %v", subject, err)
	}
}
