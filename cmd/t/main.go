package main

import (
	"fmt"
	"os"
	"time"

	"strings"

	"github.com/cv/t/codes"
)

const layout = "15:04:05"
const layoutShort = "15:04"

var clocksLow = []string{
	"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚",
	"🕛", "🕐", "🕑", "🕒", "🕓", "🕔", "🕕", "🕖", "🕗", "🕘", "🕙", "🕚",
}

var clocksHigh = []string{
	"🕧", "🕜", "🕝", "🕞", "🕟", "🕠", "🕡", "🕢", "🕣", "🕤", "🕥", "🕦",
	"🕧", "🕜", "🕝", "🕞", "🕟", "🕠", "🕡", "🕢", "🕣", "🕤", "🕥", "🕦",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "usage: t <IATA>...\n")
		os.Exit(-1)
	}

	for i, s := range os.Args[1:] {
		show(s)
		if os.Getenv("PS1_FORMAT") != "" && i < len(os.Args[1:])-1 {
			fmt.Print(" ")
		}
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

	if os.Getenv("PS1_FORMAT") != "" {
		fmt.Printf("%s %s", iata, now.Format(layoutShort))
	} else {
		fmt.Printf("%s: %s  %s (%s)\n", iata, emoji, now.Format(layout), locName)
	}
}
