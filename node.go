package main

import (
	"math"
	"strconv"
	"strings"
)

type node struct {
	channel   *chan string
	channelID string
	running   bool
	predId    string
	cstate    State
	fTable    *fingerTable
	//	stateMutex sync.RWMutex
	fingersReceived int
}

var (
	gdict *globalDictionary
)

func (n *node) start(ptr *globalDictionary) {
	gdict = ptr
	n.fTable = &fingerTable{parentNodeId: n.channelID}
	//	n.stateMutex = sync.RWMutex{}
	n.cstate = initialization
	n.fingersReceived = 0
	n.fTable.init()
	n.listen()
} //end of method

func (n *node) listen() {

	n.running = true

	for n.running {

		json := <-(*n.channel)

		if strings.Contains(json, "data") {

			hq := &hashQuery{}
			hq.unmarshal(json)

			switch hq.Do {

			case get:
				k := parseInt(hq.Data.Key)

				if n.hasKey(hq.Data.Key) {
					val := dht.get(k)
					if len(val) > 0 {
						logger.Infof("[%s] Value found: %s for Key %d ", n.channelID, val, k)
						r := &response{Data: val, StateDetail: dataResponse}
						n.sendTo(hq.RespondTo, r.marshal())
					} else {
						logger.Infof("[%s] Value NOT found for Key %d ", n.channelID, k)
					}
				} else {
					id := n.fTable.lookup(k)
					n.sendTo(id, json)
				}
				break
			case remove:
				k := parseInt(hq.Data.Key)
				if n.hasKey(hq.Data.Key) {
					val := dht.get(k)
					logger.Infof("[%s] Value removed: %s for Key %d ", n.channelID, val, k)
					dht.remove(k)
				} else {
					id := n.fTable.lookup(k)
					n.sendTo(id, json)
				}
				break
			case put:
				k := parseInt(hq.Data.Key)
				if n.hasKey(hq.Data.Key) {
					logger.Infof("[%s] Key-Value added: %d-%s ", n.channelID, k, hq.Data.Value)
					dht.put(k, hq.Data.Value)
				} else {
					id := n.fTable.lookup(k)
					n.sendTo(id, json)
				}
				break
			default:
				logger.Warning("[" + n.channelID + "] Hash Query not defined yet")
			} //end of switch

		} else if strings.Contains(json, "respond-to") {

			rq := &ringQuery{}
			rq.unmarshal(json)

			switch rq.Do {

			case notify:
				logger.Info("[" + n.channelID + "] In ring notify ")
				if n.notify(rq.RespondTo) {
					n.fixFingers() // ADDED functionality for application to work
				}
				break
			case getFinger:
				if gdict.contains(rq.RespondTo) {
					if n.cstate == initialization {
						logger.Info("[" + n.channelID + "] In initialization state... Request discarded ")
						//TODO think of better solution
					} else {
						r := &response{StateDetail: rq.StateDetail}
						r.Data = n.fTable.marshal()
						n.sendResponseTo(rq.RespondTo, r)
					}
				} else {
					logger.Warning("[" + n.channelID + "] Responder does not exist in Global dictionary ")
				}
				break
			case findSuc:
				res := n.findSuccessor(rq.TargetId, rq.RespondTo)
				if len(res) > 0 {
					r := &response{Data: res, Target: rq.TargetId, StateDetail: rq.StateDetail}
					n.sendResponseTo(rq.RespondTo, r)
				} else {
					logger.Warningf("[%s] Couldn't find-successor(%s): %s ", n.channelID, rq.TargetId, res)
				}

				break
			case findPre:
				res := n.findPredecessor(rq.TargetId, rq.RespondTo)

				if len(res) > 0 {
					r := &response{Data: res, Target: rq.TargetId, StateDetail: rq.StateDetail}
					logger.Debugf("[%s] Sending find-predecessor response to %s | %s ", n.channelID, rq.RespondTo, r.marshal())
					n.sendResponseTo(rq.RespondTo, r)
				} else {
					res = gdict.getPredecessor(n.channelID)
					r := &response{Data: res, Target: rq.TargetId, StateDetail: rq.StateDetail}
					logger.Debugf("[%s] Sending find-predecessor response to %s | %s ", n.channelID, rq.RespondTo, r.marshal())
					n.sendResponseTo(rq.RespondTo, r)
				}

				break
			default:
				logger.Warning("[" + n.channelID + "] Ring Query not defined yet")

			} //end of switch

		} else if strings.Contains(json, "state") {

			r := &response{}
			r.unmarshal(json)

			logger.Debug("[" + n.channelID + "] Response received " + json)

			switch r.StateDetail {

			case populateSuccResp:
				n.initGetSuccessorFingerTable(r)
				break
			case populateTableResp:
				n.initFillFingerTable(r)
				break
			case fixFingerResp:
				n.addFinger(r)
				break
			case stabilizeResp:
				n.stabilizeSetNewPredecessor(r)
				break
			case dataResponse:
				logger.Infof("[%s] Data Response received: %s", n.channelID, r.Data)
			default:
				logger.Infof("[%s] Unknown response: %s", n.channelID, json)

			} //end of switch

		} else {

			a := &action{}
			a.unmarshal(json)

			switch a.Do {

			case join:
				logger.Info("[" + n.channelID + "] In Join Ring")
				fallthrough
			case initRing:
				logger.Info("[" + n.channelID + "] Initializing the ring fingers ")
				n.populateFingerTable(a.Sponsor)
				break
			case stabilize:
				logger.Info("[" + n.channelID + "]  Ring stabilizing in process..")
				n.stabilize()
				break
			case fixRing:
				logger.Info("[" + n.channelID + "]  Fixing ring fingers ")
				n.fixFingers()
				break
			case leave:
				logger.Info("[" + n.channelID + "]  In Leave Ring")
				n.leaveRing(a)
				logger.Info("[" + n.channelID + "]  Ring leaved!")
				break
			default:
				logger.Warning("Action not defined yet : " + json)

			} //end of switch

		} //end of type check

		json = ""

	} //end of loop

} //end of method

