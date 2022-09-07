package main

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/gocarina/gocsv"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

type circuits struct {
	CircuitID  string `csv:"circuitId" json:"circuitId"`
	CircuitRef string `csv:"circuitRef" json:"circuitRef"`
	Name       string `csv:"name" json:"name"`
	Location   string `csv:"location" json:"location"`
	Country    string `csv:"country" json:"country"`
	Lat        string `csv:"lat" json:"lat"`
	Lng        string `csv:"lng" json:"lng"`
	Alt        string `csv:"alt" json:"alt"`
	Url        string `csv:"url" json:"url"`
}

type status struct {
	StatusID string `csv:"statusId"`
	Status   string `csv:"status"`
}

type lapTimes struct {
	RaceID       string `csv:"raceId"`
	DriverID     string `csv:"driverId"`
	Lap          string `csv:"lap"`
	Position     string `csv:"position"`
	Time         string `csv:"time"`
	Milliseconds string `csv:"milliseconds"`
}

type sprintResults struct {
	ResultID       string `csv:"resultId"`
	RaceID         string `csv:"raceId"`
	DriverID       string `csv:"driverId"`
	ConstructorID  string `csv:"constructorId"`
	Number         string `csv:"number"`
	Grid           string `csv:"grid"`
	Position       string `csv:"position"`
	PositionText   string `csv:"positionText"`
	PositionOrder  string `csv:"positionOrder"`
	Points         string `csv:"points"`
	Laps           string `csv:"laps"`
	Time           string `csv:"time"`
	Milliseconds   string `csv:"milliseconds"`
	FastestLap     string `csv:"fastestLap"`
	FastestLapTime string `csv:"fastestLapTime"`
	StatusID       string `csv:"statusId"`
}

type drivers struct {
	DriverID    string `csv:"driverId"`
	DriverRef   string `csv:"driverRef"`
	Number      string `csv:"number"`
	Code        string `csv:"code"`
	Forename    string `csv:"forename"`
	Surname     string `csv:"surname"`
	Dob         string `csv:"dob"`
	Nationality string `csv:"nationality"`
	Url         string `csv:"url"`
}

type races struct {
	RaceID     string `csv:"raceId"`
	Year       string `csv:"year"`
	Round      string `csv:"round"`
	CircuitID  string `csv:"circuitId"`
	Name       string `csv:"name"`
	Date       string `csv:"date"`
	Time       string `csv:"time"`
	Url        string `csv:"url"`
	Fp1Date    string `csv:"fp1_date"`
	Fp1Time    string `csv:"fp1_time"`
	Fp2Date    string `csv:"fp2_date"`
	Fp2Time    string `csv:"fp2_time"`
	Fp3Date    string `csv:"fp3_date"`
	Fp3Time    string `csv:"fp3_time"`
	QualiDate  string `csv:"quali_date"`
	QualiTime  string `csv:"quali_time"`
	SprintDate string `csv:"sprint_date"`
	SprintTime string `csv:"sprint_time"`
}

type constructors struct {
	ConstructorID  string `csv:"constructorId" json:"constructorId"`
	ConstructorRef string `csv:"constructorRef" json:"constructorRef"`
	Name           string `csv:"name" json:"name"`
	Nationality    string `csv:"nationality" json:"nationality"`
	Url            string `csv:"url" json:"url"`
}

type constructorStandings struct {
	ConstructorStandingsID string `csv:"constructorStandingsId"`
	RaceID                 string `csv:"raceId"`
	ConstructorID          string `csv:"constructorId"`
	Points                 string `csv:"points"`
	Position               string `csv:"position"`
	PositionText           string `csv:"positionText"`
	Wins                   string `csv:"wins"`
}

type qualifying struct {
	QualifyID     string `csv:"qualifyId"`
	RaceID        string `csv:"raceId"`
	DriverID      string `csv:"driverId"`
	ConstructorID string `csv:"constructorId"`
	Number        string `csv:"number"`
	Position      string `csv:"position"`
	Q1            string `csv:"q1"`
	Q2            string `csv:"q2"`
	Q3            string `csv:"q3"`
}

type driverStandings struct {
	DriverStandingsID string `csv:"driverStandingsId"`
	RaceID            string `csv:"raceId"`
	DriverID          string `csv:"driverId"`
	Points            string `csv:"points"`
	Position          string `csv:"position"`
	PositionText      string `csv:"positionText"`
	Wins              string `csv:"wins"`
}

