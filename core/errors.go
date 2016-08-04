package core

const (
	readCmdError   = 1
	decodeCmdError = 2
	parseCmdError  = 3
)

type cmdError int

func (this cmdError) Error() string {
	switch this {
	case readCmdError:
		return "Read Data From Client Error"
	case decodeCmdError:
		return "Decode Cmd Error"
	case parseCmdError:
		return "Parse Cmd Error"
	default:
		return "Unknown Error"
	}
}

func newCmdError(errorCode int) error {
	var err error
	err = cmdError(errorCode)
	return err
}
