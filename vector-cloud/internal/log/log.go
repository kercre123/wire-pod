package log

var logVicos func(a ...interface{}) (int, error)

func IfVicos(a ...interface{}) (int, error) {
	if logVicos == nil {
		return 0, nil
	}
	return logVicos(a...)
}
