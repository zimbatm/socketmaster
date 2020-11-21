package main

import (
	"bytes"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func Test_logOutput(t *testing.T) {
	t.Run("wg waits for logOutput", func(t *testing.T) {
		var stdErr bytes.Buffer
		log.SetOutput(&stdErr)
		log.SetFlags(0)
		ioReader, ioWriter, err := os.Pipe()
		if err != nil {
			t.Fatal("unexpected error", err)
		}
		inputs := []string{"foo", "bar"}
		want := "[1] foo\n[1] bar\n"
		var waitForWrite sync.WaitGroup
		waitForWrite.Add(1)

		go func(wg *sync.WaitGroup) {
			for _, input := range inputs {
				_, err = ioWriter.WriteString(input + "\n")
				if err != nil {
					t.Fatal("unexpected error", err)
				}
				time.Sleep(5 * time.Millisecond)
			}
			err = ioWriter.Close()
			if err != nil {
				t.Fatal("unexpected error", err)
			}
			wg.Done()
		}(&waitForWrite)

		var wg sync.WaitGroup
		logOutput(ioReader, 1, &wg)

		wg.Wait()
		got := stdErr.String()
		if got != want {
			t.Fatalf("unexpected output:\n\twanted: %q\n\tgot: %q", want, got)
		}
		waitForWrite.Wait()
	})
}