func (n *node) populateFingerTable(sponsor string) {

	if gdict.contains(sponsor) {

		//Ask for successor from sponsor
		rq := &ringQuery{Do: findSuc, TargetId: n.channelID, RespondTo: n.channelID, StateDetail: populateSuccResp}
		sChannel := *(gdict.get(sponsor))
		sChannel <- rq.marshal()
		logger.Info("[" + n.channelID + "] Waiting for successor response ")

		//Get successor finger table
		n.changeStateTo(populateSuccResp)

	} else if len(*gdict) <= 1 {
		logger.Info("[" + n.channelID + "] No sponsor found .... Assuming first node in the ring ")
		n.predId = n.channelID

		for i := 0; i < cf.RingSize; i++ {
			n.fTable.add(i, n.channelID)
		}

		n.changeStateTo(none)
		n.fTable.print()

	} else {
		sponsor := gdict.getRandomKey()
		logger.Infof("[%s] Sponsor NOT found... Random sponsor chosen from global dictionary -> %s ", n.channelID, sponsor)
		n.populateFingerTable(sponsor)
	}

} //end of method

// Get successor finger table, during populating finger table
func (n *node) initGetSuccessorFingerTable(r *response) {

	succResp := r.Data
	n.fTable.add(0, succResp)

	succChannel := *(gdict.get(succResp))
	rq := &ringQuery{Do: getFinger, RespondTo: n.channelID, StateDetail: populateTableResp}
	succChannel <- rq.marshal()

	n.changeStateTo(populateTableResp)

	if n.channelID != succResp {
		logger.Infof("[%s] Notifying new successor: %s ", n.channelID, succResp)
		rq := &ringQuery{Do: notify, RespondTo: n.channelID}
		succChannel <- rq.marshal()
	}

} //end of method