type constructorResults struct {
	ConstructorResultsID string `csv:"constructorResultsId"`
	RaceID               string `csv:"raceId"`
	ConstructorID        string `csv:"constructorId"`
	Points               string `csv:"points"`
	Status               string `csv:"status"`
}

type pitStops struct {
	RaceID       string `csv:"raceId"`
	DriverID     string `csv:"driverId"`
	Stop         string `csv:"stop"`
	Lap          string `csv:"lap"`
	Time         string `csv:"time"`
	Duration     string `csv:"duration"`
	Milliseconds string `csv:"milliseconds"`
}

type seasons struct {
	Year string `csv:"year"`
	Url  string `csv:"url"`
}

type nilableUint uint64

func (n *nilableUint) UnmarshalCSV(text string) error {
	if text == "\\N" {
		return nil
	}
	v, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return err
	}
	*n = nilableUint(v)
	return nil
}

type nilableFloat float64

func (n *nilableFloat) UnmarshalCSV(text string) error {
	if text == "\\N" {
		return nil
	}
	v, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return err
	}
	*n = nilableFloat(v)
	return nil
}

type results struct {
	ResultID        string       `csv:"resultId"`
	RaceID          string       `csv:"raceId"`
	DriverID        string       `csv:"driverId"`
	ConstructorID   string       `csv:"constructorId"`
	Number          nilableUint  `csv:"number"`
	Grid            int          `csv:"grid"`
	Position        nilableUint  `csv:"position"`
	PositionText    string       `csv:"positionText"`
	PositionOrder   int          `csv:"positionOrder"`
	Points          string       `csv:"points"`
	Laps            int          `csv:"laps"`
	Time            string       `csv:"time"`
	Milliseconds    nilableUint  `csv:"milliseconds"`
	FastestLap      string       `csv:"fastestLap"`
	Rank            string       `csv:"rank"`
	FastestLapTime  string       `csv:"fastestLapTime"`
	FastestLapSpeed nilableFloat `csv:"fastestLapSpeed"`
	StatusID        string       `csv:"statusId"`
}

type LapTime struct {
	Driver     drivers `json:"driver"`
	Position   int     `json:"position"`
	Time       string  `json:"time"`
	TimeMillis uint64  `json:"time_millis"`
}

type Result struct {
	Driver          drivers      `json:"driver"`
	Constructor     constructors `json:"constructor"`
	Number          int          `json:"number"`
	Grid            int          `json:"grid"`
	Position        int          `json:"position"`
	PositionText    string       `json:"positionText"`
	PositionOrder   int          `json:"positionOrder"`
	Points          float64      `json:"points"`
	Laps            int          `json:"laps"`
	Time            string       `json:"time"`
	Milliseconds    uint64       `json:"milliseconds"`
	FastestLap      string       `json:"fastestLap"`
	Rank            string       `json:"rank"`
	FastestLapTime  string       `json:"fastestLapTime"`
	FastestLapSpeed float64      `json:"fastestLapSpeed"`
	Status          string       `json:"status"`
}

