package main

import (
	"flag"
	"github.com/ideade/epic-notifier/app"
)

func main() {
	params := new(app.Parameters)

	flag.BoolVar(&params.PostCurrent, "c", true, "Specify to not post current games.")
	flag.BoolVar(&params.Silent, "s", false, "Specify to post games silently.")
	flag.BoolVar(&params.Next, "next", false, "Create new post with games of the next giveaway.")
	flag.BoolVar(&params.Remind, "remind", false, "Resend remind post to the channel.")
	flag.StringVar(&params.TestChannel, "test", "", "Post to the test channel.")
	flag.StringVar(&params.ConfigFile, "config", "config.json", "Config filename")
	flag.Parse()

	app.Prepare(*params).Loop()
}
