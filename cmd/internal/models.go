package internal

import "time"

type Config struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Start       string `json:"start"`
	StartDelta  string `json:"startDelta"`
	NumOfComp   int
}

type Event struct {
	RawLine     string
	Time        string
	EventID     int
	Competitor  int
	ExtraParams []string
}

type Status string

const (
	NotStarted   Status = "NotStarted"
	NotFinished  Status = "NotFinished"
	Disqualified Status = "Disqualified"
	Finished     Status = "Finished"
)

type Result struct {
	ID           int
	StartTime    string
	ActualStart  string
	FinishTime   string
	LapTimes     []string
	PenaltyTime  time.Duration
	LapSpeeds    []string
	PenaltySpeed float64
	StartPenalty string
	HitTargets   int
	TargetsAll   int
	LapNow       int
	StartLap     string
	Status       Status
	TotalTime    float64
}
