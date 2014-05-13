package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type PhonemeGroup []string

type Phoneme struct {
	GroupVariable string
	PercentChance int
}

type Syllable []Phoneme

type SyllableSequence []string

func (syls SyllableSequence) IsContainedIn(sequence []string) bool {
	if len(syls) > len(sequence) {
		return false
	}

	index := 0
	for _, str := range sequence {
		if str == syls[index] {
			index++
		} else {
			index = 0
		}

		if index == len(syls) {
			return true
		}
	}

	return false
}

type Specification struct {
	MeanSyllables       int
	LowDeviation        int
	HighDeviation       int
	GenerateCount       int
	Seed                int64
	PhonemeVariables    map[string]PhonemeGroup
	SyllableVariables   map[string]Syllable
	SyllableNames       []string
	DisallowedSequences []SyllableSequence
}

func LoadSpecification(filename string) Specification {
	read, err := os.Open(filename)
	if err != nil {
		panic(err.Error())
	}

	defer func() {
		if err := read.Close(); err != nil {
			panic(err)
		}
	}()

	input, err := ioutil.ReadAll(read)
	if err != nil {
		panic(err.Error())
	}

	return parseSpecification(string(input))
}

func parseSpecification(input string) Specification {
	var spec Specification
	spec.PhonemeVariables = make(map[string]PhonemeGroup)
	spec.SyllableVariables = make(map[string]Syllable)

	lexer := NewLexer(input)
	empty := false
	for !empty {
		select {
		case lexeme, ok := <-lexer.lexemes:
			if !ok {
				empty = true
			} else {
				switch lexeme.lexType {
				case LEX_PHONEME_VARIABLE:
					spec.PhonemeVariables[lexeme.value] = parsePhonemeVariable(lexer)
				case LEX_SYLLABLE_VARIABLE:
					spec.SyllableVariables[lexeme.value] = parseSyllableVariable(lexer)
					spec.SyllableNames = append(spec.SyllableNames, lexeme.value)
				case LEX_DISALLOWED:
					spec.DisallowedSequences = append(spec.DisallowedSequences, parseDisallowed(lexer))
				case LEX_CONFIG_VARIABLE:
					parseConfigVariable(&spec, lexeme, lexer)
				}
			}
		}
	}

	return spec
}

func parsePhonemeVariable(l *Lexer) PhonemeGroup {
	group := PhonemeGroup{}
	for lexeme := range l.lexemes {
		if lexeme.lexType == LEX_END_DECLARATION {
			break
		}
		group = append(group, lexeme.value)
	}
	return group
}

func parseSyllableVariable(l *Lexer) Syllable {
	phonemes := []Phoneme{}
	for lexeme := range l.lexemes {
		if lexeme.lexType == LEX_END_DECLARATION {
			break
		}
		p := Phoneme{}
		if lexeme.lexType == LEX_NUMBER {
			num, err := strconv.ParseInt(lexeme.value, 10, 32)
			if err != nil {
				fmt.Printf("Bad number format: %s", lexeme.value)
				os.Exit(1)
			}
			lexeme = <-l.lexemes
			p.PercentChance = int(num)
		} else {
			p.PercentChance = 100
		}
		if lexeme.lexType == LEX_PHONEME_VARIABLE {
			p.GroupVariable = lexeme.value
		} else {
			fmt.Printf("Expected a phoneme variable name, but got %s", lexeme.value)
			os.Exit(1)
		}

		phonemes = append(phonemes, p)
	}
	return phonemes
}

func parseDisallowed(l *Lexer) SyllableSequence {
	seq := SyllableSequence{}
	for lexeme := range l.lexemes {
		if lexeme.lexType == LEX_END_DECLARATION {
			break
		}
		seq = append(seq, lexeme.value)
	}
	return seq
}

func parseConfigVariable(spec *Specification, lexeme Lexeme, l *Lexer) {
	num := int64(-1)
	next := <-l.lexemes
	if next.lexType == LEX_NUMBER {
		temp, err := strconv.ParseInt(next.value, 10, 64)
		if err != nil {
			fmt.Printf("Bad number format: %s", next.value)
			os.Exit(1)
		}
		num = temp
	} else {
		fmt.Printf(next.value)
		os.Exit(1)
	}
	switch lexeme.value {
	case "mean":
		spec.MeanSyllables = int(num)
	case "lowDeviation":
		spec.LowDeviation = int(num)
	case "highDeviation":
		spec.HighDeviation = int(num)
	case "words":
		spec.GenerateCount = int(num)
	case "seed":
		spec.Seed = num
	default:
		fmt.Printf("Unknown config variable '%s'", lexeme.value)
		os.Exit(1)
	}
}