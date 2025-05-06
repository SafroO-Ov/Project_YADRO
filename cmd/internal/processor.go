package internal

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

type Processor struct {
	cfg     *Config
	logger  *log.Logger
	reg     map[int]bool
	results map[int]*Result
}

func NewProcessor(cfg *Config, logger *log.Logger) *Processor {
	return &Processor{
		cfg:     cfg,
		logger:  logger,
		reg:     make(map[int]bool),
		results: make(map[int]*Result),
	}
}

func NotStartBool(now string, start string, delta string) bool {
	t_now, _ := time.Parse("15:04:05", now)
	t_start, _ := time.Parse("15:04:05", start)
	t_delta, _ := time.Parse("15:04:05", delta)
	tdur := t_delta.Sub(time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC))
	return t_now.After(t_start.Add(tdur))
}

func DisqualLog(p *Processor, id int, t string) {
	p.logger.Printf("[%s] The competitor(%d) is disqualified", t, id)
}
func FinishedLog(p *Processor, id int, t string) {
	p.logger.Printf("[%s] The competitor(%d) has finished", t, id)
}
func (p *Processor) Process(events []Event) {
	for _, ev := range events {
		res, ok := p.results[ev.Competitor]
		if !ok {
			res = &Result{ID: ev.Competitor, Status: NotStarted}
			p.results[ev.Competitor] = res
		}

		switch ev.EventID {
		case 1:
			p.reg[ev.Competitor] = true
			p.cfg.NumOfComp++
			p.logger.Printf("[%s] The competitor(%d) registered", ev.Time, ev.Competitor)
		case 2:
			res.StartTime = ev.ExtraParams[0]
			p.logger.Printf("[%s] The start time for the competitor(%d) was set by a draw to %s", ev.Time, ev.Competitor, res.StartTime)
		case 3:
			p.logger.Printf("[%s] The competitor(%d) is on the start line", ev.Time, ev.Competitor)
			if NotStartBool(ev.Time, res.StartTime, p.cfg.StartDelta) {
				res.Status = NotStarted
				DisqualLog(p, ev.Competitor, ev.Time)
				res.FinishTime = ev.Time
			}
		case 4:
			res.ActualStart = ev.Time
			res.LapTimes = make([]string, p.cfg.Laps)
			res.LapSpeeds = make([]string, p.cfg.Laps)
			res.LapNow = 0
			p.logger.Printf("[%s] The competitor(%d) has started", ev.Time, ev.Competitor)
			if NotStartBool(ev.Time, res.StartTime, p.cfg.StartDelta) {
				res.Status = NotStarted
				DisqualLog(p, ev.Competitor, ev.Time)
				res.FinishTime = ev.Time
			}
		case 5:
			p.logger.Printf("[%s] The competitor(%d) is on the firing range(%s)", ev.Time, ev.Competitor, ev.ExtraParams[0])
		case 6:
			res.HitTargets++
			p.logger.Printf("[%s] The target(%s) has been hit by competitor(%d)", ev.Time, ev.ExtraParams[0], ev.Competitor)
		case 7:
			res.TargetsAll += 5
			p.logger.Printf("[%s] The competitor(%d) left the firing range", ev.Time, ev.Competitor)
		case 8:
			res.StartPenalty = ev.Time
			p.logger.Printf("[%s] The competitor(%d) entered the penalty laps", ev.Time, ev.Competitor)
		case 9:
			start, _ := time.Parse("15:04:05", res.StartPenalty)
			end, _ := time.Parse("15:04:05", ev.Time)
			dif := end.Sub(start)
			speed := float64(p.cfg.PenaltyLen) / dif.Seconds()
			if res.PenaltyTime.Milliseconds() != 0 {
				res.PenaltySpeed = (res.PenaltySpeed + speed) / 2
				res.PenaltyTime += dif
			} else {
				res.PenaltySpeed = speed
				res.PenaltyTime = dif
			}
			p.logger.Printf("[%s] The competitor(%d) left the penalty laps", ev.Time, ev.Competitor)
		case 10:
			end, _ := time.Parse("15:04:05", ev.Time)
			if res.StartLap == "" {
				start, _ := time.Parse("15:04:05", res.StartTime)
				dif := end.Sub(start)
				speed := fmt.Sprintf("%.3f", float64(int((float64(p.cfg.LapLen)/dif.Seconds())*1000))/1000.0)
				res.LapSpeeds[res.LapNow] = speed
				lapTime := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(dif)
				res.LapTimes[res.LapNow] = lapTime.Format("15:04:05.000")
			} else {
				start, _ := time.Parse("15:04:05", res.StartLap)
				dif := end.Sub(start)
				speed := fmt.Sprintf("%.3f", float64(int((float64(p.cfg.LapLen)/dif.Seconds())*1000))/1000.0)
				res.LapSpeeds[res.LapNow] = speed
				lapTime := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(dif)
				res.LapTimes[res.LapNow] = lapTime.Format("15:04:05.000")
			}
			res.StartLap = ev.Time
			res.LapNow++
			p.logger.Printf("[%s] The competitor(%d) ended the main lap", ev.Time, ev.Competitor)
			if res.LapNow == p.cfg.Laps {
				res.Status = Finished
				res.FinishTime = ev.Time
				FinishedLog(p, ev.Competitor, ev.Time)
			}
		case 11:
			res.Status = NotFinished
			res.FinishTime = ev.Time
			p.logger.Printf("[%s] The competitor(%d) can't continue: %s", ev.Time, ev.Competitor, strings.Join(ev.ExtraParams, " "))
		}

		for _, r := range p.results {
			start, _ := time.Parse("15:04:05", r.StartTime)
			end, _ := time.Parse("15:04:05", r.FinishTime)
			diff := end.Sub(start).Seconds()
			r.TotalTime = diff
		}
	}
}

func (p *Processor) WriteResults(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	type pair struct {
		id   int
		time float64
	}
	var list []pair
	for id, r := range p.results {
		if p.reg[id] {
			list = append(list, pair{id, r.TotalTime})
		}
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].time < list[j].time
	})
	for _, item := range list {
		r := p.results[item.id]
		fmt.Fprintf(f, "[%s] ", r.Status)
		fmt.Fprintf(f, "%d [", r.ID)
		for i := 0; i < p.cfg.Laps-1; i++ {
			fmt.Fprintf(f, "{%s, %s}, ", r.LapTimes[i], r.LapSpeeds[i])
		}
		t := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(r.PenaltyTime).Format("15:04:05.000")
		fmt.Fprintf(f, "{%s, %s}] {%s, %.3f} ", r.LapTimes[p.cfg.Laps-1], r.LapSpeeds[p.cfg.Laps-1], t, float64(int((r.PenaltySpeed*1000)))/1000.0)
		fmt.Fprintf(f, "%d/%d\n", r.HitTargets, r.TargetsAll)
	}
	return nil
}
