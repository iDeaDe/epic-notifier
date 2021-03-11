package app

import (
	"flag"
	"log"
	"reflect"
)

type BotArgs struct {
	PostCurrent     bool
	PostSilently    bool
	RecreateNext    bool
	TestChannelName string
}

func GetArgs() BotArgs {
	var bArgs BotArgs

	flag.BoolVar(&bArgs.PostCurrent, "c", true, "Specify to not post current games.")
	flag.BoolVar(&bArgs.PostSilently, "s", false, "Specify to post games silently.")
	flag.BoolVar(&bArgs.RecreateNext, "next", false, "Create new post with games of the next giveaway.")
	flag.StringVar(&bArgs.TestChannelName, "test", "", "Post to the test channel.")
	flag.Parse()

	log.Println("Flags parsed.")
	v := reflect.ValueOf(bArgs)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {
		log.Printf("%s is set to %v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
	}

	return bArgs
}
