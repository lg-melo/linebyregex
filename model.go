package regex

const (
	infinite int = 1 << 31 - 1
)

type State struct{
	transition map[byte]map[*State]bool
	epsTransition map[*State]bool
}

type NFA struct{
	initial *State
	endPoints map[*State]bool
	final *State
}

func newNFA(hasInitial, hasEndPoints, hasFinal bool) *NFA {
	var (
		initialState *State
		endPoints map[*State]bool
		finalState *State
	)

	if hasInitial {
		initialState = &State{
			transition: make(map[byte]map[*State]bool),
			epsTransition: make(map[*State]bool),
		}
	}

	if hasEndPoints {
		endPoints = make(map[*State]bool)
	}

	if hasFinal {
		finalState = &State{
			transition: make(map[byte]map[*State]bool),
			epsTransition: make(map[*State]bool),
		}
	}

	return &NFA{
		initial: initialState,
		endPoints: endPoints,
		final: finalState,
	}
}

func dotNFA() *NFA{
	first := &State{
		transition: make(map[byte]map[*State]bool, 128),
		epsTransition: make(map[*State]bool),
	}

	second := &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool),
	}

	for i := 0; i < 128; i++ {
		first.transition[byte(i)] = make(map[*State]bool)
		first.transition[byte(i)][second] = true
	}

	delete(first.transition['\n'], second)

	resp := newNFA(false, true, false)
	resp.initial = first
	resp.endPoints[second] = true

	return resp
}

func simpleNFA(c byte) *NFA{
	first := &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool),
	}
	
	second := &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool),
	}

	first.transition[c] = make(map[*State]bool)
	first.transition[c][second] = true

	resp := newNFA(false, true, false)
	resp.initial = first
	resp.endPoints[second] = true

	return resp
}

func classNFA(class byte) *NFA{
	first := &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool),
	}
	second := &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool),
	}

	matchesClass := func(c byte) bool {
		return class == 'w' && isAlphaNum(c) ||
			class == 'W' && !isAlphaNum(c) ||
			class == 'd' && ('0' <= c && c <= '9') ||
			class == 'D' && !('0' <= c && c <= '9') ||
			class == 's' && (c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f') ||
			class == 'S' && !(c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f')
	}
	for i := 0; i < 128; i++ {
		if matchesClass(byte(i)) {
			first.transition[byte(i)] = make(map[*State]bool)
			first.transition[byte(i)][second] = true				
		} 
	}

	resp := newNFA(false, true, false)
	resp.initial = first
	resp.endPoints[second] = true
	
	return resp
}

func (nfa *NFA) applyRepetition(op byte){
	var newInit *State

	switch op {
	case '?': {
		newInit = &State{
			transition: make(map[byte]map[*State]bool),
			epsTransition: make(map[*State]bool),
		}

		newInit.epsTransition[nfa.initial] = true
		nfa.initial = newInit
		
		if nfa.endPoints == nil {
			nfa.endPoints = make(map[*State]bool)
		}
		nfa.endPoints[newInit] = true
	}
	case '+', '*': {
		if op == '*' {
			newInit = &State{
				transition: make(map[byte]map[*State]bool),
				epsTransition: make(map[*State]bool),
			}
			
			newInit.epsTransition[nfa.initial] = true
			nfa.initial = newInit
		}

		newEnd := &State{
			transition: make(map[byte]map[*State]bool),
			epsTransition: make(map[*State]bool),
		}

		for endPoint := range nfa.endPoints {
			endPoint.epsTransition[newEnd] = true
			delete(nfa.endPoints, endPoint)
		}
		nfa.endPoints[newEnd] = true
		newEnd.epsTransition[nfa.initial] = true

		if op == '*' {
			nfa.endPoints[newInit] = true
		}
	}
	}
}

func (nfa *NFA) concat(nfa2 *NFA) {
	for endPoint := range nfa.endPoints {
		endPoint.epsTransition[nfa2.initial] = true
	}
	nfa.endPoints = nfa2.endPoints
}

func (nfa *NFA) applyCardinality(min, max int) {
	var lastInit *State
	initialCopy := nfa.copy()

	if min == 0 {
		newInit := &State{
			transition: make(map[byte]map[*State]bool),
			epsTransition: make(map[*State]bool),
		}

		newInit.epsTransition[nfa.initial] = true
		nfa.initial = newInit

		if nfa.endPoints == nil {
			nfa.endPoints = make(map[*State]bool)
		}
		nfa.endPoints[newInit] = true
	} else {
		for i := 0; i < min - 1; i++ {
			nfasCopy := initialCopy.copy()

			for endPoint := range nfa.endPoints {
				endPoint.epsTransition[nfasCopy.initial] = true
			}

			lastInit = nfasCopy.initial
			nfa.endPoints = nfasCopy.endPoints
		}
	}

	if lastInit == nil { // case min == 0 or min == 1
		lastInit = nfa.initial
	}

	if max == infinite {
		for endPoint := range nfa.endPoints {
			endPoint.epsTransition[lastInit] = true
		}
	} else {
		lastEndPoints := nfa.endPoints
		for i := min; i < max; i++ {
			nfasCopy := initialCopy.copy()

			for endPoint := range lastEndPoints {
				endPoint.epsTransition[nfasCopy.initial] = true
			}
			
			for newEndPoint := range nfasCopy.endPoints {
				nfa.endPoints[newEndPoint] = true
			}

			lastEndPoints = nfasCopy.endPoints
		}
	}
}

func (nfa *NFA) applyDisjunction(nfa2 *NFA) {
	newInit := &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool),
	}

	newInit.epsTransition[nfa.initial] = true
	newInit.epsTransition[nfa2.initial] = true
	nfa.initial = newInit

	for endPoint2 := range nfa2.endPoints {
		nfa.endPoints[endPoint2] = true
	}
}

func (nfa *NFA) copy() *NFA{
	// maps *State from nfa to corresponding *State from resp
	match := make(map[*State]*State)

	hasInitial := nfa.initial != nil
	hasEndPoints := nfa.endPoints != nil
	hasFinal := nfa.final != nil
	resp := newNFA(hasInitial, hasEndPoints, hasFinal)

	if !hasInitial {
		return resp
	}

	match[nfa.initial] = resp.initial 
	copyRecursion(nfa.initial, match)

	// Is the final state also an endpoint?
	// if it is, then it might not be matched and,
	// as a result, resp.endPoints[match[nfa.final] == nil] = true
	// execution of line 34
	for originalEndPoint := range nfa.endPoints {
		resp.endPoints[match[originalEndPoint]] = true
	}

	return resp
}

func copyRecursion(original *State, match map[*State]*State) {
	copy := match[original]
	
	// normal transitions
	for c, nextStates := range original.transition {
		if copy.transition[c] == nil {
			copy.transition[c] = make(map[*State]bool)
		}

		copyNextStates := copy.transition[c]
		for nextState := range nextStates {
			if match[nextState] == nil {
				newState := &State{
					transition: make(map[byte]map[*State]bool),
					epsTransition: make(map[*State]bool),
				}
				match[nextState] = newState
				copyRecursion(nextState, match)
			}

			copyNextStates[match[nextState]] = true
		}
	}

	// epsilon transitions
	for nextState := range original.epsTransition {
		if match[nextState] == nil {
			newState := &State{
				transition: make(map[byte]map[*State]bool),
				epsTransition: make(map[*State]bool),
			}
			match[nextState] = newState
			copyRecursion(nextState, match)
		}

		copy.epsTransition[match[nextState]] = true
	}
}