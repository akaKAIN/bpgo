package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	search, err := InitSearch("../../../.")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	go func() {
		select {
		case <-ctx.Done():
			search.IncreaseDeep()
		}
	}()

	if err = search.Start(); err != nil {
		log.Println(err)
		os.Exit(1)
	}

	result := search.Find("js")
	fmt.Println("LEN: ", len(result))

	for index, fi := range result {
		fmt.Println("RESULT: ", index, fi.Path())
	}
}
