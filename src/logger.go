package main

import (
	"log"
	"strings"

	"github.com/ttacon/chalk"
)

// LogBlackOnWhite logs with black text on the white background
func LogBlackOnWhite(s ...string) {
	blackOnWhite := chalk.Black.NewStyle().WithBackground(chalk.White)
	join := strings.Join(s, " ")
	log.Printf("%s%s%s\n", blackOnWhite, join, chalk.Reset)
}

// LogColor logs text in the given color:
// Black, Red, Green, Yellow, Blue, Magenta, Cyan, White
func LogColor(c chalk.Color, s ...string) {
	join := strings.Join(s, " ")
	log.Println(c.Color(join))
}

// LogRed logs text in red.
func LogRed(s ...string) {
	LogColor(chalk.Red, s...)
}

// LogGreen logs text in green.
func LogGreen(s ...string) {
	LogColor(chalk.Green, s...)
}

// LogYellow logs text in yellow.
func LogYellow(s ...string) {
	LogColor(chalk.Yellow, s...)
}

// LogCyan logs text in cyan.
func LogCyan(s ...string) {
	LogColor(chalk.Cyan, s...)
}

// LogMagenta logs text in magenta.
func LogMagenta(s ...string) {
	LogColor(chalk.Magenta, s...)
}
