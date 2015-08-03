package sched

import (
	"fmt"
	"time"

	"bosun.org/cmd/bosun/cache"
	"bosun.org/cmd/bosun/conf"
	"bosun.org/slog"
)

// Run should be called once (and only once) to start all schedule activity.
func (s *Schedule) Run() error {
	if s.Conf == nil {
		return fmt.Errorf("sched: nil configuration")
	}
	s.nc = make(chan interface{}, 1)
	if s.Conf.Ping {
		go s.PingHosts()
	}
	go s.Poll()
	go s.performSave()
	for _, a := range s.Conf.Alerts {
		go s.RunAlert(a)
	}
	return nil
}

func (s *Schedule) RunAlert(a *conf.Alert) {
	for {
		wait := time.After(s.Conf.CheckFrequency * time.Duration(a.RunEvery))
		slog.Infof("starting check on %s", a.Name)
		//run the alert
		start := time.Now()
		checkTime := time.Now()    // TODO: use schedule time
		checkCache := cache.New(0) // TODO: schedule maintains current cache to share between alerts
		rh := s.NewRunHistory(checkTime, checkCache)
		for _, ak := range s.findUnknownAlerts(checkTime) { //TODO: findUnknown more targeted on alert
			if ak.Name() == a.Name {
				rh.Events[ak] = &Event{Status: StUnknown}
			}
		}
		s.CheckAlert(nil, rh, a)
		dur := time.Since(start)
		slog.Infof("check on %s took %v\n", a.Name, dur)
		start = time.Now()
		s.RunHistory(rh)
		dur = time.Since(start)
		slog.Infof("runHistory on %s took %v\n", a.Name, dur)
		<-wait
	}
}
