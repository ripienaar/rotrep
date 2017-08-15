package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/Songmu/prompter"
	"github.com/ripienaar/rotrep/filesums"
	log "github.com/sirupsen/logrus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug    bool
	verbose  bool
	yes      bool
	quiet    bool
	progress bool
	path     string
	cmd      string
	workers  = runtime.NumCPU()
)

func Run() {
	app := kingpin.New("rotrep", "Detect and Report Bit Rot")
	app.Author("R.I.Pienaar <rip@devco.net>")
	app.Flag("path", "Root path to traverse").Short('p').Required().StringVar(&path)
	app.Flag("workers", "Amount of worker routines to use for checksumming").Short('w').IntVar(&workers)
	app.Flag("verbose", "Enable verbose logging").Short('v').Default("false").BoolVar(&verbose)
	app.Flag("debug", "Enable debug logging").Short('d').Default("false").BoolVar(&debug)
	app.Flag("progress", "Show a progress bar and summary").Default("false").BoolVar(&progress)

	u := app.Command("update", "Store new checksums and update existing ones that do not match")
	u.Flag("yes", "Assume yes to any questions").Short('y').Default("false").BoolVar(&yes)

	v := app.Command("verify", "Verify previously recorded checksums")
	v.Flag("quiet", "Do not produce output, only exit with 0 or 1").Short('q').Default("false").BoolVar(&quiet)

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	configureLogging()

	if verbose || debug || quiet {
		progress = false
	}

	log.WithFields(log.Fields{"debug": debug, "verbose": verbose, "workers": workers, "quiet": quiet, "yes": yes, "progress": progress}).Infof("Managing checksums for path %s", path)

	switch cmd {
	case "verify":
		verify()
	case "update":
		update()
	}
}

func createFileSums() *filesums.FileSums {
	sums, err := filesums.NewFileSums(path, workers, quiet, progress)
	if err != nil {
		log.Errorf("Could not initialize filesum tool: %s\n", err.Error())
		os.Exit(1)
	}

	return sums
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
	sums := createFileSums()

	err := sums.Verify()

	if !quiet {
		if progress {
			if sums.Stats.Failed() > 0 {
				fmt.Println("")

				for _, f := range sums.Stats.FailedFiles() {
					fmt.Printf("failed: %s\n", *f)
				}
			}

			fmt.Println("")
		} else if sums.Stats.Failed() > 0 {
			fmt.Println("")
		}

		fmt.Println("Summary:")
		fmt.Printf("    Root Directory: %s\n", sums.Root)
		fmt.Printf("   Sub Directories: %d\n", sums.Stats.Directories())
		fmt.Printf("    Files Verified: %d\n", sums.Stats.Verified())
		fmt.Printf("      Files Failed: %d\n", sums.Stats.Failed())
		if err != nil {
			fmt.Println("")
			fmt.Printf("Verify failed: %s\n", err.Error())
		}
	}

	if err != nil || sums.Stats.Failed() > 0 {
		os.Exit(1)
	}

	os.Exit(0)
}

func update() {
	if !yes {
		if !prompter.YN("Are you sure you wish to update checksums and add new files", false) {
			os.Exit(1)
		}
	}

	sums := createFileSums()
	err := sums.Update()

	if !quiet {
		if progress {
			if sums.Stats.Updated() > 0 || sums.Stats.New() > 0 {
				fmt.Println("")

				for _, f := range sums.Stats.UpdatedFiles() {
					fmt.Printf("updated: %s\n", *f)
				}

				for _, f := range sums.Stats.NewFiles() {
					fmt.Printf("new: %s\n", *f)
				}
			}

			fmt.Println("")
		} else if sums.Stats.Updated() > 0 || sums.Stats.New() > 0 {
			fmt.Println("")
		}

		fmt.Println("Summary:")
		fmt.Printf("    Root Directory: %s\n", sums.Root)
		fmt.Printf("   Sub Directories: %d\n", sums.Stats.Directories())
		fmt.Printf("    Files Verified: %d\n", sums.Stats.Verified())
		fmt.Printf("       Files Added: %d\n", sums.Stats.New())
		fmt.Printf("     Files Changed: %d\n", sums.Stats.Updated())
	}

	if err != nil {
		fmt.Printf("Updating failed: %s\n", err.Error())
		os.Exit(1)
	}

	if sums.Stats.Updated() > 0 || sums.Stats.New() > 0 {
		os.Exit(1)
	}
}
