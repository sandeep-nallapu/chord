package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type coordinator struct {
	liveChanges []liveConfig
	gDict       globalDictionary
}

var (
	stabilizelock      sync.WaitGroup
	initializationlock sync.WaitGroup
	chanGroup          []chan string
	dht                DHT
)

func (cd *coordinator) start() {

	cd.gDict = make(globalDictionary)
	ringSize := powOfTwo(int(cf.RingSize))
	chanGroup = make([]chan string, ringSize)
	dht = make([]string, ringSize)
	liveChanges := cf.LiveChanges

	/** initiate the first node in the ring */
	sponsor := ""
	cd.initiateJoin(cf.Node, sponsor)

	time.Sleep(time.Duration(2500 * 1000000)) //Wait for 2.5 seconds for the first node to initialize
	go cd.sendStabilize()

	/** Send live changes from the config */
	for i := 0; i < len(liveChanges); i++ {

		timetoWait := liveChanges[i].Time * 1000000 // Nanoseconds
		actionTodo := liveChanges[i].Action
		query := liveChanges[i].Query
		nodeIp := liveChanges[i].NodeId

		if len(query) > 0 {
			cd.sendQuery(query, liveChanges[i].Data, nodeIp, timetoWait)
		} else {
			cd.sendAction(actionTodo, nodeIp, timetoWait)
		}

	} //end of loop

} //end of method

func (cd *coordinator) initiateJoin(nodeId string, sponsorId string) *node {

	index := consistentHashing(nodeId)
	chanGroup[index] = make(chan string, 100)
	cp := &chanGroup[index]

	channelId := fmt.Sprintf("%d", index)
	logger.Debug(" Initiating Join to node " + channelId)
	logger.Infof(" [%s] Mapped to %s with sponsor %s", nodeId, channelId, sponsorId)

	if cd.gDict.contains(channelId) == false {

		act := &action{Do: join, Sponsor: sponsorId}
		n := &node{channel: cp, channelID: channelId}
		cd.gDict.add(channelId, n)
		s := act.marshal()

		logger.Info(" [" + channelId + "] Starting Node ")
		go n.start(&cd.gDict)
		logger.Infof(" Sending action %s to %s ", s, channelId)
		*n.channel <- s
		return n

	} else {
		logger.Debug(" Node already exists, to initiate ")
	}

	return nil

} //end of method

func (cd *coordinator) sendAction(actionTODO NodeAction, nodeIp string, waitTime uint64) {

	logger.Debug("Sleeping before sending action to node ")
	time.Sleep(time.Duration(waitTime))
	logger.Debug("Sleep ended!")

	if actionTODO == join {
		sponsor := cd.gDict.getRandomKey()
		cd.initiateJoin(nodeIp, sponsor)
	} else {
		/** Assume ring is already there */
		chanId := consistentHashing(nodeIp)
		act := &action{Do: actionTODO}
		cd.updateAction(act, chanId)
		s := act.marshal()

		ch := gdict.get(strconv.Itoa(chanId))

		if ch != nil {
			logger.Info("Sending action: ", s)
			nChan := *ch
			nChan <- s
		} else {
			logger.Errorf("Node %s not found to send action %s", nodeIp, actionTODO)
		}

	}

} //end of method

func (cd *coordinator) updateAction(act *action, chanId int) {

	//	chid := fmt.Sprintf("%d", chanId )

	switch act.Do {

	case join:
		sponsor := cd.gDict.getRandomKey()
		act.Sponsor = sponsor
	case leave:
		rValue := rand.Intn(2)
		if rValue == 0 {
			act.Mode = orderly
		} else {
			act.Mode = immediate
		}
		break
	} //end of method

} //end of method

func (cd *coordinator) sendQuery(query HashQuery, keyValue string, nodeIp string, waitTime uint64) {

	logger.Debug("Sleeping before sending query to node ")
	time.Sleep(time.Duration(waitTime))
	logger.Debug("Sleep ended!")

	chanId := consistentHashing(nodeIp)
	nid := strconv.Itoa(chanId)

	arr := strings.Split(keyValue, "=")
	d := &data{}

	if len(arr) == 1 {
		d.Key = arr[0]
	} else if len(arr) > 1 {
		d.Key = arr[0]
		d.Value = arr[1]
	} else {
		logger.Error("No Data to perform hash query")
		return
	}

	act := &hashQuery{Do: query, Data: d, RespondTo: nid}
	s := act.marshal()

	ch := gdict.get(nid)

	if ch != nil {
		logger.Info("Sending query: ", s)
		nChan := *ch
		nChan <- s
	} else {
		logger.Errorf("Node %s[%s] not found to send {query:%s}", nodeIp, nid, query)
	}

} //end of method

func (cd *coordinator) sendStabilize() {

	stabilizelock.Add(1)

	for true {

		time.Sleep(time.Duration(cf.StabilizePeriod * 1000000))

		for k := range cd.gDict {
			logger.Info("Sending Stabilize to " + k)
			a := &action{Do: stabilize}
			channel := *(cd.gDict.get(k))
			channel <- a.marshal()
		} //end of nodes iteration

	}

	stabilizelock.Done()

} //end of method
