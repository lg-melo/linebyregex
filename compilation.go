package regex

import (
	"errors"
	"strconv"
)

func isSpecial(c byte) bool {
	switch c {
	case '(', ')', '[', ']', '{', ',', '}', '?', '+', '*', '\\', '|', '.':
		return true
	}

	return false
}

func isRepetition(c byte) bool {
	return c == '?' || c == '+' || c == '*'
}

func isClass(c byte) bool {
	switch c {
	case 'w', 'W', 'd', 'D', 's', 'S':
		return true
	}

	return false
}

func isAlphaNum(c byte) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_';
}

func cardinality(pattern string, start int) (err error, min, max, nextInd int){
	var start2 int
	
	if pattern[start] == ',' {
		min = 0
		start2 = start + 1
	} else {
		for i := start; i < len(pattern); i++ {
			if '0' <= pattern[i] && pattern[i] <= '9' {
				continue
			}

			if pattern[i] != ',' {
				return errors.New("invalid separator"), -1, -1, i
			}

			min, err = strconv.Atoi(pattern[start:i])
			start2 = i + 1
			break
		}
	}

	if start2 == 0 {
		return errors.New("separator never reached"), -1, -1, len(pattern)
	}

	if start2 < len(pattern) && pattern[start2] == ' ' {
		start2++
	}
	if start2 >= len(pattern) {
		return errors.New("no closing braces"), -1, -1, len(pattern)
	}

	if pattern[start2] == '}' {
		max = infinite
		nextInd = start2 + 1
	} else {
		for i := start2; i < len(pattern); i++ {
			if '0' <= pattern[i] && pattern[i] <= '9' {
				continue
			}

			if pattern[i] != '}' {
				return errors.New("wrong closing symbol"), -1, -1, i
			}

			max, err = strconv.Atoi(pattern[start2:i])
			nextInd = i + 1;
			break
		}
	}

	if nextInd == 0 {
		return errors.New("closing braces never reached"), -1, -1, len(pattern)
	}
	if min > max {
		return errors.New("wrong order of cardinalities"), -1, -1, nextInd
	}

	return nil, min, max, nextInd
}

func compileCharSet(pattern string, start int) (err error, nextInd int, resp *NFA) {
	if start >= len(pattern) {
		return errors.New("unclosed charset"), start, nil
	}

	var negation bool
	if pattern[start] == '^' {
		negation = true
		start++
	}
	if start >= len(pattern) {
		return errors.New("unclosed charset"), start, nil
	}

	var permissible [128]bool
	if negation {
		for i := range permissible {
			permissible[i] = true
		}
	}

	var hasClosed bool
	for i := start; i < len(pattern); i++ {
		v := pattern[i]

		if v != ']' && v != '-' {
			if i + 1 >= len(pattern) {
				return errors.New("charset is never closed"), i + 1, nil
			}

			if pattern[i+1] != '-' {
				permissible[v] = !negation
				continue
			}
			
			if i + 2 >= len(pattern) {
				return errors.New("charset is never closed"), i + 2, nil
			}

			if pattern[i+2] != ']'{
				if pattern[i] > pattern[i+2] {
					return errors.New("wrong charset interval"), nextInd, nil
				}
				
				for j := pattern[i]; j <= pattern[i+2]; j++ {
					permissible[j] = !negation
				}
				i = i + 2
			}

			continue
		}

		if v == ']' {
			hasClosed = true
			nextInd = i + 1
			break
		}
			
		// v == '-'
		if (i+1 < len(pattern) && pattern[i+1] == ']') ||
			(i == start && pattern[start - 1] != ']') {
			permissible[v] = !negation
		} else{
			return errors.New("invalid use of - at charset"), i + 1, nil
		}
	}

	if !hasClosed {
		return errors.New("unclosed charset"), nextInd, nil
	}

	resp = newNFA(true, true, false)
	for i, v := range permissible {
		if v {
			resp.applyDisjunction(simpleNFA(byte(i)))
		}
	}

	return nil, nextInd, resp
}

func compile(pattern string) (err error, nfa *NFA) {
	err, _, resp := auxCompile(pattern, 0)

	if err != nil {
		return err, nil
	}

	resp.final = &State{
		transition: make(map[byte]map[*State]bool),
		epsTransition: make(map[*State]bool),
	}

	for endPoint := range resp.endPoints {
		endPoint.epsTransition[resp.final] = true
	}

	return nil, resp
}

func auxCompile(pattern string, start int) (err error, nextInd int, resp *NFA){
	var tempNFA *NFA
	var matched bool
	var min, max int
	resp = nil
	
	for i := start; i < len(pattern); i++ {
		v := pattern[i]

		if v == '|' {
			err, nextInd, tempNFA := auxCompile(pattern, i + 1)

			if err != nil {
				return err, nextInd, nil
			}

			if resp == nil {
				resp = tempNFA
			} else {
				resp.applyDisjunction(tempNFA)
			}
			
			return nil, nextInd, resp
		}

		if v == ')' {
			if resp == nil {
				resp = newNFA(true, true, false)
			}
			return nil, i + 1, resp
		}

		tempNFA = nil
		matched = false
		
		if !isSpecial(v) || v == '.' || v == '\\' {
			if v == '\\' {
				if i + 1 >= len(pattern) {
					return errors.New("regex ends with \\"), -1, nil
				} 
				
				if isSpecial(pattern[i+1]) {
					tempNFA = simpleNFA(pattern[i+1])
				} else if isClass(pattern[i+1]) {
					tempNFA = classNFA(pattern[i+1])
				} else {
					return errors.New("invalid use of \\ operator"), -1, nil
				} // must also treat \b
				
				i++
			} else if v == '.' {
				tempNFA = dotNFA()
			} else {
				tempNFA = simpleNFA(v)
			}

			nextInd = i + 1
			matched = true
		}
		
		if v == '(' || v == '[' {
			if v == '(' {
				err, nextInd, tempNFA = auxCompile(pattern, i + 1)
			} else {
				err, nextInd, tempNFA = compileCharSet(pattern, i + 1)
			}
			
			if err != nil {
				return err, nextInd, nil
			}

			matched = true
		}
	
		if matched {
			if nextInd < len(pattern) {
				if isRepetition(pattern[nextInd]) {
					tempNFA.applyRepetition(pattern[nextInd])
					nextInd++
				} else if (pattern[nextInd] == '{') {
					err, min, max, nextInd = cardinality(pattern, nextInd + 1)
	
					if err != nil {
						return err, nextInd, nil
					}
	
					tempNFA.applyCardinality(min, max)
				}
			}
			
			i = nextInd - 1
			if resp == nil {
				resp = tempNFA
			} else {
				resp.concat(tempNFA)
			}
			
			continue
		}

		return errors.New("invalid pattern"), i + 1, nil
	}

	return errors.New("unclosed pattern"), -1, nil
}

func preparePattern(pattern string) string {
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