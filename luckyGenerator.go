package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	routinesNumber      = 10
	luckyNumberMaxRange = math.MaxInt32
)

type generateNumber func(int32) int32

func guessLuckyNumberWorker(luckyNumber int32, counterChannel chan int, wg *sync.WaitGroup, ctx context.Context, cancel context.CancelFunc, generateFunc generateNumber) {
	defer wg.Done()

	generated := generateFunc(luckyNumberMaxRange)
	counter := 1

	for generated != luckyNumber {
		select {
		case <-ctx.Done():
			counterChannel <- counter
			return
		default:
			generated = generateFunc(luckyNumberMaxRange)
			counter++
		}
	}

	counterChannel <- counter
	cancel()
}

func guessLuckyNumber(luckyNumber int32, timeout time.Duration, generateFunc generateNumber) {
	guessCounterChannel := make(chan int, routinesNumber)
	workersWg := new(sync.WaitGroup)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	for i := 0; i < routinesNumber; i++ {
		workersWg.Add(1)
		go guessLuckyNumberWorker(luckyNumber, guessCounterChannel, workersWg, ctx, cancel, generateFunc)
	}

	workersWg.Wait()
	close(guessCounterChannel)

	totalGuesses := 0
	for guess := range guessCounterChannel {
		totalGuesses += guess
	}

	if ctx.Err().Error() == "context deadline exceeded" {
		fmt.Println("Reached timeout trying generating lucky number -", luckyNumber)
	} else {
		fmt.Println("Successfully generated lucky number -", luckyNumber)
	}

	fmt.Println("Total guesses -", totalGuesses)
}

func main() {
	args := os.Args

	if len(args) != 3 {
		panic("Wrong number of arguments. Please insert the lucky number as the first argument and timeout (seconds) as the second argument")
	}

	luckyNumber, luckyNumberformatErr := strconv.ParseInt(args[1], 10, 32)
	timeoutSeconds, timeoutFormatErr := strconv.ParseInt(args[2], 10, 32)

	if luckyNumberformatErr != nil {
		panic("Could not parse lucky number")
	}

	if timeoutFormatErr != nil {
		panic("Could not parse timeout (seconds)")
	}

	fmt.Println("Trying to guess lucky number -", luckyNumber, "with timeout -", timeoutSeconds, "seconds")

	guessLuckyNumber(int32(luckyNumber), time.Duration(timeoutSeconds)*time.Second, rand.Int31n)
}
