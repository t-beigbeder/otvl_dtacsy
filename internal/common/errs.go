package common

type errorConst string

const ErrUnhandledFileType errorConst = "unhandled file type"
const ErrReadClosedQueue errorConst = "all is read on closed queue"

func (e errorConst) Error() string {
	return string(e)
}
