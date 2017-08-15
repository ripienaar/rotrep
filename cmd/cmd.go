package cmd

import (
	"fmt"
	"os"

	"github.com/ripienaar/rotrep/filesums"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug   bool
	verbose bool
	path    string
	cmd     string
	workers int
	quiet   bool
	sums    *filesums.FileSums
)

func Run() {
	app := kingpin.New("rotrep", "Detect and Report Bit Rot")
	app.Author("R.I.Pienaar <rip@devco.net>")
	app.Flag("path", "Root path to traverse").Short('p').Required().StringVar(&path)
	app.Flag("workers", "Amount of worker routines to use for checksumming").Short('w').Default("1").IntVar(&workers)
	app.Flag("verbose", "Enable verbose logging").Short('v').Default("false").BoolVar(&verbose)
	app.Flag("debug", "Enable debug logging").Short('d').Default("false").BoolVar(&debug)

	app.Command("update", "Store new checksums and update existing ones that do not match")
	v := app.Command("verify", "Verify previously recorded checksums")
	v.Flag("quiet", "Do not produce output, only exit with 0 or 1").Short('q').Default("false").BoolVar(&quiet)

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	configureLogging()

	log.WithFields(log.Fields{"debug": debug, "verbose": verbose, "workers": workers, "quiet": quiet}).Infof("Managing checksums for path %s", path)

	var err error

	sums, err = filesums.NewFileSums(path, workers, quiet)
	if err != nil {
		log.Errorf("Could not initialize filesum tool: %s\n", err.Error())
		os.Exit(1)
	}

	switch cmd {
	case "verify":
		verify()
	case "update":
		update()
	}
}

func configureLogging() {
	log.SetLevel(log.ErrorLevel)

	if verbose {
		log.SetLevel(log.InfoLevel)
	}

	if debug {
		log.SetLevel(log.DebugLevel)
	}
}

func verify() {
	failed, err := sums.Verify()

	if !quiet || failed > 0 {
		fmt.Println("Summary:")
		fmt.Printf("\tFailed: %d\n", failed)
		if err != nil {
			fmt.Printf("Verify failed: %s\n", err.Error())
		}
	}

	if err != nil || failed > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}

func update() {
	changed, added, err := sums.Update()

	fmt.Println("Summary:")
	fmt.Printf("\t  Added: %d\n", added)
	fmt.Printf("\tChanged: %d\n", changed)

	if err != nil {
		fmt.Printf("Updating failed: %s\n", err.Error())
		os.Exit(1)
	}
}
