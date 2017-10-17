package main

import (
	"os"
	"fmt"
	"time"

	"github.com/cv/t/codes"
	"strings"
)

const layout = "15:04:05"

var clocksLow = []string{"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚",
	"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚"}
var clocksHigh = []string{"🕧", "🕜", "🕝", "🕞", "🕟", "🕠", "🕡", "🕢", "🕣", "🕤", "🕥", "🕦",
	"🕧", "🕜", "🕝", "🕞", "🕟", "🕠", "🕡", "🕢", "🕣", "🕤", "🕥", "🕦",}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "usage: t <IATA>...\n")
		os.Exit(-1)
	}

	for _, i := range os.Args[1:] {
		show(i)
	}
}

func show(iata string) {
	iata = strings.ToUpper(iata)

	locName, found := codes.IATA[iata]
	if !found {
		fmt.Printf("%s: ??:??:?? (Unknown)\n", iata)
		return
	}
	loc, _ := time.LoadLocation(locName)
	now := time.Now().In(loc)

	var emoji string
	if now.Minute() > 30 {
		emoji = clocksHigh[now.Hour()]
	} else {
		emoji = clocksLow[now.Hour()]
	}

	fmt.Printf("%s: %s  %s (%s)\n", iata, emoji, now.Format(layout), locName)
}
