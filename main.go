package main

import (
	"fmt"
	"github.com/romanoff/fsmonitor"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"time"
)

/*
Example gg.yaml

```yaml
watch:

- pattern: "*.txt"
  command: "echo hello world, txt"

- pattern: "*.go"
  command: "echo hello world, go"
```
*/

type Config struct {
	Watch []struct {
		Pattern string
		Command string
	}
}

func main() {
	commandTriggerDelay := 250 * time.Millisecond
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to get current directory. Wtf?")
		os.Exit(1)
	}

	if _, err := os.Stat("gg.yaml"); err != nil {
		fmt.Fprintln(os.Stderr, "Please create gg.yaml.")
		os.Exit(1)
	}

	f, err := ioutil.ReadFile("gg.yaml")
	if err != nil {
		panic(err)
	}

	c := Config{}
	if err := yaml.Unmarshal(f, &c); err != nil {
		panic(err)
	}

	watcher, err := fsmonitor.NewWatcherWithSkipFolders([]string{".git"})
	if err != nil {
		panic(err)
	}

	err = watcher.Watch(currentDir)

	commandTriggerDelays := make(map[string]time.Time)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for _ = range ch {
			fmt.Println("\nAuf Wiederschaun!")
			os.Exit(0)
		}
	}()

	for {
		select {
		case ev := <-watcher.Event:
			for _, w := range c.Watch {
				if ev.IsModify() {

					// http://golang.org/pkg/path/#Match
					basename := path.Base(ev.Name)
					match, err := path.Match(w.Pattern, basename)
					if err != nil {
						log.Printf("[error] [%v] %v for pattern `%v`", basename, err, w.Pattern)
					}

					if match {
						if _, ok := commandTriggerDelays[ev.Name]; !ok {
							commandTriggerDelays[ev.Name] = time.Now()
						}

						if commandTriggerDelays[ev.Name].Add(commandTriggerDelay).Before(time.Now()) {
							// log.Printf("Run %v", ev.Name)
							cmd := exec.Command("sh", "-c", w.Command)
							cmd.Stdin = os.Stdin
							cmd.Stdout = os.Stdout
							cmd.Stderr = os.Stderr
							if err := cmd.Run(); err != nil {
								log.Printf("[error] [%v] %v", basename, err)
							}
							fmt.Printf("\n")
							commandTriggerDelays[ev.Name] = time.Now()
						}
					}
				}
			}
		case err := <-watcher.Error:
			log.Printf("[error] %v", err)
		}
	}

}
