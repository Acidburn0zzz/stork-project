package agentcomm

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	dbops "isc.org/stork/server/database"
	dbmodel "isc.org/stork/server/database/model"
)

// Structure representing a periodic puller which is configured to
// execute a function specified by a caller according to the timer
// interval specified in the database. The user's fuction typically
// pulls and manipulates the data from multiple apps.
type PeriodicPuller struct {
	pullerName          string
	intervalSettingName string
	pullFunc            func() (int, error)
	Db                  *dbops.PgDB
	Agents              ConnectedAgents
	Ticker              *time.Ticker
	Done                chan bool
	Wg                  *sync.WaitGroup
}

// Creates an instance of a new periodic puller. The periodic puller offers a mechanism
// to periodically trigger an action. This action is supplied as a function instance.
// This function is executed within a goroutine periodically according to the timer
// interval available in the database. The intervalSettingName is a name of this
// setting in the database. The pullerName is used for logging purposes.
func NewPeriodicPuller(db *dbops.PgDB, agents ConnectedAgents, pullerName, intervalSettingName string, pullFunc func() (int, error)) (*PeriodicPuller, error) {
	log.Printf("Starting %s Puller", pullerName)

	interval, err := dbmodel.GetSettingInt(db, intervalSettingName)
	if err != nil {
		return nil, err
	}

	periodicPuller := &PeriodicPuller{
		pullerName:          pullerName,
		intervalSettingName: intervalSettingName,
		pullFunc:            pullFunc,
		Db:                  db,
		Agents:              agents,
		Ticker:              time.NewTicker(time.Duration(interval) * time.Second),
		Done:                make(chan bool),
		Wg:                  &sync.WaitGroup{},
	}

	periodicPuller.Wg.Add(1)
	go periodicPuller.pullerLoop()

	log.Printf("Started %s Puller", periodicPuller.pullerName)
	return periodicPuller, nil
}

// Terminates the puller, i.e. the puller no longer triggers the
// user defined function.
func (puller *PeriodicPuller) Shutdown() {
	log.Printf("Stopping %s Puller", puller.pullerName)
	puller.Done <- true
	puller.Wg.Wait()
	log.Printf("Stopped %s Puller", puller.pullerName)
}

// This function controls the timing of the function execution and captures the
// termination signal.
func (puller *PeriodicPuller) pullerLoop() {
	defer puller.Wg.Done()
	for {
		select {
		// every N seconds execute user defined function
		case <-puller.Ticker.C:
			_, err := puller.pullFunc()
			if err != nil {
				log.Errorf("errors were encountered while pulling data from Kea apps: %+v", err)
			}
		// wait for done signal from shutdown function
		case <-puller.Done:
			// Make sure this function is never called again.
			puller.Ticker.Stop()
			return
		}

		// Check if the interval has changed in the database. If so, recreate the ticker.
		interval, err := dbmodel.GetSettingInt(puller.Db, puller.intervalSettingName)
		if err != nil {
			log.Errorf("problem with getting interval setting %s from db: %+v",
				puller.intervalSettingName, err)
		} else {
			puller.Ticker.Stop()
			puller.Ticker = time.NewTicker(time.Duration(interval) * time.Second)
		}
	}
}
