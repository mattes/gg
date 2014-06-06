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
	"strings"
	"time"
)

/*
Example gg.yaml

```yaml
watch:

- pattern: "*.txt"
  commands:
    hello: "echo hello world, txt"
	hello2: echo again, mate"
  bindkey: t @todo triggers this command when running gg

- pattern: "*.go"
  commands:
    mytest: "echo hello world, go"
  bindkey: g @todo triggers this command when running gg

- pattern: "(.*)_test.go" @todo use pattern matches in command
  command: "go run $1_test.go"
```
*/

type Config struct {
	Watch []struct {
		Pattern  string
		Commands map[string]string
	}
}

func main() {
	commandTriggerDelay := 250 * time.Millisecond
	workingDir, err := os.Getwd()
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

	err = watcher.Watch(workingDir)

	commandTriggerDelays := make(map[string]time.Time)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	go func() {
		for _ = range ch {
			fmt.Println(" Auf Wiederschaun!")
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
							for name, command := range w.Commands {
								log.Printf("Run %s on %v ...\n", name, strings.Replace(ev.Name, workingDir+"/", "", 1))
								cmd := exec.Command("sh", "-c", command)
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
			}
		case err := <-watcher.Error:
			log.Printf("[error] %v", err)
		}
	}

}
