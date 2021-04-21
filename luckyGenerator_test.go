package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

const luckyNumberExample = 5

func generateLucky(luckyNumber int32) int32 {
	return luckyNumberExample
}

func TestGuessLuckyNumberWorkerSanity(t *testing.T) {
	guessCounterChannel := make(chan int, routinesNumber)
	workersWg := new(sync.WaitGroup)
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	workersWg.Add(1)
	go guessLuckyNumberWorker(luckyNumberExample, guessCounterChannel, workersWg, ctx, cancel, generateLucky)

	workersWg.Wait()
	close(guessCounterChannel)

	totalGuesses := <-guessCounterChannel

	if totalGuesses != 1 {
		t.Errorf("Wrong number of guesses")
	}
}

var secondWorkerActualGuessCount int = 0

func neverGenerateLuckyWithCount(luckyNumber int32) int32 {
	secondWorkerActualGuessCount++
	return luckyNumberExample + 1
}

func TestGuessLuckyNumberOneWorkerSucceedSecondNot(t *testing.T) {

	generateLuckyWithSleep := func(luckyNumber int32) int32 {
		time.Sleep(time.Millisecond * 1)
		return luckyNumberExample
	}

	guessCounterChannel := make(chan int, routinesNumber)
	workersWg := new(sync.WaitGroup)
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	workersWg.Add(2)
	go guessLuckyNumberWorker(luckyNumberExample, guessCounterChannel, workersWg, ctx, cancel, generateLuckyWithSleep)
	go guessLuckyNumberWorker(luckyNumberExample, guessCounterChannel, workersWg, ctx, cancel, neverGenerateLuckyWithCount)

	workersWg.Wait()
	close(guessCounterChannel)

	workerOneGuessCount, workerTwoGuessCount := <-guessCounterChannel, <-guessCounterChannel

	if workerOneGuessCount+workerTwoGuessCount != secondWorkerActualGuessCount+1 {
		t.Errorf("Wrong number of guesses")
	}
}

func neverGenerateLucky(luckyNumber int32) int32 {
	return luckyNumberExample + 1
}

func TestGuessLuckyNumberWorkerCanceled(t *testing.T) {
	guessCounterChannel := make(chan int, routinesNumber)
	workersWg := new(sync.WaitGroup)
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	workersWg.Add(1)
	go guessLuckyNumberWorker(luckyNumberExample, guessCounterChannel, workersWg, ctx, cancel, neverGenerateLucky)

	time.Sleep(time.Millisecond * 1)
	cancel()

	workersWg.Wait()
	close(guessCounterChannel)

	guessCount := <-guessCounterChannel

	if guessCount == 0 {
		t.Errorf("Worker didn't try generating number until canceled")
	}
}

var totalGuessCountSanity int = 0

func generateLuckyWithCount(luckyNumber int32) int32 {
	if totalGuessCountSanity != 0 {
		time.Sleep(time.Millisecond * 10)
	}

	totalGuessCountSanity++

	return luckyNumberExample
}

func TestGuessLuckyNumberSanity(t *testing.T) {
	guessLuckyNumber(luckyNumberExample, 10, generateLuckyWithCount)

	if totalGuessCountSanity > 10 {
		t.Errorf("Generated too many numbers")
	}
}

var totalGuessCountTimeout int = 0

func generateLuckyForTimeout(luckyNumber int32) int32 {
	totalGuessCountTimeout++
	time.Sleep(time.Second * 2)
	return luckyNumber + 1
}

func TestGuessLuckyNumberTimeout(t *testing.T) {

	guessLuckyNumber(luckyNumberExample, 1, generateLuckyForTimeout)

	if totalGuessCountTimeout > 10 {
		t.Errorf("workers didn't finished after timeout")
	}
}
