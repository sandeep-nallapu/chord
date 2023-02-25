package main

import (
	"encoding/json"
	"math"
)

func testMain() {
	//testJson()
	//	testFingerTable()
	//	testHashingModulo()
	//	testDictionary()
	testLog2()
}

func testJson() {
	//str := "{\"do\":\"stabilize-ring\"}"
	str := "{\"do\":\"stabilize-ring\"}"
	s := &action{}
	s.unmarshal(str)
	println(s.Do)

	a := &action{Do: fixRing, Mode: immediate}
	println(a.marshal())

	a = &action{Mode: immediate}
	println(a.marshal())

	d := &data{Key: "wow"}
	println(d.marshal())

	hq := &hashQuery{Data: d, Do: get, RespondTo: "Me-123"}
	println(hq.marshal())

	d2 := &data{Key: "wow2"}

	//arr := [2]data{*d, *d2}
	arr := [1]data{*d2}

	bytarr, err := json.Marshal(&arr)
	if check(err) {
		println(string(bytarr))
	}

	f := &fileInfo{filename: "config"}
	fb := f.read()
	c := &config{}
	c.unmarshal(string(fb))
	l := c.LiveChanges
	println(l[0].NodeId)
	println(l[1].Time)
	println(l[2].Action)
}

func testFingerTable() {
	fname := "config"
	finfo := &fileInfo{filename: fname}
	fbyte := finfo.read()
	cf = &config{}
	cf.unmarshal(string(fbyte))

	f := &fingerTable{}
	f.init()
	f.add(1, "sid")
	f.add(0, "tamshi")
	f.add(4, "sesh")
	f.add(7, "phani")
	f.delete(1)
}

func testHashingModulo() {

	s := "127.0.0.1"
	b := hash(s)
	i := powerOffset(b, 1, 5)
	logger.Info(i)

	s = "127.0.0.2"
	b = hash(s)
	i = powerOffset(b, 1, 5)
	logger.Info(i)

	s = "127.0.0.3"
	b = hash(s)
	i = powerOffset(b, 1, 5)
	logger.Info(i)

}

func testDictionary() {

	gd := globalDictionary{}
	gd.add("15", &node{})
	gd.add("27", &node{})

	logger.Info(gd.getSuccessor("3"))
	logger.Info(gd.getSuccessor("20"))
	logger.Info(gd.getSuccessor("16"))
	logger.Info(gd.getSuccessor("28"))

	logger.Info(gd.getPredecessor("3"))
	logger.Info(gd.getPredecessor("20"))
	logger.Info(gd.getPredecessor("16"))
	logger.Info(gd.getPredecessor("28"))

}

func testLog2() {
	diff := 15 - 5
	index := int(math.Log2(float64(diff)))
	println(index)

	diff = 5 - 15
	index = int(math.Log2(float64(diff)))
	println(index)
}
