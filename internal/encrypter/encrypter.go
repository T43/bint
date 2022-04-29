package encrypter

import (
	"bint.com/internal/options"
	. "bint.com/pkg/serviceTools"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func insertStr(a []string, index int, value string) []string {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

func Encrypt(rootSource string, rootDest string, keyDest string) {
	var encryptDest *os.File
	var kDest *os.File
	var key string
	var filesListToExecute []string
	var fileToExecute string
	var splitedChunk []string
	var encryptedChunk []string
	var err error
	var res string
	var trash []string

	trash = []string{"AND", "OR", "NOT", "XOR", ".", "+", "-", "*", "/", "^", "print", "len",
		"index", "is_letter", "is_digit", "pop", "push", "input", "next_command", "get_root_source",
		"get_root_dest", "send_command", "goto", "SET_SOURCE", "SET_DEST", "SEND_DEST", "DEL_DEST",
		"UNDEFINE", "UNSET_SOURCE", "REROUTE", "UNSET_DEST", "RESET_SOURCE", "str", "int", "float",
		"bool", "(", ")", "[", "]", ":", "True", "False", "=", "<=", ">=", "==", "<", ">", ",", "string",
		"stack"}
	rand.Seed(time.Now().Unix())

	encryptDest, err = os.Create(rootDest)
	if nil != err {
		panic(err)
	}
	kDest, err = os.Create(keyDest)
	if nil != err {
		panic(err)
	}

	filesListToExecute = []string{rootSource}

	for _, fileToExecute = range filesListToExecute {
		f, err := os.Open(fileToExecute)

		if nil != err {
			panic(err)
		}

		newChunk := EachChunk(f)
		var trashNumbersList []int
		var trashNumber int

		for chunk := newChunk(); "end" != chunk; chunk = newChunk() {
			splitedChunk = strings.Split(chunk, options.BendSep)
			encryptedChunk = splitedChunk
			for i := 0; i < rand.Intn(options.MaxTrashLen); i++ {
				t := trash[rand.Intn(len(trash))]
				randArg := len(encryptedChunk) - 1 - trashNumber
				if randArg <= 0 {
					continue
				}
				trashNumber = rand.Intn(randArg) + trashNumber
				trashNumbersList = append(trashNumbersList, trashNumber)

				encryptedChunk = insertStr(encryptedChunk, trashNumber, t)
			}
			for i := 0; i < len(encryptedChunk); i++ {
				if i < len(encryptedChunk)-1 {
					res += encryptedChunk[i] + options.BendSep
				} else {
					res += encryptedChunk[i]
				}
			}
			_, err = encryptDest.WriteString(res + ";")
			if nil != err {
				panic(err)
			}
			var isTrash bool

			for i := 0; i < len(encryptedChunk); i++ {
				for _, tNum := range trashNumbersList {
					if i == tNum {
						isTrash = true
						break
					}
				}
				if !isTrash {
					key += fmt.Sprintf("%v", i) + ", "
				}
				isTrash = false
			}

			key = key[:len(key)-2]
			key += ";\n"

			_, err = kDest.WriteString(key)
			if nil != err {
				panic(err)
			}

			res = ""
			key = ""
			trashNumbersList = nil
		}
	}
}
