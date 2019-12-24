package main

import (
	"os"
	"log"
	"fmt"
	"bufio"
	regex "github.com/lg-melo/linebyregex"
)

/* TO REVIEW */

// Starts the application, ie:
	// Receives the command's arguments;
	// Compiles the input regex into nfa;
	// Parses each input file one line at a time.
func main(){
	if len(os.Args) < 3 {
		fmt.Println("Usage: command pattern file...")
		return
	}

	pattern, filepaths := os.Args[1], os.Args[2:]

	if len(pattern) == 0 {
		return
	}
	
	pattern = prepare(pattern)
	err, nfa := regex.Compile(pattern)

	if err != nil {
		log.Fatal(err)
	}

	for _, filepath := range filepaths {
		file, err := os.Open(filepath)
		
		if err != nil {
			log.Println(err)
			continue
		}
		defer file.Close()

		process(nfa, filepath, file)
	}
}

func prepare(pattern string) string {
	// match whole line
	if pattern[0] == '^' && pattern[len(pattern) - 1] == '$' {
		return pattern[1:len(pattern) - 1] + ")"
	}
	// match prefix
	if pattern[0] == '^' {
		return pattern[1:] + ".*)"
	}
	// match sufix
	if pattern[len(pattern) - 1] == '$' {
		return ".*" + pattern[:len(pattern) - 1] + ")"
	}

	return ".*" + pattern + ".*)"
}

// computes each of the file's lines, printing the valid ones.
func process(nfa *regex.NFA, filepath string, file *os.File) {
	scanner := bufio.NewScanner(file)
	
	for lineNum := 1; scanner.Scan(); lineNum++ {
		line := []byte(scanner.Text())

		if nfa.Accepts(line) {
			fmt.Printf("%q:%d: %q\n", filepath, lineNum, line)
		}
	}

	err := scanner.Err()
	if err != nil {
		log.Printf("error %q when parsing file %q\n", err.Error(), filepath)
	}
}