package robot

import (
	"fmt"
	"os"
)

// WriteFaceErrorCode writes the given numerical code to the robot's face with an error message
func WriteFaceErrorCode(code uint16) error {
	file, err := os.OpenFile("/run/fault_code", os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = fmt.Fprintf(file, "%d\n", code)
	return err
}
