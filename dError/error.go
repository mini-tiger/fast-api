package dError

type ErrorType struct {
	SourceErr []error
	UserMag   string
}

func NewError(userMag string, sourceErrList ...error) *ErrorType {
	//var err error
	//if 0 == len(sourceErrList) {
	//	err = errors.New("")
	//} else {
	//	err = sourceErrList[0]
	//}
	return &ErrorType{
		UserMag:   userMag,
		SourceErr: sourceErrList,
	}
}

func (e *ErrorType) Error() string {
	return e.UserMag
}

func (e *ErrorType) GetContent() *ErrorType {
	return e
}
