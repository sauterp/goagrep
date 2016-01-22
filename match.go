package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/arbovm/levenshtein"
	"github.com/boltdb/bolt"
)

var matches map[string]int

func getMatch(s string, path string) (string, int) {
	// normalize
	s = strings.ToLower(s)
	// Open a new bolt database
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tupleLength := 3

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("vars"))
		v := b.Get([]byte("tupleLength"))
		tupleLength, _ = strconv.Atoi(string(v))
		return nil
	})

	partials := getPartials(s, tupleLength)
	matches := make(map[string]int)
	// var wg sync.WaitGroup

	for _, partial := range partials {
		// wg.Add(1)
		func(partial string, path string) {
			// defer wg.Done()
			db.View(func(tx *bolt.Tx) error {

				b1 := tx.Bucket([]byte("tuples-1"))
				b2 := tx.Bucket([]byte("tuples-2"))
				b3 := tx.Bucket([]byte("tuples-3"))
				b4 := tx.Bucket([]byte("tuples-4"))
				b5 := tx.Bucket([]byte("tuples-5"))
				b6 := tx.Bucket([]byte("tuples-6"))
				b7 := tx.Bucket([]byte("tuples-7"))
				b8 := tx.Bucket([]byte("tuples-8"))
				var v []byte
				if string(partial[0]) <= "c" { // DIVIDED 6x: 32MB 84 ms...UNDIVIDED: 188ms
					v = b1.Get([]byte(string(partial)))
				} else if string(partial[0]) <= "f" {
					v = b2.Get([]byte(string(partial)))
				} else if string(partial[0]) <= "i" {
					v = b3.Get([]byte(string(partial)))
				} else if string(partial[0]) <= "l" {
					v = b4.Get([]byte(string(partial)))
					log.Println(partial)
					log.Println(v)
				} else if string(partial[0]) <= "o" {
					v = b5.Get([]byte(string(partial)))
				} else if string(partial[0]) <= "r" {
					v = b6.Get([]byte(string(partial)))
				} else if string(partial[0]) <= "u" {
					v = b7.Get([]byte(string(partial)))
				} else {
					v = b8.Get([]byte(string(partial)))
				}

				vals := string(v)
				// log.Println(partial)
				// log.Printf("The answer is: %v\n", vals)
				if len(v) > 0 {
					for _, k := range strings.Split(vals, " ") {
						db.View(func(tx *bolt.Tx) error {
							b := tx.Bucket([]byte("words"))
							v := string(b.Get([]byte(k)))
							_, ok := matches[v]
							if ok != true {
								matches[v] = levenshtein.Distance(s, v)
								// fmt.Printf("Word match: %v\n", v)
								// fmt.Printf("Distance : %v\n", matches[v])
							}
							return nil
						})
					}
				}
				return nil
			})
		}(partial, path)
	}

	// wg.Wait()
	bestMatch := "none"
	bestVal := 100
	for k, v := range matches {
		if v < bestVal {
			bestMatch = k
			bestVal = v
		}

	}

	return bestMatch, bestVal
}
