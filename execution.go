package regex

import (
	"os"
	"log"
	"fmt"
	"bufio"
	"container/list"
)

// Starts the application, that is:
	// Receives the command's arguments;
	// Compiles the input regex into nfa;
	// Parses each input file one line at a time.
func Run(){
	if len(os.Args) < 3 {
		fmt.Println("Usage: command pattern file...")
		return
	}

	pattern, filenames := os.Args[1], os.Args[2:]

	if len(pattern) == 0 {
		return
	}
	
	pattern = preparePattern(pattern)
	err, nfa := compile(pattern)

	if err != nil {
		log.Fatal(err)
	}

	for _, filename := range filenames {
		file, err := os.Open(filename)
		
		if err != nil {
			log.Println(err)
			continue
		}
		defer file.Close()

		nfa.processFile(filename, file)
	}
}

// computes each of the file's lines, printing the valid ones.
func (nfa *NFA) processFile(filename string, file *os.File) {
	scanner := bufio.NewScanner(file)
	
	for i := 1; scanner.Scan(); i++ {
		line := []byte(scanner.Text())

		if nfa.accepts(line) {
			fmt.Printf("file %q, line %d: %q\n", filename, i, line)
		}
	}

	err := scanner.Err()
	if err != nil {
		log.Printf("error %q during parsing of file %s\n", err.Error(), filename)
	}
}

func (nfa *NFA) accepts(line []byte) bool{
	if nfa.initial == nil {
		return false
	}

	currentStates := make(map[*State]bool)
	nfa.initial.reaches(currentStates)

	// process line
	temp := make(map[*State]bool)
	for _, c := range line {
		if len(currentStates) == 0 {
			break
		}

		for state := range currentStates {
			nextStates := state.transition[c]
			
			if nextStates == nil {
				continue
			}

			for nextState := range nextStates {
				nextState.reaches(temp)
			}
		}

		// updates the current set of states
		for oldState := range currentStates {
			if !temp[oldState] {
				delete(currentStates, oldState)
			}
		}
		for newState := range temp {
			currentStates[newState] = true
		}
		for k := range temp {
			delete(temp, k)
		}
	}

	return currentStates[nfa.final]
}

// puts in 'result' all states reachable from s through epsilon transitions
func (s *State) reaches(result map[*State]bool) {
	push := func(l *list.List, s *State){
		l.PushBack(s);
	}
	pop := func(l *list.List) *State{
		element := l.Front()
		l.Remove(element)
		return element.Value.(*State)
	}

	visited := make(map[*State]bool)
	queue := list.New()
	
	push(queue, s)
	visited[s] = true
	for queue.Len() > 0 {
		state := pop(queue)
		result[state] = true

		for nextState := range state.epsTransition {
			if !visited[nextState] {
				push(queue, nextState)
				visited[nextState] = true
			}
		}
	}
}
