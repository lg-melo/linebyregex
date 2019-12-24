package linebyregex

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

// Returns a new NFA
func newNFA(hasInitial, hasEndPoints, hasFinal bool) *NFA {
	var (
		initial *State
		endPoints map[*State]bool
		final *State
	)

	if hasInitial {
		initial = &State{
			transition: make(map[byte]map[*State]bool),
			epsTransition: make(map[*State]bool),
		}
	}

	if hasEndPoints {
		endPoints = make(map[*State]bool)
	}

	if hasFinal {
		final = &State{
			transition: make(map[byte]map[*State]bool),
			epsTransition: make(map[*State]bool),
		}
	}

	return &NFA{
		initial: initial,
		endPoints: endPoints,
		final: final,
	}
}

// Returns an NFA that recognizes anything but '\n'
func dotNFA() *NFA{
	first := &State{
		transition: make(map[byte]map[*State]bool, 128),
		epsTransition: make(map[*State]bool), // prob. unnecessary
	}

	second := &State{
		transition: make(map[byte]map[*State]bool), // prob. unnecessary
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

// Returns an NFA that recognizes 'c'
func simpleNFA(c byte) *NFA{
	first := &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool), // prob. unnecessary
	}
	
	second := &State{
		transition: make(map[byte]map[*State]bool), // prob. unnecessary
		epsTransition: make(map[*State]bool),
	}

	first.transition[c] = make(map[*State]bool)
	first.transition[c][second] = true

	resp := newNFA(false, true, false)
	resp.initial = first
	resp.endPoints[second] = true

	return resp
}

// Returns an NFA that recognizes a class (\d, \w, \s and their complements)
func classNFA(class byte) *NFA{
	first := &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool), // prob. unnecessary
	}
	second := &State{
		transition: make(map[byte]map[*State]bool), // prob. unnecessary
		epsTransition: make(map[*State]bool),
	}

	matchesClass := func(c byte) bool {
		return class == 'w' && isAlphaNum(c) ||
			class == 'W' && !isAlphaNum(c) ||
			class == 'd' && isDigit(c) ||
			class == 'D' && !isDigit(c) ||
			class == 's' && isSpace(c) ||
			class == 'S' && !isSpace(c)
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

// Each endPoint gets an epsTransition to nfa.initial
func (nfa *NFA) applyCycle(){
	for endPoint := range nfa.endPoints {
		endPoint.epsTransition[nfa.initial] = true
	}
}

// Initial state becomes an endPoint
func (nfa *NFA) applyOptionality(){
	newInit := &State{
		transition: make(map[byte]map[*State]bool), // prob. unnecessary,
		epsTransition: make(map[*State]bool),
	}
	
	newInit.epsTransition[nfa.initial] = true
	nfa.initial = newInit


	if nfa.endPoints == nil { // maybe unnecessary
		nfa.endPoints = make(map[*State]bool)
	}
	nfa.endPoints[newInit] = true
}

// Applies '?', '+' or '*'
func (nfa *NFA) applyRepetition(op byte){
	if op != '?' {
		nfa.applyCycle()
	}
	
	if op != '+' {
		nfa.applyOptionality()
	}
}

// Concatenates nfa2 onto nfa
func (nfa *NFA) concat(nfa2 *NFA) {
	for endPoint := range nfa.endPoints {
		endPoint.epsTransition[nfa2.initial] = true
	}
	nfa.endPoints = nfa2.endPoints
}

// Applies cardinality rule
func (nfa *NFA) applyCardinality(min, max int) {
	var lastInit *State
	initialCopy := nfa.copy()

	if min == 0 {
		nfa.applyOptionality()
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

	// min == 0 or min == 1
	if lastInit == nil { 
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

// Applies "r1 | r2" rule to nfa
func (nfa *NFA) applyDisjunction(nfa2 *NFA) {
	newInit := &State{
		transition: make(map[byte]map[*State]bool), // prob. unnecessary,
		epsTransition: make(map[*State]bool),
	}

	newInit.epsTransition[nfa.initial] = true
	newInit.epsTransition[nfa2.initial] = true
	nfa.initial = newInit

	for endPoint2 := range nfa2.endPoints {
		nfa.endPoints[endPoint2] = true
	}
}

// Returns an equivalent, but different, NFA wrt nfa
func (nfa *NFA) copy() *NFA{
	// maps each *State from nfa to corresponding *State from resp
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

	for originalEndPoint := range nfa.endPoints {
		resp.endPoints[match[originalEndPoint]] = true
	}

	return resp
}

// Auxiliary function to nfa's copy method
func copyRecursion(original *State, match map[*State]*State) {
	copy := match[original]
	
	// normal transitions
	for c, nextStates := range original.transition {
		copy.transition[c] = make(map[*State]bool)

		copysTrans := copy.transition[c]
		for nextState := range nextStates {
			if match[nextState] == nil {
				newState := &State{
					transition: make(map[byte]map[*State]bool),
					epsTransition: make(map[*State]bool),
				}
				match[nextState] = newState
				copyRecursion(nextState, match)
			}

			copysTrans[match[nextState]] = true
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