type RaceData struct {
	RaceID      string            `json:"raceId"`
	Year        int               `json:"year"`
	Round       int               `json:"round"`
	Circuit     circuits          `json:"circuit"`
	Name        string            `json:"name"`
	Date        string            `json:"date"`
	Time        string            `json:"time"`
	Url         string            `json:"url"`
	Fp1Date     string            `json:"fp1_date"`
	Fp1Time     string            `json:"fp1_time"`
	Fp2Date     string            `json:"fp2_date"`
	Fp2Time     string            `json:"fp2_time"`
	Fp3Date     string            `json:"fp3_date"`
	Fp3Time     string            `json:"fp3_time"`
	QualiDate   string            `json:"quali_date"`
	QualiTime   string            `json:"quali_time"`
	SprintDate  string            `json:"sprint_date"`
	SprintTime  string            `json:"sprint_time"`
	LapTimes    [][]LapTime       `json:"lap_times"`
	Results     []Result          `json:"results"`
	tmpLapTimes map[int][]LapTime `json:"-"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func must1[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	circuitsByID := make(map[string]circuits)
	fd, err := os.Open("_tmp/f1/circuits.csv")
	must(err)
	defer fd.Close()
	var rawCircuits []circuits
	must(gocsv.UnmarshalFile(fd, &rawCircuits))
	for _, c := range rawCircuits {
		circuitsByID[c.CircuitID] = c
	}

	statusesById := make(map[string]string)
	fd, err = os.Open("_tmp/f1/status.csv")
	must(err)
	defer fd.Close()
	var rawStatus []status
	must(gocsv.UnmarshalFile(fd, &rawStatus))
	for _, rs := range rawStatus {
		statusesById[rs.StatusID] = rs.Status
	}

	resultRaces := make(map[string]RaceData)

	fd, err = os.Open("_tmp/f1/races.csv")
	must(err)
	defer fd.Close()
	var rawRaces []races
	must(gocsv.UnmarshalFile(fd, &rawRaces))
	for _, r := range rawRaces {
		resultRaces[r.RaceID] = RaceData{
			RaceID:      r.RaceID,
			Year:        must1(strconv.Atoi(r.Year)),
			Round:       must1(strconv.Atoi(r.Round)),
			Circuit:     circuitsByID[r.CircuitID],
			Name:        r.Name,
			Date:        r.Date,
			Time:        r.Fp1Time,
			Url:         r.Url,
			Fp1Date:     r.Fp1Date,
			Fp1Time:     r.Time,
			Fp2Date:     r.Fp2Date,
			Fp2Time:     r.Fp2Time,
			Fp3Date:     r.Fp3Date,
			Fp3Time:     r.Fp3Time,
			QualiDate:   r.QualiDate,
			QualiTime:   r.QualiTime,
			SprintDate:  r.SprintDate,
			SprintTime:  r.SprintTime,
			tmpLapTimes: make(map[int][]LapTime),
		}
	}

	driversByID := make(map[string]drivers)
	fd, err = os.Open("_tmp/f1/drivers.csv")
	must(err)
	defer fd.Close()
	var rawDrivers []drivers
	must(gocsv.Unmarshal(fd, &rawDrivers))
	for _, d := range rawDrivers {
		driversByID[d.DriverID] = d
	}

	constructorsByID := make(map[string]constructors)
	fd, err = os.Open("_tmp/f1/constructors.csv")
	must(err)
	defer fd.Close()
	var rawConstructors []constructors
	must(gocsv.Unmarshal(fd, &rawConstructors))
	for _, d := range rawConstructors {
		constructorsByID[d.ConstructorID] = d
	}

	fd, err = os.Open("_tmp/f1/lap_times.csv")
	must(err)
	defer fd.Close()

	var rawLapTimes []lapTimes
	must(gocsv.Unmarshal(fd, &rawLapTimes))
	for _, r := range rawLapTimes {
		race := resultRaces[r.RaceID]
		lap := must1(strconv.Atoi(r.Lap))
		race.tmpLapTimes[lap-1] = append(race.tmpLapTimes[lap-1], LapTime{
			Driver:     driversByID[r.DriverID],
			Position:   must1(strconv.Atoi(r.Position)),
			Time:       r.Time,
			TimeMillis: must1(strconv.ParseUint(r.Milliseconds, 10, 64)),
		})
		resultRaces[r.RaceID] = race
	}

	fd, err = os.Open("_tmp/f1/results.csv")
	must(err)
	defer fd.Close()
	var rawResults []results
	must(gocsv.UnmarshalFile(fd, &rawResults))
	for _, r := range rawResults {
		race := resultRaces[r.RaceID]
		res := Result{
			Driver:          driversByID[r.DriverID],
			Constructor:     constructorsByID[r.ConstructorID],
			Number:          int(r.Number),
			Grid:            r.Grid,
			Position:        int(r.Position),
			PositionText:    r.PositionText,
			PositionOrder:   r.PositionOrder,
			Laps:            r.Laps,
			Time:            r.Time,
			Milliseconds:    uint64(r.Milliseconds),
			FastestLap:      r.FastestLap,
			Rank:            r.Rank,
			FastestLapTime:  r.FastestLapTime,
			FastestLapSpeed: float64(r.FastestLapSpeed),
			Status:          statusesById[r.StatusID],
		}
		if r.Points != "\\N" {
			res.Points = must1(strconv.ParseFloat(r.Points, 64))
		}
		race.Results = append(race.Results, res)
		resultRaces[r.RaceID] = race
	}

	for rid, race := range resultRaces {
		race.LapTimes = make([][]LapTime, len(race.tmpLapTimes))
		for lap, lt := range race.tmpLapTimes {
			slices.SortFunc(lt, func(a, b LapTime) bool {
				return a.Position < b.Position
			})
			race.LapTimes[lap] = lt
		}
		resultRaces[rid] = race
	}

	fd, err = os.OpenFile("races.json", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	must(err)
	defer fd.Close()
	must(json.NewEncoder(fd).Encode(maps.Values(resultRaces)))
}