// If key > myID ; add to my finger table
func (n *node) initFillFingerTable(r *response) {

	succResp := n.fTable.Ftable[0]

	fingTableJson := r.Data
	ft := &fingerTable{}
	ft.unmarshal(fingTableJson)

	logger.Infof("[%s] Received table: %v | Populating my finger table ", n.channelID, ft.Ftable)

	myid := parseInt(n.channelID)
	sid := parseInt(succResp)
	totalNodes := int(powOfTwo(cf.RingSize))
	pow := 2

	for i := 1; i < n.fTable.size(); i++ {

		id := (myid + pow) % totalNodes

		if id < sid {
			n.fTable.add(i, succResp)
		} else {
			val := gdict.getSuccessor(strconv.Itoa(id))
			//			val := ft.lookup(id)
			n.fTable.add(i, val)
		} //end of if-else

		pow = pow * 2

	} //end of loop(i)

	n.changeStateTo(none)
	n.fTable.print()

} //end of method

func (n *node) leaveRing(act *action) {
	switch act.Mode {
	case orderly:

		gdict.remove(n.channelID)

		succ := n.fTable.Ftable[0]

		if len(succ) == 0 {
			succ = gdict.getSuccessor(n.channelID)
		}
		sChan := *gdict.get(succ)

		rq := &ringQuery{Do: notify, RespondTo: n.predId}
		sChan <- rq.marshal()

		n.running = false
		break
	case immediate:
		gdict.remove(n.channelID)
		n.running = false
		break
	} //end of switch

} //end of method

func (n *node) findSuccessor(targetId string, respondTo string) string {

	logger.Infof("[%s] Inside find successor | requested by: %s | target: %s ", n.channelID, respondTo, targetId)

	for i := 0; i < 5; i++ {

		succ := n.fTable.Ftable[i]

		if succ != n.channelID && gdict.contains(succ) {

			nid := parseInt(n.channelID)
			tid := parseInt(targetId)
			sid := parseInt(succ)

			if targetId == n.channelID {
				return succ
			} else if tid < sid && tid > nid {
				return succ
			} else {
				//				rq := &ringQuery{Do: findSuc, TargetId: targetId, RespondTo: respondTo}
				//				sChannel := *(gdict.get(succ))
				//				sChannel <- rq.marshal()
				f := gdict.getSuccessor(targetId)
				logger.Debugf("[%s] Successor(%s) found: %s", n.channelID, targetId, f)
				return f
			}

			break
		} //end of dictionary check

	} //end of loop

	return n.channelID

} //end of method

func (n *node) findPredecessor(targetId string, respondTo string) string {

	logger.Debug("[" + n.channelID + "]  Inside predecessor | target = " + targetId)

	if targetId == n.channelID {

		/*		if len(n.predId) == 0	{
				n.predId = gdict.getPredecessor(n.channelID)
			}*/

		return n.predId
	}

	for i := 0; i < 5; i++ {

		succ := n.fTable.Ftable[i]

		if succ != n.channelID && gdict.contains(succ) {

			if succ == targetId {
				return n.channelID
			} else {
				//				sChannel := *(gdict.get(succ))
				//				rq := &ringQuery{Do: findSuc, TargetId: targetId, RespondTo: respondTo}
				//				sChannel <- rq.marshal()
				return gdict.getPredecessor(targetId)
			}

			break
		} //end of dictionary check

	} //end of loop

	return n.channelID

} //end of method

func (n *node) fixFingers() {

	pow := 1
	myid := parseInt(n.channelID)
	totalNodes := int(powOfTwo(cf.RingSize))

	for i := 0; i < n.fTable.size(); i++ {

		rNodeId := gdict.getRandomKey()
		tid := strconv.Itoa(int(myid+pow) % totalNodes)
		rq := &ringQuery{Do: findSuc, TargetId: tid, RespondTo: n.channelID, StateDetail: fixFingerResp}

		rChannel := *(gdict.get(rNodeId))
		rChannel <- rq.marshal()

		n.changeStateTo(fixFingerResp)
		pow = pow * 2

	} //end of loop(i)

} //end of method

