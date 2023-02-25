package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type table []string

type fingerTable struct {
	parentNodeId string
	Ftable       table `json:"table"`
	Length       int   `json:"Length"`
}

func (ft *fingerTable) init() {
	ft.Length = int(cf.RingSize)
	ft.Ftable = make(table, ft.Length)
} //end of method

func (ft *fingerTable) add(index int, value string) bool {
	if index >= 0 && index < ft.Length {
		ft.Ftable[index] = value
		logger.Debugf("["+ft.parentNodeId+"] Added to FingerTable | Index: %d | Value: %s ", index, value)
		return true
	} else {
		logger.Errorf("["+ft.parentNodeId+"] Invalid Index to add to FingerTable | Size: %d | Index given: %d", ft.Length, index)
		return false
	}
} //end of method

func (ft *fingerTable) lookup(key int) string {

	i := 0
	logger.Debugf("["+ft.parentNodeId+"] Looking for ", key, " with ", ft.Ftable)

	for i < ft.Length {
		n := parseInt(ft.Ftable[i])
		if int(key) < n {
			return ft.Ftable[i]
		}
		i++
	}

	l := len(ft.Ftable)
	if l > 0 {
		return ft.Ftable[l-1]
	} else {
		return ft.Ftable[0]
	}

} //end of method

func (ft *fingerTable) delete(index int) bool {
	if index >= 0 && index < ft.Length {
		ft.Ftable[index] = ""
		//todo do lookup again
		//		logger.Info(ft.Ftable)
		return true
	} else {
		logger.Errorf("["+ft.parentNodeId+"] Invalid Index to delete from FingerTable | Size: %d | Index given: %d", ft.Length, index)
		return false
	}
} //end of method

func (ft *fingerTable) size() int {
	return len(ft.Ftable)
}

func (ft *fingerTable) print() {

	var buffer bytes.Buffer
	logger.Info(" Finger Table of " + ft.parentNodeId + " is")
	buffer.WriteString("-----------\n")

	for i := 0; i < ft.size(); i++ {
		t := fmt.Sprintf("| %d | %s |\n", (i + 1), ft.Ftable[i])
		buffer.WriteString(t)
	}

	buffer.WriteString("-----------")
	fmt.Println(buffer.String())

} //end of method

func (ft *fingerTable) marshal() string {
	bytarr, err := json.Marshal(*ft)
	if check(err) {
		return string(bytarr)
	} else {
		panic(err)
	}
}
func (ft *fingerTable) unmarshal(jsonStr string) {
	json.Unmarshal([]byte(jsonStr), ft)
}

func powOfTwo(exp int) int64 {
	var res int64 = 1
	for i := 1; i <= exp; i++ {
		res *= 2
	}
	return res
} //end of method
