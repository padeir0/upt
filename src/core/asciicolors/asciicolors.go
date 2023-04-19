package asciicolors

type Color = string

const (
	Bold = "\u001b[1m"

	Black   = "\u001b[30m"
	Red     = "\u001b[31m"
	Green   = "\u001b[32m"
	Yellow  = "\u001b[33m"
	Blue    = "\u001b[34m"
	Magenta = "\u001b[35m"
	Cyan    = "\u001b[36m"
	White   = "\u001b[37m"
	Reset   = "\u001b[0m"

	BackgroundBlack   = "\u001b[40m"
	BackgroundRed     = "\u001b[41m"
	BackgroundGreen   = "\u001b[42m"
	BackgroundYellow  = "\u001b[43m"
	BackgroundBlue    = "\u001b[44m"
	BackgroundMagenta = "\u001b[45m"
	BackgroundCyan    = "\u001b[46m"
	BackgroundWhite   = "\u001b[47m"
)

func MakeBlue(s string) string {
	return Blue + s + Reset
}