func (n *node) addFinger(r *response) {

	mid := parseInt(n.channelID)
	tid := parseInt(r.Target)

	if tid < mid {
		tid = int(powOfTwo(cf.RingSize)) + tid
	}

	logger.Debugf(" [%s] AddFinger | Id: %d ", n.channelID, tid)

	diff := tid - mid
	index := int(math.Log2(float64(diff)))
	logger.Infof("[%s]  Received finger entry for %d | Adding to index %d ", n.channelID, tid, index)

	v := gdict.getSuccessor(r.Target)
	//	n.fTable.add(index, r.Data)
	n.fTable.add(index, v)
	n.fingersReceived++

	if n.fingersReceived == cf.RingSize {
		logger.Infof("[%s] All fingers received | Predecessor: %s | New Table: %v ", n.channelID, n.predId, n.fTable.Ftable)
		n.fingersReceived = 0
		n.changeStateTo(none)
	}

} //end of method

func (n *node) notify(respondTo string) bool {

	sid := parseInt(n.fTable.Ftable[0])
	mid := parseInt(n.channelID)
	pid := parseInt(n.predId)
	rid := parseInt(respondTo)

	logger.Infof("[%d]  Current Predecessor Id: %d | Checking for: %d", mid, pid, rid)

	if len(n.predId) == 0 {
		logger.Infof("[%s] Predecessor updated to: %s ", n.channelID, respondTo)
		n.predId = respondTo
		return true
	} else if mid == pid {
		logger.Infof("[%s] Predecessor updated to: %s ", n.channelID, respondTo)
		n.predId = respondTo
		return true
	} else if gdict.contains(n.predId) == false {
		logger.Infof("[%s] Current Predecessor not found | New Updated to: %s ", n.channelID, respondTo)
		n.predId = respondTo
		return true
	} else if sid > mid && pid < mid {
		/** Numerically, I lie exactly between my pred and succ */

		if rid < mid && rid > pid {
			/* between me and my pred */
			logger.Infof("[%s] Predecessor updated to: %s ", n.channelID, respondTo)
			n.predId = respondTo
			return true
		} else {
			logger.Infof("[%s] Notification from %d ignored ", n.channelID, rid)
		}

	} else if pid > mid && sid > mid && sid < pid {
		/** --> so numerically my successor lie between me and pred */

		if pid < rid || mid > rid {
			logger.Infof("[%s] Predecessor updated to: %s ", n.channelID, respondTo)
			n.predId = respondTo
			return true
		} else {
			logger.Infof("[%s] Notification ignored from %d", n.channelID, rid)
		}
	} else if pid < mid && sid < mid && sid < pid {
		/** --> so numerically my predecessor lie between me and succ */

		if rid < mid && rid > sid {
			logger.Infof("[%s] Predecessor updated to: %s ", n.channelID, respondTo)
			n.predId = respondTo
			return true
		} else {
			logger.Infof("[%s] Notification ignored from %d", n.channelID, rid)
		}
	} else if pid == sid {

		logger.Infof("[%s] Predecessor updated to: %s ", n.channelID, respondTo)
		n.predId = respondTo
		return true

	} else {
		logger.Infof("[%s] Notification dicarded from %d", n.channelID, rid)
	}
	return false

} //end of method

func (n *node) stabilize() {

	succ := n.fTable.Ftable[0]

	if len(succ) > 0 && succ != n.channelID {

		rq := &ringQuery{Do: findPre, TargetId: succ, RespondTo: n.channelID, StateDetail: stabilizeResp}

		if gdict.contains(succ) {

			rjson := rq.marshal()
			logger.Infof("[%s]  Notify-> find-predecessor(%s) requested : %s ", n.channelID, succ, rjson)
			sChannel := *(gdict.get(succ))
			sChannel <- rjson

			// Predecessor of my successor
			n.changeStateTo(stabilizeResp)

		} else {
			logger.Infof(" [%s] Notify-> find-predecessor(%s) request failed | Node not found ", n.channelID, succ)
			logger.Infof(" [%s] Fixing fingers... ", n.channelID)
			n.fTable.Ftable[0] = gdict.getSuccessor(n.channelID)
			n.fixFingers()
		}

	} else if succ == n.channelID {
		logger.Infof(" [%s] Solo Node detection... ", n.channelID)
		n.predId = gdict.getPredecessor(n.channelID)
		succ = gdict.getSuccessor(n.channelID)

		if n.fTable.Ftable[0] != succ {
			n.fTable.add(0, succ)
			n.fixFingers()

			if n.channelID != succ {
				logger.Infof("[%s] Notifying new successor: %s ", n.channelID, succ)
				rq := &ringQuery{Do: notify, RespondTo: n.channelID}
				chanId := *(gdict.get(succ))
				chanId <- rq.marshal()
			}
		}
	} //end of solo check

} //end of method

