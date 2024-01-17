package golsv

import (
	"flag"
	"fmt"
	"github.com/pkg/profile"
	"os"
	"os/signal"
	"syscall"
)

type ProfileArgs struct {
	ProfileType string
	Toggled bool
	stopper interface{ Stop() }
}

func (P *ProfileArgs) ConfigureFlags() {
	flag.StringVar(&P.ProfileType, "profile", "cpu", "enable profiling: cpu or mem")
	flag.BoolVar(&P.Toggled, "profile-toggle", true, "toggle profiling on/off by sending USR1 signal")
}	

func (P *ProfileArgs) prepareToggle() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		for {
			<-c
			if P.stopper != nil {
				P.stopper.Stop()
				P.stopper = nil
			} else {
				P.startProfiling()
			}
		}
	}()
}

func (P *ProfileArgs) Start() {
	if P.Toggled {
		P.prepareToggle()
	} else {
		P.startProfiling()
	}
}

func (P *ProfileArgs) startProfiling() {
	switch P.ProfileType {
	case "cpu":
		P.stopper = profile.Start(profile.CPUProfile, profile.ProfilePath("."))
	case "mem":
		P.stopper = profile.Start(profile.MemProfile, profile.ProfilePath("."))
	case "":
		// do nothing
	default:
		panic(fmt.Sprintf("Uknown profile type: %s", P.ProfileType))
	}
}

func (P *ProfileArgs) Stop() {
	if P.stopper != nil {
		P.stopper.Stop()
		P.stopper = nil
	}
}

