package utils

import "time"

const (
	ProgramFmt = "2006-01-02 15:04:05.000"

	NoSepFmt = "20060102150405"

	DefaultFmt = "2006-01-02 15:04:05"
)

func FmtProgramNow() string {
	return FmtNow(ProgramFmt)
}

func FmtProgram(t time.Time) string {
	return t.Format(ProgramFmt)
}

func FmtDef(t time.Time) string {
	return t.Format(DefaultFmt)
}

func FmtDefNow() string {
	return FmtNow(DefaultFmt)
}

func FmtNow(patten string) string {
	return time.Now().Format(patten)
}

func Fmt(t time.Time, patten string) string {
	return t.Format(patten)
}
