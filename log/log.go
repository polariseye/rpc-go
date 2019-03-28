package log

import "fmt"

type Logger interface {
	Debug(fmtString string, paramList ...interface{})

	Info(fmtString string, paramList ...interface{})

	Warn(fmtString string, paramList ...interface{})

	Error(fmtString string, paramList ...interface{})

	Fatal(fmtString string, paramList ...interface{})
}

type DefaultLog struct {
}

func (this *DefaultLog) Debug(fmtString string, paramList ...interface{}) {
	if len(paramList) > 0 {
		fmt.Println(fmt.Sprintf(fmtString, paramList...))
	} else {
		fmt.Println(fmtString)
	}
}

func (this *DefaultLog) Info(fmtString string, paramList ...interface{}) {
	if len(paramList) > 0 {
		fmt.Println(fmt.Sprintf(fmtString, paramList...))
	} else {
		fmt.Println(fmtString)
	}
}

func (this *DefaultLog) Warn(fmtString string, paramList ...interface{}) {
	if len(paramList) > 0 {
		fmt.Println(fmt.Sprintf(fmtString, paramList...))
	} else {
		fmt.Println(fmtString)
	}
}

func (this *DefaultLog) Error(fmtString string, paramList ...interface{}) {
	if len(paramList) > 0 {
		fmt.Println(fmt.Sprintf(fmtString, paramList...))
	} else {
		fmt.Println(fmtString)
	}
}

func (this *DefaultLog) Fatal(fmtString string, paramList ...interface{}) {
	if len(paramList) > 0 {
		fmt.Println(fmt.Sprintf(fmtString, paramList...))
	} else {
		fmt.Println(fmtString)
	}
}
