package util

import (
	"strings"
)

type multierror []error

type Errors struct {
	err multierror
}

func NewErrors(errs ...error) *Errors {
	ret := &Errors{}
	ret.AppendMulti(errs...)
	return ret
}

func (e *Errors) Append(err error) {
	if err == nil {
		return
	}
	e.err = append(e.err, err)
}

func (e *Errors) AppendMulti(errs ...error) {
	for _, err := range errs {
		if err != nil {
			e.err = append(e.err, err)
		}
	}
}

func (e *Errors) Error() error {
	if len(e.err) == 0 {
		return nil
	} else if len(e.err) == 1 {
		return e.err[0]
	} else {
		return e.err
	}
}

func (m multierror) Error() string {
	if len(m) == 0 {
		return ""
	} else if len(m) == 1 {
		return m[0].Error()
	} else {
		strs := make([]string, len(m))
		for i, e := range m {
			strs[i] = e.Error()
		}
		return "[{" + strings.Join(strs, "}, {") + "}]"
	}
}
