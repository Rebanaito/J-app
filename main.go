package main

import (
	"bufio"
	"fmt"
	"japp/cmdoutput"
	"japp/env"
	"japp/wordsearch"
	"os"
	"time"

	"github.com/inancgumus/screen"
)

func main() {
	screen.Clear()
	screen.MoveTopLeft()
	fmt.Println("Initializing, please wait a moment...")
	env, err := env.Initialize()
	screen.Clear()
	screen.MoveTopLeft()
	if err == nil {
		var query string
		scanner := bufio.NewScanner(os.Stdin)
		query_channel := make(chan string)
		go func() {
			for {
				time.Sleep(time.Millisecond * 200)
				fmt.Println("Write the word you would like to find or just press Enter to exit the program")
				scanner.Scan()
				query = scanner.Text()
				if query == "" {
					close(query_channel)
					break
				} else {
					query_channel <- query
				}
			}
		}()
		for query_string := range query_channel {
			result := wordsearch.SearchQuery(*env, query_string)
			screen.Clear()
			screen.MoveTopLeft()
			fmt.Printf("You searched for '%v'\n\n", query_string)
			cmdoutput.PrintResults(*env, result, query_string)
		}
	}
	screen.Clear()
	screen.MoveTopLeft()
	fmt.Println("Thank you for using my program!")
	time.Sleep(time.Second * 2)
	screen.Clear()
	screen.MoveTopLeft()
}
