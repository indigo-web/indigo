package main

type templateParserState uint8

const (
	eStatic templateParserState = iota + 1
	eSlash
	ePartName
	eEndPartName
)
