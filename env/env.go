package env

// This package declares the working environment (i.e. the structure of the data) and contains all the functions that perform file manipulations to prepare the environment

import (
	"bufio"
	"encoding/gob"
	"japp/searchgrids"
	"log"
	"os"

	"foosoft.net/projects/jmdict"
)

// The structure below holds pointers to a dictionary struct for JMdict as well as three pointers to search grids
// These grids are basically 3D arrays of linked lists that would allow us to perform quick lookup of words in English, Hiragana/Katakana and Kanji

type Environment struct {
	Dict    *jmdict.Jmdict
	English *searchgrids.EngAlphabet
	Kana    *searchgrids.KanaAlphabet
	Kanji   *searchgrids.KanjiAlphabet
	// Groups *searchgrids.Groups
}

// This is the first function that is called on bootup of the program - it checks for the pre-made environment encoded into a binary file
// If the file is missing, it creates one using the functions below
// If the read is successful, we simply return the pointer to the environment to the main function

func Initialize() (*Environment, error) {
	var env *Environment
	var err error
	envfilename := "env/envfile"
	if envfile, err := os.Open(envfilename); err == nil {
		env, err = readGobENV(envfile)
		if err != nil {
			log.Fatal("gob env: ", err)
		}
		defer envfile.Close()
	} else if os.IsNotExist(err) {
		env, err = writeGobENV()
		if err != nil {
			log.Fatal("gob env write: ", err)
		}
	}
	return env, err
}

// If the binary environment file is present, we decode it using this function

func readGobENV(file *os.File) (*Environment, error) {
	var env Environment
	var err error
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&env)
	return &env, err
}

// If it is missing, we create an environment struct, encode it into binary, and write it to the file
// First element of the struct is JMDict dictionary that we get by parsing the XML file using foosoft's module. These guys are our saviors!
// Elements 2-4 are more interesting and are explained in the searchgrids package

func writeGobENV() (*Environment, error) {
	var env Environment
	var err error
	env.Dict, err = dictInit()
	if err != nil {
		log.Fatal()
	}
	env.English, env.Kana, env.Kanji = searchgrids.GenerateAlphabets(*env.Dict)
	// env.Furigana = searchgrids.GenerateFuriganaSearchGrid(env.Dict)
	// env.Kanji = searchgrids.GenerateKanjiSearchGrid(env.Dict)
	envfile, err := os.Create("env/envfile")
	if err != nil {
		log.Fatal("env file write: ", err)
	}
	defer envfile.Close()
	encoder := gob.NewEncoder(envfile)
	err = encoder.Encode(env)
	if err != nil {
		log.Fatal("env encode: ", err)
	}
	return &env, err
}

// This function is the one that uses the foosoft parser to create a dictionary element

func dictInit() (*jmdict.Jmdict, error) {
	var dict jmdict.Jmdict
	var err error
	file, err := os.Open("env/JMdict_e")
	if err != nil {
		log.Fatal("JMdict file missing or corrupted: ", err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	dict, _, err = jmdict.LoadJmdict(reader)
	if err != nil {
		log.Fatal("JMdict file parsing error: ", err)
	}
	return &dict, err
}
