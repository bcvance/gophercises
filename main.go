package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	csvFilename := "problems.csv"
	// readern for stdin for user input of time limit
	stdReader := bufio.NewReader(os.Stdin)
	fmt.Print("Please enter time limit (in seconds): ")
	timeLimit, _ := stdReader.ReadString('\n')
	timeLimit = strings.Replace(timeLimit, "\n", "", -1)
	limitInt, err := strconv.Atoi(timeLimit)
	var wg sync.WaitGroup
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Open("problems.csv")
	if err != nil {
		log.Fatal(err)
	}
	// count number of questions (lines) in csv (not working, 1 too low)
	problems, err := lineCounter(f)
	if err != nil {
		log.Fatal(err)
	}
	correct := 0
	total := problems
	// channel to hold stream of csv rows
	var quizCh = make(chan []string, 100)

	// only add 1 to wait group because either timer goroutine will complete (time is up)
	// OR quizRead goroutine will complete (all questions were finished). NEVER both (i think).
	wg.Add(1)
	go timer(limitInt, &wg)
	go quizWrite(csvFilename, quizCh)
	go quizRead(quizCh, &correct, &wg)

	// this wait means that main function won't exit until timer or quizRead goroutine has finished
	wg.Wait()
	fmt.Printf("\n%d correct of %d\n", correct, total)
}

func timer(seconds int, wg *sync.WaitGroup) {
	time.Sleep(time.Duration(seconds) * time.Second)
	wg.Done()
}

func quizWrite(csvFilename string, ch chan<- []string) {
	f, err := os.Open("problems.csv")
	if err != nil {
		log.Fatal(err)
	}
	reader := csv.NewReader(f)
	for {
		row, err := reader.Read()
		// if we reach end of file, close channel so for loop is quizRead will exit
		if err == io.EOF {
			close(ch)
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// if no errors, push csv row to channel
		ch <- row
	}
}

func quizRead(ch <-chan []string, correct *int, wg *sync.WaitGroup) {
	stdReader := bufio.NewReader(os.Stdin)
	for row := range ch {
		// for each row, ask user for answer and check for correctness, then increment number of correct answers accordingly
		fmt.Printf("%s=", row[0])
		answer, _ := stdReader.ReadString('\n')
		answer = strings.Replace(answer, "\n", "", -1)

		if strings.Compare(row[1], answer) == 0 {
			*correct++
		}
	}
	// once all csv rows have been read and handled, exit and indicate to WaitGroup that goroutine has finished
	wg.Done()
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}
