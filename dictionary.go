package main

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// type globalDictionary map[string](*chan string)
type globalDictionary map[string](*node)

var (
	dictMutex = sync.RWMutex{}
)

func (gd globalDictionary) add(cid string, nptr *node) {
	gd[cid] = nptr
} //end of method

func (gd globalDictionary) remove(cid string) bool {
	dictMutex.RLock()
	defer dictMutex.RUnlock()

	if gd.contains(cid) {
		delete(gd, cid)
		return true
	} else {
		return false
	}
} //end of method

func (gd globalDictionary) contains(cid string) bool {
	//	dictMutex.RLock()
	//	defer dictMutex.RUnlock()

	j := gd[cid]

	if j == nil {
		return false
	} else {
		return true
	}

} //end of method

func (gd globalDictionary) get(cid string) *chan string {
	dictMutex.RLock()
	defer dictMutex.RUnlock()

	if gd.contains(cid) {
		return gd[cid].channel
	} else {
		return nil
	}
} //end of method

func (gd globalDictionary) getRandomKey() string {
	dictMutex.RLock()
	defer dictMutex.RUnlock()

	lenDict := len(gd)

	if lenDict > 0 {

		a := make([]string, lenDict)

		i := 0

		// populating all keys in an array
		for k := range gd {
			if gd[k].cstate != initialization {
				a[i] = k
				i++
			}
		} //end of loop

		rand.Seed(time.Now().Unix())
		rVal := rand.Intn(len(a))
		return a[rVal]

	} else {
		logger.Info("Expected dictionary size to be greater than 0 ")
		return ""
	}

} //end of method

func (gd globalDictionary) getSuccessor(key string) string {
	dictMutex.RLock()
	defer dictMutex.RUnlock()

	ky := parseInt(key)
	limit := int(powOfTwo(cf.RingSize))
	id := ky + 1

	for id != ky {

		indx := strconv.Itoa(id)

		if gd[indx] != nil {
			break
		}

		id = (id + 1) % limit

	} //end of loop

	return strconv.Itoa(id)

} //end of method

func (gd globalDictionary) getPredecessor(key string) string {
	dictMutex.RLock()
	defer dictMutex.RUnlock()

	ky := parseInt(key)
	limit := int(powOfTwo(cf.RingSize))
	pred := -1
	id := ky + 1

	for id != ky {

		indx := strconv.Itoa(id)

		if gd[indx] != nil {
			pred = id
		}

		id = (id + 1) % limit

	} //end of loop

	return strconv.Itoa(pred)

} //end of method
