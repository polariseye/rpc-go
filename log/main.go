package log

var logOBj Logger = new(DefaultLog)

func SetLogger(tmpLogObj Logger) {
	logOBj = tmpLogObj
}

func Debug(fmtString string, paramList ...interface{}) {
	logOBj.Debug(fmtString, paramList...)
}

func Info(fmtString string, paramList ...interface{}) {
	logOBj.Info(fmtString, paramList...)

}

func Warn(fmtString string, paramList ...interface{}) {
	logOBj.Warn(fmtString, paramList...)

}

func Error(fmtString string, paramList ...interface{}) {
	logOBj.Error(fmtString, paramList...)

}

func Fatal(fmtString string, paramList ...interface{}) {
	logOBj.Fatal(fmtString, paramList...)
}

/*
func DebugObj(info string, data ...interface{}) {

}

func InfoObj(info string, data ...interface{}) {

}

func WarnObj(info string, data ...interface{}) {

}

func ErrorObj(info string, data ...interface{}) {

}

func FatalObj(info string, data ...interface{}) {

}
*/
