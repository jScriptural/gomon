package logger


const (
	COLORRESET = "\033[0m"
	COLORDEBUG = "\033[3;36m"
	COLORERROR = "\033[1;31m"
	COLORINFO = "\033[1;32m"
	COLORWARN = "\033[1;33m"
	COLORGRAY = "\033[1;30m"
	COLORCYAN = "\033[1;36m"
	COLORNEUTRAL = "\033[0;37m"
)

func Colorize(color, text string) string {
	return color + text + COLORRESET;
}


