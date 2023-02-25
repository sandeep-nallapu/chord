package main

import "encoding/json"

type jsonObject interface {
	marshal() string
	unmarshal(str string)
}

/** =========================================== JSON structure ============================================== */
type action struct {
	Do      NodeAction `json:"do"`
	Sponsor string     `json:"sponsoring-node,omitempty"`
	Mode    ActionMode `json:"mode,omitempty"`
}

func (act *action) marshal() string {
	bytarr, err := json.Marshal(*act)
	if check(err) {
		return string(bytarr)
	} else {
		panic(err)
	}
}
func (act *action) unmarshal(jsonStr string) {
	json.Unmarshal([]byte(jsonStr), act)
}

/** =========================================== JSON structure ============================================== */
type ringQuery struct {
	Do          NodeQuery `json:"do"`
	RespondTo   string    `json:"respond-to"`
	TargetId    string    `json:"target-id,omitempty"`
	StateDetail State     `json:"state,omitempty"`
}

func (rq *ringQuery) marshal() string {
	bytarr, err := json.Marshal(*rq)
	if check(err) {
		return string(bytarr)
	} else {
		panic(err)
	}
}
func (rq *ringQuery) unmarshal(jsonStr string) {
	json.Unmarshal([]byte(jsonStr), rq)
}

/** =========================================== JSON structure ============================================== */
type hashQuery struct {
	Do        HashQuery `json:"do"`
	Data      *data     `json:"data"`
	RespondTo string    `json:"respond-to"`
}

func (hq *hashQuery) marshal() string {
	bytarr, err := json.Marshal(*hq)
	if check(err) {
		return string(bytarr)
	} else {
		panic(err)
	}
}
func (hq *hashQuery) unmarshal(jsonStr string) {
	json.Unmarshal([]byte(jsonStr), hq)
}

/** =========================================== JSON structure ============================================== */
type data struct {
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

func (dt *data) marshal() string {
	bytarr, err := json.Marshal(*dt)
	if check(err) {
		return string(bytarr)
	} else {
		panic(err)
	}
}
func (dt *data) unmarshal(jsonStr string) {
	json.Unmarshal([]byte(jsonStr), dt)
}

/** =========================================== JSON structure ============================================== */
type response struct {
	StateDetail State  `json:"state"`
	Data        string `json:"response"`
	Target      string `json:"target",omitempty`
}

func (r *response) marshal() string {
	bytarr, err := json.Marshal(*r)
	if check(err) {
		return string(bytarr)
	} else {
		panic(err)
	}
}
func (r *response) unmarshal(jsonStr string) {
	json.Unmarshal([]byte(jsonStr), r)
}

/** =========================================== JSON structure ============================================== */
type config struct {
	RingSize        int          `json:"ring.size"`
	Node            string       `json:"startup.node.id"`
	StabilizePeriod int64        `json:"stabilize.period.millis"`
	LiveChanges     []liveConfig `json:"liveChanges,omitempty"`
}
type liveConfig struct {
	NodeId string     `json:"id"`
	Time   uint64     `json:"timeInMillis"`
	Action NodeAction `json:"action,omitempty"`
	Query  HashQuery  `json:"query,omitempty"`
	Data   string     `json:"data,omitempty"`
}

func (cfg *config) marshal() string {
	bytarr, err := json.Marshal(*cfg)
	if check(err) {
		return string(bytarr)
	} else {
		panic(err)
	}
}
func (cfg *config) unmarshal(jsonStr string) {
	json.Unmarshal([]byte(jsonStr), cfg)
}

func (lc *liveConfig) marshal() string {
	bytarr, err := json.Marshal(*lc)
	if check(err) {
		return string(bytarr)
	} else {
		panic(err)
	}
}
func (lc *liveConfig) unmarshal(jsonStr string) {
	json.Unmarshal([]byte(jsonStr), lc)
}
