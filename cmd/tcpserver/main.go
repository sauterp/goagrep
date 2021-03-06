package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli"
	"github.com/firstrow/tcp_server"
	"github.com/schollz/goagrep/goagrep"
)

var VersionNum string
var Build string
var BuildTime string

func main() {
	if len(Build) > 6 {
		Build = Build[0:6]
	}
	app := cli.NewApp()
	app.Name = "goagrep"
	app.Usage = "Fuzzy matching of big strings.\n   Before use, make sure to make a data file (goagrep build)."
	app.Version = VersionNum + " (" + Build + ")"
	var wordlist, subsetSize, outputFile, searchWord string
	var verbose, listAll bool
	var tcpServer bool
	var port string
	goagrep.Normalize = true
	app.Commands = []cli.Command{
		{
			Name:    "match",
			Aliases: []string{"m"},
			Usage:   "fuzzy match word",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "database, d",
					Usage:       "input database name (built using 'goagrep build')",
					Destination: &wordlist,
				},
				cli.StringFlag{
					Name:        "word, w",
					Usage:       "word to use",
					Destination: &searchWord,
				},
				cli.BoolFlag{
					Name:        "all, a",
					Usage:       "list all matches",
					Destination: &listAll,
				},
			},
			Action: func(c *cli.Context) error {
				if len(wordlist) == 0 || len(searchWord) == 0 {
					cli.ShowCommandHelp(c, "match")
				} else {
					if listAll {
						words, scores, err := goagrep.GetMatches(strings.ToLower(searchWord), wordlist)
						if err != nil {
							fmt.Printf("Not found|||-1")
							fmt.Println(err)
							fmt.Println(words, scores)
						} else {
							for i, word := range words {
								fmt.Printf("%v|||%v\n", word, scores[i])
							}
						}
					} else {
						word, score, err := goagrep.GetMatch(strings.ToLower(searchWord), wordlist)
						if err != nil {
							fmt.Printf("Not found|||-1")
						} else {
							fmt.Printf("%v|||%v", word, score)
						}
					}
				}
				return nil
			},
		},
		{
			Name:    "build",
			Aliases: []string{"b"},
			Usage:   "builds the database subsequent fuzzy matching",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "list, l",
					Usage:       "wordlist to use, seperated by newlines",
					Destination: &wordlist,
				},
				cli.StringFlag{
					Name:        "database, d",
					Usage:       "output database name (default: words.db)",
					Destination: &outputFile,
				},
				cli.StringFlag{
					Name:        "size, s",
					Usage:       "subset size (default: 3)",
					Destination: &subsetSize,
				},
				cli.BoolFlag{
					Name:        "verbose, v",
					Usage:       "show more output",
					Destination: &verbose,
				},
			},
			Action: func(c *cli.Context) error {
				if len(subsetSize) == 0 {
					subsetSize = "3"
				}
				if len(outputFile) == 0 {
					outputFile = "words.db"
				}
				if len(wordlist) == 0 {
					cli.ShowCommandHelp(c, "build")
				} else {
					fmt.Println("Generating '" + outputFile + "' from '" + wordlist + "' with subset size " + subsetSize)
					tupleLength, _ := strconv.Atoi(subsetSize)
					goagrep.GenerateDB(wordlist, outputFile, tupleLength, verbose)
					fmt.Println("Finished building db")
				}
				return nil
			},
		},
		{
			Name:    "serve",
			Aliases: []string{"s"},
			Usage:   "serve database on TCP for subsequent fuzzy matching",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "list, l",
					Usage:       "wordlist to use, seperated by newlines",
					Destination: &wordlist,
				},
				cli.StringFlag{
					Name:        "size, s",
					Usage:       "subset size (default: 3)",
					Destination: &subsetSize,
				},
				cli.StringFlag{
					Name:        "port, p",
					Usage:       "port to use (default: 3334)",
					Destination: &port,
				},
				cli.BoolFlag{
					Name:        "verbose, v",
					Usage:       "show more output",
					Destination: &verbose,
				},
			},
			Action: func(c *cli.Context) error {
				if len(subsetSize) == 0 {
					subsetSize = "3"
				}
				if len(port) == 0 {
					port = "3334"
				}
				if len(wordlist) == 0 {
					cli.ShowCommandHelp(c, "serve")
				} else {
					tcpServer = true
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
	if !tcpServer {
		os.Exit(0)
	}

	tupleLength, _ := strconv.Atoi(subsetSize)
	words, tuples := goagrep.GenerateDBInMemory(wordlist, tupleLength, verbose)
	start := time.Now()

	server := tcp_server.New("localhost:9992")
	server.OnNewMessage(func(c *tcp_server.Client, message string) {
		if verbose {
			start = time.Now()
		}
		matches, scores, _ := goagrep.GetMatchesInMemoryInParallel(message, words, tuples, tupleLength, true)
		if verbose {
			elapsed := time.Since(start)
			log.Printf("Searched for '%s' in %s", strings.TrimSpace(message), elapsed)
			for i := range matches {
				fmt.Printf("%d: '%s'\n", scores[i], matches[i])
				if i > 10 {
					break
				}
			}
		}
		c.Send(matches[0])
	})
	server.Listen()
}
