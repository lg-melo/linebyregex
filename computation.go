package linebyregex

import (
	"container/list"
)

func (nfa *NFA) Accepts(line []byte) bool{
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
			delete(temp, newState)
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
