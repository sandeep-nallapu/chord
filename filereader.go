package main

import "os"

/* Structure to save file information */
type fileInfo struct {
	filename   string
	filesize   uint64
	start, end uint64
}

/* Method to update file information like file size, used to distribute load to workers   */
func (freader *fileInfo) info() {

	file, err := os.Open(freader.filename)

	if check(err) {
		freader.filesize = getFileSize(file)
		logger.Infof("[%s] Filesize: %d bytes", freader.filename, freader.filesize)
		file.Close()
	} else {
		logger.Alert("Unable to open file:", freader.filename)
		logger.Critical("Closing application...")
		panic("Need filesize to distribute load to workers")
	} //end of file open check

} //end of method

/* Reads the file, based on start and end location specified in the structure */
func (freader *fileInfo) read() []byte {

	file, err := os.Open(freader.filename)

	if check(err) {

		freader.filesize = getFileSize(file)

		/* Double check seek endpoints */
		/* No need of freader.start>=0 check, as its unsigned int */
		if freader.start < freader.end && freader.end <= freader.filesize {

			offset := int(freader.end - freader.start + 1)
			_, err := file.Seek(int64(freader.start), 0)

			if check(err) {
				var data = make([]byte, offset)
				logger.Debugf(" Reading %d bytes from %s ", offset, freader.filename)
				file.Read(data)
				return data
			} else {
				logger.Errorf("Unable to read file using seek values [%d,%d] ", freader.start, freader.end)
			}

		} else if freader.start == 0 && freader.end == 0 {
			/* Only if start and end is not defined the values of both would be 0, implying read the whole file */
			var data = make([]byte, freader.filesize)
			logger.Debugf(" Reading whole file %s ", freader.filename)
			file.Read(data)
			return data
		} else {
			logger.Errorf("Invalid Seek options to read file [%d,%d] ", freader.start, freader.end)
		}

		file.Close()

	} else {
		logger.Error("Unable to open file: ", freader.filename)
		panic("Error opening file...Exit initiated! ")
	} //end of file open check

	logger.Warningf("[%s] NO Bytes to return ", freader.filename)
	return make([]byte, 0)

} //end of method

func getFileSize(file *os.File) uint64 {
	stat, err := file.Stat()
	if check(err) {
		return uint64(stat.Size())
	} else {
		logger.Alert("Unable to get filesize")
		logger.Critical("Closing application...")
		panic("Need filesize to distribute load to workers")
	}
}
