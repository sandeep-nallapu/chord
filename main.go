package main

import (
	log "github.com/alexcesaro/log/stdlog"

	"strconv"
)

/** Can be accessed from anywhere within the package */
var (
	logger = log.GetFromFlags()
	finfo  *fileInfo
	cf     *config
)

func main() {

	fname := "config"
	finfo := &fileInfo{filename: fname}
	fbyte := finfo.read()
	cf = &config{}
	cf.unmarshal(string(fbyte))

	//	testMain()

	c := coordinator{liveChanges: cf.LiveChanges}
	c.start()

	stabilizelock.Wait()
	logger.Close()
} //end of main

/* Checks for error and prints if found */
func check(err error) bool {
	if err != nil {
		logger.Alert(err)
		return false
	}
	return true
}

func parseInt(str string) int {

	if len(str) == 0 {
		return int(0)
	} else {
		val, err := strconv.ParseInt(str, 10, 64)

		if err != nil {
			logger.Error("Unable to parse", str, ":", err)
			return int(val)
		} else {
			return int(val)
		}
	}

}
