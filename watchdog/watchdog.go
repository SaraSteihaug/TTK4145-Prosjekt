package watchdog

import (
	"fmt"
	"time"
)

func WatchdogCheckAlive(chWatchdogRx chan int, chActiveWatchdogs chan [3]bool) {
	fmt.Println("Starting watchdog check alive")
	elevator0Timer := time.NewTimer(5 * time.Second)
	elevator1Timer := time.NewTimer(5 * time.Second)
	elevator2Timer := time.NewTimer(5 * time.Second)

	for {
		select {
		case <-elevator0Timer.C:
			temp := <-chActiveWatchdogs
			temp[0] = false
			chActiveWatchdogs <- temp
		case <-elevator1Timer.C:
			temp := <-chActiveWatchdogs
			temp[1] = false
			chActiveWatchdogs <- temp
		case <-elevator2Timer.C:
			temp := <-chActiveWatchdogs
			temp[2] = false
			chActiveWatchdogs <- temp
		case id := <-chWatchdogRx:
			switch id {
			case 0:
				temp := <-chActiveWatchdogs
				temp[0] = true
				chActiveWatchdogs <- temp
				elevator0Timer.Reset(5 * time.Second)
			case 1:
				temp := <-chActiveWatchdogs
				temp[1] = true
				chActiveWatchdogs <- temp
				elevator1Timer.Reset(5 * time.Second)
			case 2:
				temp := <-chActiveWatchdogs
				temp[2] = true
				chActiveWatchdogs <- temp
				elevator2Timer.Reset(5 * time.Second)
			}
		}
	}
}

func WatchdogSendAlive(id int, watchdogTx chan int) {
	fmt.Println("Starting watchdog send alive")
	sendAliveTimer := time.NewTimer(1 * time.Millisecond)
	for {
		select {
		case <-sendAliveTimer.C:
			watchdogTx <- id
			sendAliveTimer.Reset(1 * time.Millisecond)
		}
	}
}