// Predecessor of my successor
func (n *node) stabilizeSetNewPredecessor(r *response) {

	succ := n.fTable.Ftable[0]
	predOfSucc := r.Data

	sid := parseInt(succ)
	mid := parseInt(n.channelID)
	posid := parseInt(predOfSucc)

	update := false

	if posid == mid {

		logger.Infof("[%s] I am the Predecessor of Successor ", n.channelID)
		n.notify(succ)

	} else if sid == posid {

		logger.Infof("[%s] Successor and Predecessor of Successor is same ", n.channelID)
		n.notify(succ)

	} else if n.checkValidSuccessor(posid) {

		logger.Infof("[%s] Successor updated %s ", n.channelID, predOfSucc)
		n.fTable.add(0, predOfSucc)
		update = true

	} else {
		logger.Infof("[%s] Successor not updated. ", n.channelID)
		logger.Debugf("[%d] Successor Id: %d | Pred Of Succ: %d", mid, sid, posid)
	}

	n.changeStateTo(none)

	if update == true {

		rq := &ringQuery{Do: notify, RespondTo: n.channelID}
		nsucc := n.fTable.Ftable[0]
		sChannelPtr := gdict.get(nsucc)

		if sChannelPtr != nil {
			sChannel := *sChannelPtr
			sChannel <- rq.marshal()
		} else {
			logger.Errorf("[%s] Notify(%s) failed! ", n.channelID, nsucc)
		}

	}

	n.fTable.print()

} //end of method

func (n *node) sendTo(nodeid string, data string) bool {
	if gdict.contains(nodeid) {
		rchannel := *(gdict.get(nodeid))
		rchannel <- data
		return true
	} else {
		logger.Warning("[" + n.channelID + "] Node " + nodeid + " NOT found in global dictionary to send " + data)
		return false
	}
} //end of method

func (n *node) sendResponseTo(nodeid string, r *response) bool {
	if gdict.contains(nodeid) {
		rchannel := *(gdict.get(nodeid))
		rchannel <- r.marshal()
		return true
	} else {
		logger.Warning("[" + n.channelID + "] Node " + nodeid + " NOT found in global dictionary to respond ")
		return false
	}
} //end of method

func (n *node) hasKey(key string) bool {
	if len(key) > 0 {

		mid := parseInt(n.channelID)
		pid := parseInt(n.predId)
		k := parseInt(key)
		total := int(powOfTwo(cf.RingSize))

		if pid > mid {
			if k > pid && k < total {
				return true
			} else if k >= 0 && k <= mid {
				return true
			} else {
				return false
			}
		} else if k > pid && k <= mid {
			return true
		} else {
			return false
		}
	} else {
		logger.Infof(" [%s] Invalid key Length", n.channelID)
		return false
	}
} //end of method

func (n *node) changeStateTo(cstate State) {
	//	n.stateMutex.RLock()
	//	defer n.stateMutex.RUnlock()
	logger.Debugf(" [%s] Changing state from %d to %d ", n.channelID, n.cstate, cstate)
	n.cstate = cstate
} //end of method

func (n *node) checkValidSuccessor(id int) bool {

	mid := parseInt(n.channelID)
	sid := parseInt(n.fTable.Ftable[0])

	if sid > mid {

		if id < sid && id > mid {
			return true
		} else {
			return false
		}

	} else {

		if id > sid && id < mid {
			return false
		} else {
			return true
		}

	}

} //end of method
