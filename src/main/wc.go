package main

import "os"
import "fmt"
import "mapreduce"
import (
	"container/list"
	"unicode"
	"strings"
	"strconv"
	"log"
)

// our simplified version of MapReduce does not supply a
// key to the Map function, as in the paper; only a value,
// which is a part of the input file contents
func Map(value string) *list.List {
	// Note: The value argument holds one line of text from the file.
	// You need to:
	// (1) Split up the string into words, discarding any punctuation
	// (2) Add each word to the list with a mapreduce.KeyValue struct
	separator := func(r rune) bool {
		return !unicode.IsLetter(r)
	}

	words := strings.FieldsFunc(value, separator)

	m := make(map[string]int)
	for _, word := range words {
		if count, ok := m[word]; ok {
			m[word] = count + 1
		} else {
			m[word] = 1
		}
	}

	res := list.New()
	for word, count := range m {
		res.PushBack(mapreduce.KeyValue{word, strconv.Itoa(count)})
	}

	return res
}

// iterate over list and add values
func Reduce(key string, values *list.List) string {
	// Note:
	// The key argument holds the key common too all values in the values argument
	// The values argument is a list of value strings with the given key.
	// You need to:
	// (1) Reduce the all of the values in the values list
	// (2) Return the reduced/summed up values as a string

	sum := 0
	for kv := values.Front(); kv != nil ; kv = kv.Next() {
		if count, err := strconv.Atoi(kv.Value.(string)); err != nil {
			log.Fatal(err)
		} else {
			sum += count
		}
	}

	return strconv.Itoa(sum)
}

// Can be run in 3 ways:
// 1) Sequential (e.g., go run wc.go master x.txt sequential)
// 2) Master (e.g., go run wc.go master x.txt localhost:7777)
// 3) Worker (e.g., go run wc.go worker localhost:7777 localhost:7778 &)
func main() {
	if len(os.Args) != 4 {
		fmt.Printf("%s: see usage comments in file\n", os.Args[0])
	} else if os.Args[1] == "master" {
		if os.Args[3] == "sequential" {
			mapreduce.RunSingle(5, 3, os.Args[2], Map, Reduce)
		} else {
			mr := mapreduce.MakeMapReduce(5, 3, os.Args[2], os.Args[3])
			// Wait until MR is done
			<-mr.DoneChannel
		}
	} else {
		mapreduce.RunWorker(os.Args[2], os.Args[3], Map, Reduce, 100)
	}
}
