package podonwin

import "syscall"

const (
	// needs to be able to see adrmin processes as well
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
)

func IsProcessRunning(pid int) (bool, error) {
	h, err := syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		// if the error is "The parameter is incorrect.", it usually means the process does not exist
		if err.Error() == "The parameter is incorrect." {
			return false, nil
		}
		return false, err
	}
	defer syscall.CloseHandle(h)

	var code uint32
	err = syscall.GetExitCodeProcess(h, &code)
	if err != nil {
		return false, err
	}

	// STILL_ACTIVE
	return code == 259, nil
}
