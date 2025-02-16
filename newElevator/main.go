package main

import (
	"elevatorlib/elevator"
	"elevatorlib/elevator/runElevator"
	"elevatorlib/elevio"
	"elevatorlib/network/bcast"
	"elevatorlib/requestAsigner"
	"elevatorlib/watchdog"
	"flag"
	"fmt"
)

func checkMaster(chMasterState chan bool, id int, activeElevators [3]bool) {
	fmt.Println("activeElevators changed checking master")
	if id != 0 {
		if activeElevators[id-1] || activeElevators[0] {
			fmt.Println("Elevator:", id, "is slave")
			chMasterState <- false
		} else {
			fmt.Println("Elevator:", id, "is master")
			fmt.Println("activeElevators", activeElevators)
			chMasterState <- true
		}
	} else {
		fmt.Println("Elevator:", id, "is master")
		fmt.Println("activeElevators", activeElevators)
		chMasterState <- true
	}
	return
}

type hallRequests map[string][][2]int

func main() {
	fmt.Println("Starting main")
	id := 0
	port := 15657
	flag.IntVar(&id, "id", 0, "id of this elevator")
	flag.IntVar(&port, "port", 15657, "port of this elevator")
	flag.Parse()

	//check master state based on flag input
	//chanels
	chMasterState := make(chan bool)

	chElevatorTx := make(chan elevator.Elevator)
	chElevatorRx := make(chan elevator.Elevator)

	elevatorStatuses := make([]elevator.Elevator, 3)
	chElevatorStatuses := make(chan []elevator.Elevator)

	//Used for sending hall request too elevators from request assigner
	chAssignedHallRequestsTx := make(chan requestAsigner.HallRequests)
	chAssignedHallRequestsRx := make(chan requestAsigner.HallRequests)

	//used for sending new hall requests from elevators to request assigner
	chNewHallRequestTx := make(chan elevio.ButtonEvent)
	chNewHallRequestRx := make(chan elevio.ButtonEvent)
	//used to send information about cleared hall requests from elevators to request assigner 
	chHallRequestClearedTx := make(chan elevio.ButtonEvent)
	chHallRequestClearedRx := make(chan elevio.ButtonEvent)

	//used for sending elevator alive signal to watchdog
	chWatchdogTx := make(chan int)
	chWatchdogRx := make(chan int)
	activeWatchdogs := [3]bool{false, false, false}
	chActiveWatchdogs := make(chan [3]bool)

	//used for updating the active watchdogs array (checking which elevators are still alive)
	fmt.Println("Starting broadcast of, elevator, hallRequest and watchdog")
	//transmitter and receiver for elevator states
	go bcast.Transmitter(2000, chElevatorTx)
	go bcast.Receiver(2000, chElevatorRx)
	//transmitter and receiver for assigned hall requests
	go bcast.Transmitter(3001, chAssignedHallRequestsTx)
	go bcast.Receiver(3001, chAssignedHallRequestsRx)
	//transmitter and receiver for local hall requests
	go bcast.Transmitter(3002, chNewHallRequestTx)
	go bcast.Receiver(3002, chNewHallRequestRx)
	//transmitter and receiver for cleared hall requests
	go bcast.Transmitter(3003, chHallRequestClearedTx)
	go bcast.Receiver(3003, chHallRequestClearedRx)
	//transmitter and receiver for watchdog
	go bcast.Transmitter(4001, chWatchdogTx)
	go bcast.Receiver(4001, chWatchdogRx)

	//functions for checking the watchdog and sending alive signal
	go watchdog.WatchdogSendAlive(id, chWatchdogTx)
	go watchdog.WatchdogCheckAlive(chWatchdogRx, chActiveWatchdogs)

	//functions for running the local elevator
	go runElevator.RunLocalElevator(chElevatorTx, chNewHallRequestTx, chAssignedHallRequestsRx, chHallRequestClearedTx, id, port)

	//function for assigning hall request to slave elevators
	go requestAsigner.RequestAsigner(chNewHallRequestRx, chElevatorStatuses, chMasterState, chHallRequestClearedRx, chAssignedHallRequestsTx) //jobbe med den her
	
	activeWatchdogs[id] = true
	chActiveWatchdogs <- activeWatchdogs
	
	fmt.Println("Starting main loop")
	for {
		select {
		case elevator := <-chElevatorRx:
			elevatorStatuses[elevator.Id] = elevator
			chElevatorStatuses <- elevatorStatuses

		case temp := <-chActiveWatchdogs:
			for i := 0; i < 3; i++ {
				if !temp[i] {
					elevatorStatuses[i].Behaviour = elevator.EB_Disconnected
				}
			}
			activeWatchdogs = temp
			go checkMaster(chMasterState, id, activeWatchdogs)
		}
	}
}
