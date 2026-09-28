package check

// the parts of a check boot the tests drive one by one
var (
	Credentials = credentials
	Booted      = booted
	Run         = run
)

type (
	Notices = noticeListener
	Conn    = shellConn
)
