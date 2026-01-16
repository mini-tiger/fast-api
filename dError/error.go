package dError

import "errors"

type errorType struct {
	SourceErr error
	UserMag   string
}

func NewError(userMag string, sourceErrList ...error) *errorType {
	var err error
	if 0 == len(sourceErrList) {
		err = errors.New("")
	} else {
		err = sourceErrList[0]
	}
	return &errorType{
		UserMag:   userMag,
		SourceErr: err,
	}
}

func (e *errorType) Error() string {
	return e.UserMag
}

func (e *errorType) GetContent() *errorType {
	return e
}
