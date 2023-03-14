package cmd

import (
	"bufio"
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vitorfhc/spammer/pkg/spammer"
)

type CliOptions struct {
	Wordlist  string
	Hostsfile string
	Threads   uint
	RateLimit uint
	Debug     bool
	Silent    bool
}

var cliOptions *CliOptions

var rootCmd = &cobra.Command{
	Use:   "spammer",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		if cliOptions.Debug {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if cliOptions.Silent {
			logrus.SetLevel(logrus.PanicLevel)
		}

		internalCtx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		c := make(chan os.Signal, 1)
		defer close(c)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		go func() {
			logrus.Debug("starting interrupt handler")
			s := <-c
			if s == nil {
				return
			}
			cancel()
			logrus.Errorf("interrupt received: waiting 5 seconds before forcing exit")
			time.Sleep(5 * time.Second)
			logrus.Error("forcing exit")
			os.Exit(1)
		}()

		paths := []string{}
		hosts := []string{}

		hostsFile, err := os.OpenFile(cliOptions.Hostsfile, os.O_RDONLY, 0)
		if err != nil {
			logrus.Errorf("error opening hosts file: %s", err)
			os.Exit(1)
		}
		defer hostsFile.Close()
		hostsScanner := bufio.NewScanner(hostsFile)

		wordlistFile, err := os.OpenFile(cliOptions.Wordlist, os.O_RDONLY, 0)
		if err != nil {
			logrus.Errorf("error opening wordlist file: %s", err)
			os.Exit(1)
		}
		defer wordlistFile.Close()
		wordlistScanner := bufio.NewScanner(wordlistFile)

		wg := &sync.WaitGroup{}

		wg.Add(1)
		go func() {
			defer wg.Done()
			logrus.Debugf("reading hosts from %q", cliOptions.Hostsfile)
			for {
				select {
				case <-internalCtx.Done():
					return
				default:
					scan := hostsScanner.Scan()
					if !scan {
						logrus.Debugf("loaded %d hosts", len(hosts))
						return
					}
					hosts = append(hosts, hostsScanner.Text())
				}
			}
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			logrus.Debugf("reading paths from %q", cliOptions.Wordlist)
			for {
				select {
				case <-internalCtx.Done():
					return
				default:
					scan := wordlistScanner.Scan()
					if !scan {
						logrus.Debugf("loaded %d paths", len(paths))
						return
					}
					paths = append(paths, wordlistScanner.Text())
				}
			}
		}()

		wg.Wait()

		spamOptions := &spammer.SpamOptions{
			Paths:   paths,
			Hosts:   hosts,
			Threads: cliOptions.Threads,
			Rate:    cliOptions.RateLimit,
		}

		err = spammer.Spam(internalCtx, spamOptions)
		if err != nil {
			logrus.Errorf("error spamming: %s", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cliOptions = &CliOptions{}
	rootCmd.Flags().StringVarP(&cliOptions.Wordlist, "wordlist", "w", "wordlist.txt", "Wordlist containing paths")
	rootCmd.Flags().StringVarP(&cliOptions.Hostsfile, "hostsfile", "f", "hosts.txt", "File containing hosts")
	rootCmd.Flags().UintVarP(&cliOptions.Threads, "threads", "t", 10, "Number of threads")
	rootCmd.Flags().UintVarP(&cliOptions.RateLimit, "rate-limit", "r", 100, "Number of requests per second")
	rootCmd.Flags().BoolVarP(&cliOptions.Debug, "debug", "d", false, "Enable debug mode")
	rootCmd.Flags().BoolVarP(&cliOptions.Silent, "silent", "s", false, "Disable error messages")
}
