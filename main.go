package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload" // autoload configuration
	"golang.org/x/oauth2"

	"github.com/mvandergrift/energy-sdk/auth"
	"github.com/mvandergrift/energy-sdk/driver"
	"github.com/mvandergrift/energy-sdk/healthmate"

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
		log.Println(v)

		switch modelExport := k.(type) {
		case model.Workout:
			workoutRepo := repo.NewWorkoutRepo(cn)
			check("workoutRepo.save", workoutRepo.Save(&modelExport))
		case model.Measure:
			measureRepo := repo.NewMeasureRepo(cn)
			check("measureRepo.save", measureRepo.Save(&modelExport))
		default:
			panic(fmt.Sprintf("No handler for model.Export type %v", modelExport))
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
		last := strconv.FormatInt(time.Now().Add(72*time.Hour*-1).Unix(), 10)
		payload.Set("lastupdate", last)
	}

	var result healthmate.WorkoutResult
	check("ProcessHealthmateRequest", healthmate.ProcessRequest(hc, payload, &result))

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
		// todo #10 Verify date formats for data provider @mvandergrift
		unixStart, _ := time.Parse("2006-01-02", startDate)
		unixEnd, _ := time.Parse("2006-01-02", endDate)
		payload.Set("startdate", strconv.FormatInt(unixStart.Unix(), 10))
		payload.Set("enddate", strconv.FormatInt(unixEnd.Unix(), 10))
	} else {
		last := strconv.FormatInt(time.Now().Add(72*time.Hour*-1).Unix(), 10)
		payload.Set("lastupdate", last)
	}

	var result healthmate.MeasureResult
	// todo #9 Capture and return error to caller @mvandergrift
	check("ProcessHealthmateRequest", healthmate.ProcessRequest(hc, payload, &result))

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

func check(subject string, err error) {
	if err != nil {
		log.Fatalf("FATAL | %v | %v", subject, err)
	}
}
