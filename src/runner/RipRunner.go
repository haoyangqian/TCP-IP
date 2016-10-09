package runner

// import "fmt"
import "../model"
import "time"

//import "fmt"
import "../network"

type RipRunner struct {
	ripHandler network.RipHandler

	messageChannel chan<- model.SendMessageRequest

	broadcastTimer       *time.Timer
	triggeredUpdateTimer *time.Timer
}

func (runner *RipRunner) Run() {

	runner.restartBroadcastTimer()
	runner.restartTriggeredUpdateTimer()

	for {

		select {
		case <-runner.broadcastTimer.C:
			runner.ripHandler.BroadcastAllRoutes(runner.messageChannel)
			runner.restartBroadcastTimer()
			//fmt.Println("broadcast all")
			//runner.ripHandler.UpdateRandom()

		case <-runner.triggeredUpdateTimer.C:
			runner.ripHandler.BroadcastUpdatedRoutes(runner.messageChannel)
			runner.restartTriggeredUpdateTimer()
			//fmt.Println("broadcast updated")
		}

		runner.ripHandler.ExpireRoutes()
	}
}

func (runner *RipRunner) restartBroadcastTimer() {
	runner.broadcastTimer = time.NewTimer(model.RIP_BROADCAST_INTERVAL_MILLIS * time.Millisecond)
}

func (runner *RipRunner) restartTriggeredUpdateTimer() {
	runner.triggeredUpdateTimer = time.NewTimer(model.RIP_TRIGGERED_UPDATE_INTERVAL_MILLIS * time.Millisecond)
}

func MakeRipRunner(
	ripHandler network.RipHandler,
	messageChannel chan<- model.SendMessageRequest) RipRunner {

	var timer *time.Timer

	return RipRunner{ripHandler, messageChannel, timer, timer}
}
