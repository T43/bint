package validator

import (
	"bint.com/internal/const/status"
	"bint.com/internal/executor"
	"bint.com/internal/lexer"
	"bint.com/internal/parser"
	. "bint.com/pkg/serviceTools"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var COMMAND_COUNTER int
var sourceCommandCounter int
var funcCommandCounter int
var fileName string
var sourceFile string
var lastFile string
var fileToValidate string
var retVal string
var isFunc bool
var wasRet bool
var toBlock bool

var funcTable map[string]string
var forCounter int

func checkVars(exprList interface{}, allVariables [][]interface{}) error {
	var isVar bool

	for _, expr := range exprList.([][]interface{}) {
		if "SUBEXPR" == fmt.Sprintf("%v", expr[0]) {
			err := checkVars(expr, allVariables)
			if nil != err {
				return err
			}
		}
		if "VAR" == fmt.Sprintf("%v", expr[0]) {
			newVariable := EachVariable(allVariables)
			for v := newVariable(); "end" != v[0]; v = newVariable() {
				if fmt.Sprintf("%v", expr[1]) == fmt.Sprintf("%v", v[1]) {
					isVar = true
					break
				}
			}
			if !isVar {
				return errors.New("unresolved reference: " + fmt.Sprintf("%v", expr[1]))
			} else {
				isVar = false
			}

		}

	}
	return nil
}

func getExprEnd(command string, startPos int) (int, error) {
	brOpened := 1
	brClosed := 0
	pos := startPos + 1

	for i := pos; i < len(command); i++ {
		if "(" == string(command[i]) {
			brOpened++
		}
		if ")" == string(command[i]) {
			brClosed++
		}

		if brOpened == brClosed {
			return i + 1, nil
		}
	}

	return 0, errors.New("invalid brace number")
}

func getSliceEnd(command string, startPos int) (int, error) {
	brOpened := 1
	brClosed := 0
	pos := startPos + 1

	for i := pos; i < len(command); i++ {
		if "[" == string(command[i]) {
			brOpened++
		}
		if "]" == string(command[i]) {
			brClosed++
		}

		if brOpened == brClosed {
			return i + 1, nil
		}
	}

	return 0, errors.New("invalid brace number")
}

func sysGetExprType(command string, variables [][][]interface{}) (string, error) {
	var exprList [][]interface{}
	var err error
	var modFlag bool

	var allVariables [][]interface{}

	for _, v := range variables {
		allVariables = append(allVariables, v...)
	}

	exprList, allVariables, err = lexer.LexicalAnalyze(command,
		allVariables, false, false, nil, false, nil, nil, nil, nil)

	if nil != err {
		return "", err
	}
	if "OP" == exprList[0][0] && IsKeyWordWithAssignment(fmt.Sprintf("%v", exprList[0][1])) {
		modFlag = true
		exprList = Insert(exprList, 0, []interface{}{"OP", "="})
		exprList = Insert(exprList, 0, []interface{}{"VAR", "$val"})
	}
	err = checkVars(exprList, allVariables)

	if nil != err {
		return "", err
	}

	_, infoListList, _, err := parser.Parse(exprList, allVariables, nil, false, false, false, nil, nil, nil)

	if nil != err && "slice bounds out of range" != err.Error() {
		return "", err
	}

	var res []interface{}

	if 1 == len(infoListList[0]) {
		res = infoListList[0]

		newVariable := EachVariable(allVariables)
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if fmt.Sprintf("%v", res[0]) == fmt.Sprintf("%v", v[1]) {
				return fmt.Sprintf("%v", v[0]), nil
			}
		}
	} else {
		if modFlag {
			infoListList[0] = infoListList[0][2:]
		}
		if 3 == len(infoListList[0]) && "int" == fmt.Sprintf("%v", infoListList[0][0]) &&
			"null" == fmt.Sprintf("%v", infoListList[0][2]) {
			t, err := getExprType(fmt.Sprintf("%v", infoListList[0][1]), variables)
			if nil != err {
				return ``, err
			}
			if "string" != t && "int" != t && "float" != t {
				return ``, errors.New("data type mismatch in int: " + t)
			}
			return "int", nil
		}
		if 3 == len(infoListList[0]) && "float" == fmt.Sprintf("%v", infoListList[0][0]) &&
			"null" == fmt.Sprintf("%v", infoListList[0][2]) {
			t, err := getExprType(fmt.Sprintf("%v", infoListList[0][1]), variables)
			if nil != err {
				return ``, err
			}
			if "string" != t && "int" != t && "float" != t {
				return ``, errors.New("data type mismatch in float: " + t)
			}
			return "float", nil
		}
		if 3 == len(infoListList[0]) && "bool" == fmt.Sprintf("%v", infoListList[0][0]) &&
			"null" == fmt.Sprintf("%v", infoListList[0][2]) {
			t, err := getExprType(fmt.Sprintf("%v", infoListList[0][1]), variables)
			if nil != err {
				return ``, err
			}
			if "string" != t && "bool" != t {
				return ``, errors.New("data type mismatch in bool: " + t)
			}
			return "bool", nil
		}
		if 3 == len(infoListList[0]) && "str" == fmt.Sprintf("%v", infoListList[0][0]) &&
			"null" == fmt.Sprintf("%v", infoListList[0][2]) {
			t, err := getExprType(fmt.Sprintf("%v", infoListList[0][1]), variables)
			if nil != err {
				return ``, err
			}
			if "stack" == t {
				return ``, errors.New("data type mismatch in str: stack")
			}
			return "string", nil
		}
		res, _, _ = executor.ExecuteTree(infoListList[0], allVariables, nil, false, false, nil, nil)
	}

	return WhatsType(fmt.Sprintf("%v", res[0])), nil
}

func getExprType(command string, variables [][][]interface{}) (string, error) {
	c1 := make(chan string, 1)
	e1 := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); nil != r {
				c1 <- ""
				e1 <- errors.New(fmt.Sprintf("%v", r))
			}
		}()
		text, err := sysGetExprType(command, variables)
		c1 <- text
		e1 <- err
	}()

	select {
	case res := <-c1:
		err := <-e1
		return res, err
	case <-time.After(time.Second):
		return ``, errors.New("unresolved command")
	}
}

func getLastFile() string {
	var fileToRet string
	f, err := os.Open(fileToValidate)
	if nil != err {
		handleError(err)
	}

	newChunk := EachChunk(f)

	for chunk, err := newChunk(); "end" != chunk; chunk, err = newChunk() {
		if nil != err {
			handleError(err)
		}
		inputedCode := CodeInput(chunk, false)
		if len(inputedCode) > 5 && "$file" == inputedCode[0:5] {
			fileToRet = inputedCode[5 : len(inputedCode)-1]
		}
	}

	err = f.Close()
	if nil != err {
		handleError(err)
	}

	return fileToRet
}

func filter(command string) bool {
	if "$" == string(command[0]) && "$" == string(command[len(command)-1]) {
		if len(command) > 5 && "$file" == command[0:5] {
			if "stdlib/core.b" == command[5:len(command)-1] ||
				("benv/trace/" == fileToValidate[0:11] && lastFile != command[5:len(command)-1]) {
				toBlock = true
			} else {
				if toBlock {
					toBlock = false
					funcCommandCounter = 0
					retVal = ""
					isFunc = false
					wasRet = false
					forCounter = 0
				}
			}
		}
		return false
	}

	return true
}
func dValidateUserStackCall(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:[[:alpha:]]+[[:alnum:]|_]*\.push\(.+\))`, command)

	tail2, stat2 := check(`(?:[[:alpha:]|$]+[[:alnum:]|_]*\.pop\([[:alpha:]]+[[:alnum:]|_]*\))`, command)
	if (status.Yes == stat && `` == tail) || (status.Yes == stat2 && `` == tail2) {
		var exprList [][]interface{}
		var err error

		var allVariables [][]interface{}

		for _, v := range variables {
			allVariables = append(allVariables, v...)
		}

		exprList, allVariables, err = lexer.LexicalAnalyze(command,
			allVariables, false, false, nil, false, nil, nil, nil, nil)
		if nil != err {
			return ``, status.Err, variables, err
		}

		err = checkVars(exprList, allVariables)
		if nil != err {
			return ``, status.Err, variables, err
		}

		return ``, status.Yes, variables, nil
	}
	return command, status.No, variables, nil
}

func dValidateInput(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:input\([[:alpha:]]+[[:alnum:]|_]*\))`, command)

	if status.Yes == stat && `` == tail {
		tail, _ = check(`(?:input\()`, command)
		tail = tail[:len(tail)-1]
		t, err := getExprType(tail, variables)
		if nil != err {
			return ``, status.Err, variables, err
		}
		if "string" != t {
			return ``, status.Err, variables, errors.New("data type mismatch in input: string and " + t)
		}
		return ``, status.Yes, variables, nil
	}

	return command, status.No, variables, nil
}

func dValidateNextCommand(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:next_command\([[:alpha:]]+[[:alnum:]|_]*\))`, command)

	if status.Yes == stat && `` == tail {
		tail, _ = check(`(?:next_command\()`, command)
		tail = tail[:len(tail)-1]
		t, err := getExprType(tail, variables)
		if nil != err {
			return ``, status.Err, variables, err
		}
		if "string" != t {
			return ``, status.Err, variables, errors.New("data type mismatch in next_command: string and " + t)
		}
		return ``, status.Yes, variables, nil
	}

	return command, status.No, variables, nil
}

func dValidateExit(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:exit\(.*?\))`, command)

	if status.Yes == stat && `` == tail {
		tail, _ = check(`(?:exit\()`, command)
		tail = tail[:len(tail)-1]
		t, err := getExprType(tail, variables)
		if nil != err {
			return ``, status.Err, variables, err
		}
		if "int" != t {
			return ``, status.Err, variables, errors.New("data type mismatch in exit: int and " + t)
		}
		return ``, status.Yes, variables, nil
	}

	return command, status.No, variables, nil
}

func dValidateSendCommand(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:send_command\(.*?\))`, command)

	if status.Yes == stat && `` == tail {
		tail, _ = check(`(?:send_command\()`, command)
		tail = tail[:len(tail)-1]
		t, err := getExprType(tail, variables)
		if nil != err {
			return ``, status.Err, variables, err
		}
		if "string" != t {
			return ``, status.Err, variables, errors.New("data type mismatch in send_command: string and " + t)
		}
		return ``, status.Yes, variables, nil
	}

	return command, status.No, variables, nil
}

func dValidateSetSource(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:SET_SOURCE\(.*\))`, command)

	if status.Yes == stat && `` == tail {
		tail, _ = check(`(?:SET_SOURCE\()`, command)
		tail = tail[:len(tail)-1]
		t, err := getExprType(tail, variables)
		if nil != err {
			return ``, status.Err, variables, err
		}
		if "string" != t {
			return ``, status.Err, variables, errors.New("data type mismatch in SET_SOURCE: string and " + t)
		}
		return ``, status.Yes, variables, nil
	}

	return command, status.No, variables, nil
}

func dValidateExists(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:exists\(.*\))`, command)

	if status.Yes == stat && `` == tail {
		tail, _ = check(`(?:exists\()`, command)
		tail = tail[:len(tail)-1]
		t, err := getExprType(tail, variables)
		if nil != err {
			return ``, status.Err, variables, err
		}
		if "string" != t {
			return ``, status.Err, variables, errors.New("data type mismatch in exists: string and " + t)
		}
		return ``, status.Yes, variables, nil
	}

	return command, status.No, variables, nil
}

func dValidateDelDest(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:DEL_DEST\(.*\))`, command)

	if status.Yes == stat && `` == tail {
		tail, _ = check(`(?:DEL_DEST\()`, command)
		tail = tail[:len(tail)-1]
		t, err := getExprType(tail, variables)
		if nil != err {
			return ``, status.Err, variables, err
		}
		if "string" != t {
			return ``, status.Err, variables, errors.New("data type mismatch in DEL_DEST: string and " + t)
		}
		return ``, status.Yes, variables, nil
	}

	return command, status.No, variables, nil
}

func dValidateSetDest(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:SET_DEST\(.*\))`, command)

	if status.Yes == stat && `` == tail {
		tail, _ = check(`(?:SET_DEST\()`, command)
		tail = tail[:len(tail)-1]
		t, err := getExprType(tail, variables)
		if nil != err {
			return ``, status.Err, variables, err
		}
		if "string" != t {
			return ``, status.Err, variables, errors.New("data type mismatch in SET_DEST: string and " + t)
		}
		return ``, status.Yes, variables, nil
	}

	return command, status.No, variables, nil
}

func handleError(newError error) {
	var inputedCode string
	var errorFile string

	f, err := os.Open(fileName)
	if nil != err {
		panic(err)
	}
	COMMAND_COUNTER--
	newChunk, err := SetCommandCounter(f, COMMAND_COUNTER)

	if nil != err {
		panic(err)
	}

	for chunk, err := newChunk(); "end" != chunk; chunk, err = newChunk() {
		if nil != err {
			panic(err)
		}
		CommandToExecute = strings.TrimSpace(chunk)
		inputedCode = CodeInput(chunk, false)

		if filter(inputedCode) {
			COMMAND_COUNTER--
			newChunk, err = SetCommandCounter(f, COMMAND_COUNTER)
			if nil != err {
				panic(err)
			}
		} else {
			break
		}
	}

	for chunk, err := newChunk(); "end" != chunk; chunk, err = newChunk() {
		if nil != err {
			panic(err)
		}
		CommandToExecute = strings.TrimSpace(chunk)
		errorFile = CodeInput(chunk, false)

		if filter(errorFile) {
			COMMAND_COUNTER--
			newChunk, err = SetCommandCounter(f, COMMAND_COUNTER)
			if nil != err {
				panic(err)
			}
		} else {
			if len(errorFile) > 5 && "$file" == errorFile[0:5] {
				errorFile = errorFile[5 : len(errorFile)-1]
				break
			}
		}
	}

	if "$trace" != inputedCode[0:6] {
		sourceCommandCounter = 1
	} else {
		sourceCommandCounter, err = strconv.Atoi(inputedCode[6 : len(inputedCode)-1])
		if nil != err {
			sourceCommandCounter = 1
		}
	}

	f, err = os.Open(errorFile)
	if nil != err {
		panic(err)
	}

	var chunk string

	newChunk = EachChunk(f)

	LineCounter = 1

	for counter := 0; counter < sourceCommandCounter; counter++ {
		chunk, err = newChunk()
		if nil != err {
			panic(err)
		}
		CodeInput(chunk, true)
	}

	fmt.Println("ERROR in " + errorFile + " at near line " +
		fmt.Sprintf("%v", LineCounter))
	fmt.Println(strings.TrimSpace(chunk))
	fmt.Println(newError.Error())

	err = f.Close()
	if nil != err {
		panic(err)
	}

	os.Exit(1)
}

func dValidateString(command string) (string, int, error) {
	tail := command
	re, err := regexp.Compile(`"(\\.|[^"])*"`)
	if nil != err {
		panic(err)
	}
	loc := re.FindIndex([]byte(tail))
	if nil != loc {
		i := loc[0]
		if i > 10 && "is_letter(" == tail[i-10:i] {
			if 3 != len(tail[loc[0]:loc[1]]) {
				return ``, status.Err, errors.New(`is_letter must have a string argument of length 1`)
			}
		}
		if i > 9 && "is_digit(" == tail[i-9:i] {
			if 3 != len(tail[loc[0]:loc[1]]) {
				return ``, status.Err, errors.New(`is_digit must have a string argument of length 1`)
			}
		}
		if i > 9 && "reg_find(" == tail[i-9:i] {
			_, err = regexp.Compile(tail[loc[0]+1 : loc[1]-1])
			if nil != err {
				return ``, status.Err, err
			}
		}
		tail = string(re.ReplaceAll([]byte(command), []byte(`$$val`)))
		return tail, status.Yes, nil
	}

	return tail, status.No, nil
}

func dValidateStr(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail := command
	var err error
	var re *regexp.Regexp
	var exprEnd int
	var t string
	var tail2 string

	re, err = regexp.Compile(`str\(`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), -1)
		var replacerArgs []string
		for _, pos := range poses {
			exprEnd, err = getExprEnd(tail, pos[1]-1)
			if nil != err {
				return tail, variables, err
			}
			t, err = getExprType(tail[pos[0]+4:exprEnd-1], variables)
			if nil != err {
				tail2, _, variables, err = dValidateFuncCall(tail[pos[0]+4:exprEnd-1], variables, true)
				if nil != err {
					return tail, variables, err
				}
				t, err = getExprType(tail2, variables)
				if nil != err {
					return ``, variables, err
				}
			}
			if "stack" == t {
				return tail, variables, errors.New("data type mismatch in str: stack")
			}
			replacerArgs = append(replacerArgs, tail[pos[0]:exprEnd])
			replacerArgs = append(replacerArgs, `$val`)
		}

		r := strings.NewReplacer(replacerArgs...)
		tail = r.Replace(tail)
	}
	return tail, variables, nil
}

func dValidateInt(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail := command
	var err error
	var re *regexp.Regexp
	var exprEnd int
	var t string
	var tail2 string

	re, err = regexp.Compile(`int\(`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), 1)

		for _, pos := range poses {
			if 0 == pos[0] || !(pos[0] > 0 && (unicode.IsLetter(rune(tail[pos[0]-1])) || '_' == tail[pos[0]-1] ||
				unicode.IsDigit(rune(tail[pos[0]-1])))) {
				exprEnd, err = getExprEnd(tail, pos[1]-1)
				if nil != err {
					return tail, variables, err
				}
				t, err = getExprType(tail[pos[0]+4:exprEnd-1], variables)
				if nil != err {
					tail2, _, variables, err = dValidateFuncCall(tail[pos[0]+4:exprEnd-1], variables, true)
					if nil != err {
						return tail, variables, err
					}
					t, err = getExprType(tail2, variables)
					if nil != err {
						return ``, variables, err
					}
				}
				if "stack" == t {
					return tail, variables, errors.New("data type mismatch in str: stack")
				}
				tail = tail[:pos[0]] + `$ival` + tail[exprEnd:]
				return dValidateInt(tail, variables)
			}
		}
	}
	return tail, variables, nil
}

func dValidateFloat(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail := command
	var err error
	var re *regexp.Regexp
	var exprEnd int
	var t string
	var tail2 string

	re, err = regexp.Compile(`float\(`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), 1)

		for _, pos := range poses {
			if 0 == pos[0] || !(pos[0] > 0 && (unicode.IsLetter(rune(tail[pos[0]-1])) || '_' == tail[pos[0]-1] ||
				unicode.IsDigit(rune(tail[pos[0]-1])))) {
				exprEnd, err = getExprEnd(tail, pos[1]-1)
				if nil != err {
					return tail, variables, err
				}
				t, err = getExprType(tail[pos[0]+6:exprEnd-1], variables)
				if nil != err {
					tail2, _, variables, err = dValidateFuncCall(tail[pos[0]+4:exprEnd-1], variables, true)
					if nil != err {
						return tail, variables, err
					}
					t, err = getExprType(tail2, variables)
					if nil != err {
						return ``, variables, err
					}
				}
				if "stack" == t {
					return tail, variables, errors.New("data type mismatch in str: stack")
				}
				tail = tail[:pos[0]] + `$fval` + tail[exprEnd:]
				return dValidateFloat(tail, variables)
			}
		}
	}
	return tail, variables, nil
}

func dValidateIsLetter(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail, variables, err := dValidateSlice(command, variables)
	if nil != err {
		return ``, variables, err
	}

	re, err := regexp.Compile(`is_letter\(`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), -1)
		var replacerArgs []string
		for _, pos := range poses {
			exprEnd, err := getExprEnd(tail, pos[1]-1)
			if nil != err {
				return tail, variables, err
			}
			t, err := getExprType(tail[pos[0]+10:exprEnd-1], variables)
			if nil != err {
				return tail, variables, err
			}
			if "string" != t {
				return tail, variables, errors.New("data type mismatch in is_letter: string and " + t)
			}
			replacerArgs = append(replacerArgs, tail[pos[0]:exprEnd])
			replacerArgs = append(replacerArgs, `$bval`)
		}

		r := strings.NewReplacer(replacerArgs...)
		tail = r.Replace(tail)
	}
	return tail, variables, nil
}

func dValidateIsDigit(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail, variables, err := dValidateSlice(command, variables)
	if nil != err {
		return ``, variables, err
	}
	re, err := regexp.Compile(`is_digit\(`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), -1)
		var replacerArgs []string
		for _, pos := range poses {
			exprEnd, err := getExprEnd(tail, pos[1]-1)
			if nil != err {
				return tail, variables, err
			}
			t, err := getExprType(tail[pos[0]+9:exprEnd-1], variables)
			if nil != err {
				return tail, variables, err
			}
			if "string" != t {
				return tail, variables, errors.New("data type mismatch in is_digit: string and " + t)
			}
			replacerArgs = append(replacerArgs, tail[pos[0]:exprEnd])
			replacerArgs = append(replacerArgs, `$bval`)
		}

		r := strings.NewReplacer(replacerArgs...)
		tail = r.Replace(tail)
	}
	return tail, variables, nil
}

func dValidateIndex(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail := command
	re, err := regexp.Compile(`index\(`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), -1)
		var replacerArgs []string
		for _, pos := range poses {
			exprEnd, err := getExprEnd(tail, pos[1]-1)
			if nil != err {
				return tail, variables, err
			}
			t, err := getExprType(tail[pos[0]+6:pos[0]+6+strings.Index(tail[pos[0]+6:exprEnd-1], ",")], variables)
			if nil != err {
				return tail, variables, err
			}
			if "string" != t {
				return tail, variables, errors.New("data type mismatch: first argument in index: str and " + t)
			}
			t, err = getExprType(tail[pos[0]+6+strings.Index(tail[pos[0]+6:exprEnd-1], ",")+1:exprEnd-1], variables)
			if nil != err {
				return tail, variables, err
			}
			if "string" != t {
				return tail, variables, errors.New("data type mismatch: second argument in index: str and " + t)
			}
			replacerArgs = append(replacerArgs, tail[pos[0]:exprEnd])
			replacerArgs = append(replacerArgs, `$ival`)
		}

		r := strings.NewReplacer(replacerArgs...)
		tail = r.Replace(tail)
	}
	return tail, variables, nil
}

func dValidateLen(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail := command
	re, err := regexp.Compile(`len\(`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), -1)
		var replacerArgs []string
		for _, pos := range poses {
			exprEnd, err := getExprEnd(tail, pos[1]-1)
			if nil != err {
				return tail, variables, err
			}
			t, err := getExprType(tail[pos[0]+4:exprEnd-1], variables)
			if nil != err {
				tail, variables, err = dValidateStr(command, variables)
				if nil != err {
					return tail, variables, err
				}
				tail, variables, err = dValidateInt(tail, variables)
				if nil != err {
					return tail, variables, err
				}
				tail, variables, err = dValidateSlice(tail, variables)
				if nil != err {
					return tail, variables, err
				}

				return dValidateLen(tail, variables)
			}
			if "string" != t {
				return tail, variables, errors.New("data type mismatch in len: " + t)
			}
			replacerArgs = append(replacerArgs, tail[pos[0]:exprEnd])
			replacerArgs = append(replacerArgs, `$ival`)
		}

		r := strings.NewReplacer(replacerArgs...)
		tail = r.Replace(tail)
	}
	return tail, variables, nil
}

func dValidateSliceInternal(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail, variables, err := dValidateStr(command, variables)

	if nil != err {
		return tail, variables, err
	}

	tail, variables, err = dValidateInt(tail, variables)

	if nil != err {
		return tail, variables, err
	}

	re, err := regexp.Compile(`\[`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), -1)
		var replacerArgs []string
		for _, pos := range poses {
			pos[1], err = getSliceEnd(tail, pos[0]+1)
			if nil != err {
				return ``, variables, err
			}
			colon := strings.Index(tail[pos[0]+1:pos[1]-1], ":")
			var subcommands []string
			if -1 != colon {
				subcommands = append(subcommands, tail[pos[0]+1 : pos[1]-1][0:colon])
				subcommands = append(subcommands, tail[pos[0]+1 : pos[1]-1][colon+1:len(tail[pos[0]+1:pos[1]-1])])
			} else {
				subcommands = append(subcommands, tail[pos[0]+1:pos[1]-1])
			}

			for _, com := range subcommands {
				t, err := getExprType(`$ival=`+com, variables)

				if nil != err {
					return dValidateSliceInternal(com, variables)
				}

				if "int" != t {
					return tail, variables, errors.New("data type mismatch in slice: int and " + t)
				}
			}
			replacerArgs = append(replacerArgs, tail[pos[0]+1:pos[1]-1])
			replacerArgs = append(replacerArgs, `$ival`)
		}

		r := strings.NewReplacer(replacerArgs...)
		tail = r.Replace(tail)
	}
	return tail, variables, nil
}

func dValidateSlice(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	var tail string
	var err error
	var replacerArgs []string

	tail, variables, err = dValidateSliceInternal(command, variables)

	if nil != err {
		return ``, variables, err
	}

	re, err := regexp.Compile(`[\$|[:alnum:]|_]+\[\$ival\]`)
	if nil != err {
		panic(err)
	}

	poses := re.FindAllIndex([]byte(tail), -1)

	for _, pos := range poses {
		t, err := getExprType(tail[pos[0]:pos[1]][0:strings.Index(tail[pos[0]:pos[1]], "[")], variables)
		if nil != err {
			return ``, variables, err
		}
		if "string" != t {
			return tail, variables, errors.New("slice can only be applied to a string, got: " + t)
		}

		replacerArgs = append(replacerArgs, tail[pos[0]:pos[1]])
		replacerArgs = append(replacerArgs, `$val`)
	}
	r := strings.NewReplacer(replacerArgs...)
	tail = r.Replace(tail)

	return tail, variables, nil
}
func dValidateRegFind(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail := command
	re, err := regexp.Compile(`reg_find\(`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(tail)) {
		var poses [][]int
		poses = re.FindAllIndex([]byte(tail), -1)
		var replacerArgs []string
		for _, pos := range poses {
			exprEnd, err := getExprEnd(tail, pos[1]-1)
			if nil != err {
				return tail, variables, err
			}
			t, err := getExprType(tail[pos[0]+9:pos[0]+9+strings.Index(tail[pos[0]+9:exprEnd-1], ",")], variables)
			if nil != err {
				return tail, variables, err
			}
			if "string" != t {
				return tail, variables, errors.New("data type mismatch: first argument in reg_find: str and " + t)
			}
			t, err = getExprType(tail[pos[0]+9+strings.Index(tail[pos[0]+9:exprEnd-1], ",")+1:exprEnd-1], variables)
			if nil != err {
				return tail, variables, err
			}

			if "string" != t {
				return tail, variables, errors.New("data type mismatch: second argument in reg_find: str and " + t)
			}
			replacerArgs = append(replacerArgs, tail[pos[0]:exprEnd])
			replacerArgs = append(replacerArgs, `$stackVal`)
		}

		r := strings.NewReplacer(replacerArgs...)
		tail = r.Replace(tail)
	}
	return tail, variables, nil
}

func dValidateFuncDefinition(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	var wasDef bool

	tail, stat := check(`(?m)(?:(int|float|bool|string|stack|void)[[:alnum:]|_]*?\`+
		`((?:((int|float|bool|string|stack))[[:alnum:]|_]+\,){0,})(int|float|bool|string|stack)[[:alnum:]|_]+\){`, command)
	if status.Yes == stat {
		funcCommandCounter = COMMAND_COUNTER
		isFunc = true
		wasDef = true

		closureHistory = append(closureHistory, brace{funcDefinition, LineCounter, CommandToExecute})
		reg, err := regexp.Compile(`^(?:(int|float|bool|string|stack|void))`)
		if nil != err {
			panic(err)
		}
		loc := reg.FindIndex([]byte(command))
		retVal = command[loc[0]:loc[1]]

		tail, stat = check(`^(?:(int|float|bool|string|stack|void)[[:alnum:]|_]+)`, command)
		if status.Yes != stat {
			return tail, status.Err, variables, errors.New("is not valid func definition")
		}

		reg, err = regexp.Compile(`(?:(int|float|bool|string|stack|void)[[:alnum:]|_]+)`)
		if nil != err {
			panic(err)
		}

		locs := reg.FindAllIndex([]byte(tail), -1)
		variables = append(variables, [][]interface{}{})

		for _, loc := range locs {
			_, variables[len(variables)-1], err = lexer.LexicalAnalyze(tail[loc[0]:loc[1]],
				variables[len(variables)-1], false, false, nil, false, nil, nil, nil, nil)
			if "int" == variables[len(variables)-1][len(variables[len(variables)-1])-1][0] {
				variables[len(variables)-1][len(variables[len(variables)-1])-1][2] = "1"
			}
			if "float" == variables[len(variables)-1][len(variables[len(variables)-1])-1][0] {
				variables[len(variables)-1][len(variables[len(variables)-1])-1][2] = "0.1"
			}
		}

		tail, stat = check(`(?m)(?:(int|float|bool|string|stack|void)[[:alnum:]|_]*?\`+
			`((?:((int|float|bool|string|stack))[[:alnum:]|_]+\,){0,})(int|float|bool|string|stack)[[:alnum:]|_]+\){`, command)
	}
	tail2, stat2 := check(`(?m)(?:(int|float|bool|string|stack|void)[[:alnum:]|_]*?\`+
		`(\){)`, command)
	if "" != tail2 {
		tail = tail2
	}
	if status.Yes == stat2 {
		funcCommandCounter = COMMAND_COUNTER
		isFunc = true
		wasDef = true
		variables = append(variables, [][]interface{}{})
		closureHistory = append(closureHistory, brace{funcDefinition, LineCounter, CommandToExecute})
		reg, err := regexp.Compile(`^(?:(int|float|bool|string|stack|void))`)
		if nil != err {
			panic(err)
		}
		loc := reg.FindIndex([]byte(command))
		retVal = command[loc[0]:loc[1]]
	}

	if wasDef {
		funcName, stat := check(`^(?:(int|float|bool|string|stack|void))`, command)
		if status.Yes != stat {
			return tail, status.Err, variables, errors.New("is not valid func definition")
		}

		funcName = funcName[0:strings.Index(funcName, "(")]

		if "" != funcTable[funcName] {
			return tail, status.Err, variables, errors.New("function polymorphism is not allowed")
		} else {
			funcTable[funcName] = retVal
			var err error
			if "void" != retVal {
				_, variables[0], err = lexer.LexicalAnalyze(funcTable[funcName]+"$"+funcName,
					variables[0], false, false, nil, false, nil, nil, nil, nil)
			} else {
				variables[0] = append(variables[0], []interface{}{"void", funcName, []interface{}{"func"}})
			}

			if nil != err {
				return tail, status.Err, variables, err
			}

			if "float" == funcTable[funcName] {
				variables[0][len(variables[0])-1][2] = "0.1"
			}
			if "int" == funcTable[funcName] {
				variables[0][len(variables[0])-1][2] = "1"
			}
		}
	}

	if status.Yes == stat || status.Yes == stat2 {
		stat = status.Yes
	} else {
		stat = status.No
	}

	return tail, stat, variables, nil
}
func dValidateVarDef(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?m)(?:(int|float|bool|string|stack)\$?[[:alnum:]|_]*)`, command)
	if status.Yes == stat && `` == tail {
		var err error
		_, variables[len(variables)-1], err = lexer.LexicalAnalyze(command,
			variables[len(variables)-1], false, false, nil, false, nil, nil, nil, nil)
		if nil != err {
			return tail, status.Err, variables, err
		}
		if "int" == variables[len(variables)-1][len(variables[len(variables)-1])-1][0] {
			variables[len(variables)-1][len(variables[len(variables)-1])-1][2] = "1"
		}
		if "float" == variables[len(variables)-1][len(variables[len(variables)-1])-1][0] {
			variables[len(variables)-1][len(variables[len(variables)-1])-1][2] = "0.1"
		}
	}
	return tail, stat, variables, nil
}

func dValidateIf(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:^if\([^{]+\){)`, command)

	if status.Yes == stat {
		closureHistory = append(closureHistory, brace{ifCond, LineCounter, CommandToExecute})
		variables = append(variables, [][]interface{}{})
		re, err := regexp.Compile(`(?:^if\([^{]+\){)`)
		if nil != err {
			panic(err)
		}
		loc := re.FindIndex([]byte(command))
		ifStruct := command[:loc[1]]

		ifStruct, _, variables, err = dValidateFuncCall("$bval="+ifStruct[3:len(ifStruct)-2], variables, false)
		ifStruct = ifStruct[6:]

		if nil != err {
			return tail, stat, variables, err
		}

		t, err := getExprType(ifStruct, variables)

		if "bool" != t || nil != err {
			t, err = getExprType("("+ifStruct+")", variables)
			if nil != err {
				return tail, status.Err, variables, err
			}
			if "bool" != t {
				return tail, status.Err, variables, errors.New("the expression inside if is not a boolean expression")
			}
		}
	}
	return tail, stat, variables, nil
}

func dValidateElseIf(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:^}elseif\([^{]+\){)`, command)

	if status.Yes == stat {
		closureHistory = append(closureHistory, brace{elseIfCond, LineCounter, CommandToExecute})
		variables = variables[:len(variables)-1]
		variables = append(variables, [][]interface{}{})
		re, err := regexp.Compile(`(?:^}elseif\([^{]+\){)`)
		if nil != err {
			panic(err)
		}
		loc := re.FindIndex([]byte(command))
		elseIfStruct := command[:loc[1]]
		elseIfStruct, _, variables, err = dValidateFuncCall("$bval="+elseIfStruct[7:len(elseIfStruct)-1], variables, false)
		elseIfStruct = elseIfStruct[6:]

		if nil != err {
			return tail, stat, variables, err
		}
		t, err := getExprType(elseIfStruct, variables)
		if nil != err {
			return tail, status.Err, variables, err
		}
		if "bool" != t {
			return tail, status.Err, variables, errors.New("the expression inside if is not a boolean expression")
		}

	}

	return command, stat, variables, nil
}

func dValidateElse(command string, variables [][][]interface{}) (string, int, [][][]interface{}) {
	tail, stat := check(`(?:^}else{)`, command)

	if status.Yes == stat {
		closureHistory = append(closureHistory, brace{elseCond, LineCounter, CommandToExecute})
		variables = variables[:len(variables)-1]
		variables = append(variables, [][]interface{}{})
	}
	return tail, stat, variables
}

func dValidateReturn(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	var err error
	re, err := regexp.Compile(`[^=]=[^=]`)
	if nil != err {
		panic(err)
	}
	if nil != re.FindIndex([]byte(command)) {
		return command, status.No, variables, nil
	}
	tail, stat := check(`^return.*`, command)
	if status.Yes == stat && "" == tail {
		tail = command[6:]
		if 1 == len(closureHistory) && funcDefinition == closureHistory[0].T {
			wasRet = true
		}
		if len(closureHistory) < 1 {
			return tail, status.Err, variables, errors.New("illegal position of return")
		}
		tail, _, variables, err = dValidateFuncCall(tail, variables, true)
		if nil != err {
			return tail, status.Err, variables, err
		}
		exprType, err := getExprType(tail, variables)
		if nil != err {
			return tail, status.Err, variables, err
		}

		if retVal != exprType {
			if !("float" == retVal && "int" == exprType) {
				return tail, status.Err, variables,
					errors.New("data type mismatch in func definition and return statement: " + retVal + " and " + exprType)
			}
		}
	} else {
		stat = status.No
	}
	return "", stat, variables, nil
}

func dValidateFuncCall(command string, variables [][][]interface{}, knownUsage bool) (string, int, [][][]interface{}, error) {
	var replacerArgs []string
	var thisFuncName string
	var err error

	tail := command

	for thisFuncName = range funcTable {
		if tail[1:] == thisFuncName && "void" == funcTable[thisFuncName] {
			return ``, status.Yes, variables, nil
		}
		if !knownUsage {
			if tail[1:] == thisFuncName && "void" != funcTable[thisFuncName] {
				return tail, status.Err, variables, errors.New("unused value of func " + thisFuncName)
			}
		}
	}

	for funcName := range funcTable {
		locArr := GetFuncNameEntry(funcName, tail)

		for _, loc := range locArr {
			thisFuncName = funcName
			loc[1], err = getExprEnd(tail, loc[1])
			if nil != err {
				return tail, status.Err, variables, err
			}

			replacerArgs = append(replacerArgs, tail[loc[0]:loc[1]])
			replacerArgs = append(replacerArgs, "$"+funcName)
		}

	}

	if nil != replacerArgs {
		r := strings.NewReplacer(replacerArgs...)
		tail = r.Replace(tail)

		for thisFuncName = range funcTable {

			if tail[1:] == thisFuncName && "void" == funcTable[thisFuncName] {
				return ``, status.Yes, variables, nil
			}
			if !knownUsage {
				if tail[1:] == thisFuncName && "void" != funcTable[thisFuncName] {
					return tail, status.Err, variables, errors.New("unused value of func " + thisFuncName)
				}
			}
		}
		return tail, status.Yes, variables, nil
	}

	return tail, status.No, variables, nil
}

func dValidatePrint(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:print\()`, command)
	if status.Yes == stat {
		T, err := getExprType(tail[:len(tail)-1], variables)
		if nil != err {
			return tail, status.Err, variables, err
		}
		if "string" != T {
			return tail, status.Err, variables, errors.New("print: data type mismatch: string and " + T)
		}
		return "", stat, variables, nil
	}
	return command, status.No, variables, nil
}

func dValidateUndefine(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	_, stat := check(`(?:^UNDEFINE\(\$.*?\))`, command)
	if status.Yes == stat {
		return "", stat, variables, nil
	}
	return command, status.No, variables, nil
}

func dValidateAssignment(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	_, stat := check(`(?:\$?[[:alpha:]][[:alnum:]|_]*={1}[^=]+)`, command)
	if status.Yes == stat {

		thisVar := command[0:strings.Index(command, "=")]
		expr := command[strings.Index(command, "=")+1:]

		var allVariables [][]interface{}

		for _, v := range variables {
			allVariables = append(allVariables, v...)
		}

		newVariable := EachVariable(allVariables)
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if thisVar == fmt.Sprintf("%v", v[1]) {
				T, err := getExprType(expr, variables)
				if nil != err {
					return ``, status.Err, variables, err
				}
				if v[0] != T {
					return ``, status.Err, variables,
						errors.New("data type mismatch: " + fmt.Sprintf("%v", v[0]) + " and " + T)
				} else {
					return ``, status.Yes, variables, nil
				}
			}
		}
		return ``, status.Err, variables, errors.New("unresolved reference: " + thisVar)
	}

	return ``, status.No, variables, nil
}

func dValidateWhile(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`(?:^while\([^{]+\){)`, command)

	if status.Yes == stat {
		closureHistory = append(closureHistory, brace{ifCond, LineCounter, CommandToExecute})
		variables = append(variables, [][]interface{}{})
		re, err := regexp.Compile(`(?:^while\([^{]+\){)`)
		if nil != err {
			panic(err)
		}
		loc := re.FindIndex([]byte(command))
		whileStruct := command[:loc[1]]
		t, err := getExprType(whileStruct[5:len(whileStruct)-1], variables)
		if nil != err {
			return tail, status.Err, variables, err
		}
		if "bool" != t {
			return tail, status.Err, variables,
				errors.New("the expression inside while is not a boolean expression")
		}
	}
	return tail, stat, variables, nil
}
func dValidateBreakContinue(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	var isLoop bool
	var commandType string

	commandType = "break"

	tail, stat := check(`^break`, command)
	if !("" == tail && status.Yes == stat) {
		commandType = "continue"
		tail, stat = check(`^continue`, command)
	}
	if "" == tail && status.Yes == stat {
		for _, brace := range closureHistory {
			if loop == brace.T {
				isLoop = true
				break
			}
		}
		if !isLoop {
			return tail, status.Err, variables, errors.New(commandType + " is outside of loop")
		}
	}
	return tail, stat, variables, nil
}

func dValidateDoWhile(command string, variables [][][]interface{}) (string, int, [][][]interface{}, error) {
	tail, stat := check(`^do{`, command)
	if stat == status.Yes {
		closureHistory = append(closureHistory, brace{loop, LineCounter, CommandToExecute})
		variables = append(variables, [][]interface{}{})
		return tail, stat, variables, nil
	}
	tail, stat = check(`(?:^}while\([^{]+\))`, command)

	if status.Yes == stat {
		variables = variables[:len(variables)-1]
		re, err := regexp.Compile(`(?:^}while\([^{]+\))`)
		if nil != err {
			panic(err)
		}
		loc := re.FindIndex([]byte(command))
		doWhileStruct := command[:loc[1]]
		t, err := getExprType(doWhileStruct[6:], variables)
		if nil != err {
			return tail, status.Err, variables, err
		}
		if "bool" != t {
			return tail, status.Err, variables, errors.New("the expression inside while is not a boolean expression")
		}

	}

	return tail, stat, variables, nil
}

func dValidateFor(command string, variables [][][]interface{}) (string, [][][]interface{}, error) {
	tail, stat := check(`^for\(`, command)
	if status.Yes == stat {
		closureHistory = append(closureHistory, brace{loop, LineCounter, CommandToExecute})
		variables = append(variables, [][]interface{}{})
		forCounter = 1
		return tail, variables, nil
	}
	if 1 == forCounter {
		forCounter++
		return command, variables, nil
	}
	if 2 == forCounter {
		forCounter++
		t, err := getExprType("("+command+")", variables)
		if nil != err {
			t, err = getExprType(command, variables)
			if nil != err {
				return tail, variables, err
			}
		}
		if "bool" != t {
			return tail, variables, errors.New("data type mismatch in for: bool and " + t)
		}
		return ``, variables, nil
	}
	if 3 == forCounter {
		forCounter = 0

		re, err := regexp.Compile(`\){`)

		loc := re.FindIndex([]byte(command))

		if nil == loc {
			return tail, variables, errors.New("invalid for syntax")
		}

		tempHistory := closureHistory
		variables, err = dynamicValidateCommand(command[0:loc[0]], variables)
		closureHistory = tempHistory
		if nil != err {
			return command, variables, nil
		}
		tail = command[loc[1]:]

		return tail, variables, nil
	}

	return command, variables, nil
}

func skipMark(command string) string {
	if toBlock {
		tail, stat := check(`(?:#[[:alpha:]|_]+[[:alnum:]]*:)`, command)
		if status.Yes == stat {
			return tail
		}
	}
	return command
}

func dynamicValidateCommand(command string, variables [][][]interface{}) ([][][]interface{}, error) {
	var again bool
	var err error

	command = skipMark(command)

	if isFunc && "void" != retVal && !wasRet && len(closureHistory) < 1 {
		COMMAND_COUNTER = funcCommandCounter
		return variables, errors.New("function must have unconditional return")
	}

	if isFunc && len(closureHistory) < 1 {
		retVal = ""
		isFunc = false
		wasRet = false
	}

	command, _, err = dValidateString(command)

	if nil != err {
		return variables, err
	}

	var tail string
	var stat int

	if toBlock {
		tail, stat, variables, err = dValidateFuncDefinition(command, variables)

		if nil != err {
			return variables, err
		}
		return variables, nil
	}

	command, variables, err = dValidateIndex(command, variables)
	if nil != err {
		return variables, err
	}

	command, variables, err = dValidateRegFind(command, variables)
	if nil != err {
		return variables, err
	}

	command, variables, err = dValidateIsLetter(command, variables)
	if nil != err {
		return variables, err
	}

	command, variables, err = dValidateIsDigit(command, variables)
	if nil != err {
		return variables, err
	}

	command, variables, err = dValidateLen(command, variables)
	if nil != err {
		return variables, err
	}

	command, variables, err = dValidateSlice(command, variables)

	if nil != err {
		return variables, err
	}

	command, variables, err = dValidateFor(command, variables)

	if nil != err {
		return variables, err
	}

	if "" == command {
		return variables, nil
	}

	command, stat, variables, err = dValidateUserStackCall(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateInput(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateNextCommand(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateExit(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateSendCommand(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateSetSource(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateExists(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateSetDest(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateDelDest(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	tail, stat, err = validateStandardFuncCall(command, "UNSET_SOURCE", 0, false)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		if `` == tail {
			return variables, nil
		}
	}

	tail, stat, err = validateStandardFuncCall(command, "UNSET_DEST", 0, false)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		if `` == tail {
			return variables, nil
		}
	}

	tail, stat, variables, err = dValidateFuncDefinition(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return dynamicValidateCommand(tail, variables)
	}

	tail, stat, variables, err = dValidateBreakContinue(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		if `` == tail {
			return variables, nil
		}
	}

	tail, stat, variables, err = dValidateVarDef(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		if `` == tail {
			return variables, nil
		}
	}

	tail, stat, variables, err = dValidateWhile(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return dynamicValidateCommand(tail, variables)
	}

	tail, stat, variables, err = dValidateDoWhile(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return dynamicValidateCommand(tail, variables)
	}

	_, stat, variables, err = dValidateReturn(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	_, stat, variables, err = dValidateUndefine(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	command, stat, variables, err = dValidateFuncCall(command, variables, false)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		again = true
	}

	command, variables, err = dValidateStr(command, variables)
	if nil != err {
		return variables, err
	}

	command, variables, err = dValidateInt(command, variables)
	if nil != err {
		return variables, err
	}

	command, variables, err = dValidateFloat(command, variables)
	if nil != err {
		return variables, err
	}

	_, stat, variables, err = dValidateAssignment(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	tail, stat, variables, err = dValidateIf(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return dynamicValidateCommand(tail, variables)
	}

	tail, stat, variables, err = dValidateElseIf(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return dynamicValidateCommand(tail, variables)
	}

	tail, stat, variables = dValidateElse(command, variables)

	if status.Yes == stat {
		return dynamicValidateCommand(tail, variables)
	}

	_, stat, variables, err = dValidatePrint(command, variables)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {
		return variables, nil
	}

	tail, variables, err = dValidateFor(command, variables)

	if nil != err {
		return variables, err
	}

	tail, stat, err = validateFigureBrace(command)

	if nil != err {
		return variables, err
	}

	if status.Yes == stat {

		variables = variables[:len(variables)-1]

		if tail == `` {
			return variables, nil
		}
	}

	if again {
		return dynamicValidateCommand(tail, variables)
	}

	return variables, errors.New("unresolved command")
}

func DynamicValidate(validatingFile string, rootSource string) {
	fileToValidate = validatingFile
	lastFile = getLastFile()

	fileName = validatingFile
	funcTable = make(map[string]string)

	var variables [][][]interface{}
	variables = append(variables, [][]interface{}{})
	var err error

	_, variables[len(variables)-1], err = lexer.LexicalAnalyze("string$val",
		variables[len(variables)-1], false, false, nil, false, nil, nil, nil, nil)
	if nil != err {
		panic(err)
	}
	variables[0][0][2] = "v"

	_, variables[len(variables)-1], err = lexer.LexicalAnalyze("int$ival",
		variables[len(variables)-1], false, false, nil, false, nil, nil, nil, nil)
	if nil != err {
		panic(err)
	}
	variables[0][1][2] = "1"

	_, variables[len(variables)-1], err = lexer.LexicalAnalyze("bool$bval",
		variables[len(variables)-1], false, false, nil, false, nil, nil, nil, nil)
	if nil != err {
		panic(err)
	}
	variables[0][2][2] = "False"

	_, variables[len(variables)-1], err = lexer.LexicalAnalyze("float$fval",
		variables[len(variables)-1], false, false, nil, false, nil, nil, nil, nil)
	if nil != err {
		panic(err)
	}
	variables[0][3][2] = "1.5"

	_, variables[len(variables)-1], err = lexer.LexicalAnalyze("stack$stackVal",
		variables[len(variables)-1], false, false, nil, false, nil, nil, nil, nil)
	if nil != err {
		panic(err)
	}

	sourceFile = rootSource
	COMMAND_COUNTER = 1

	f, err := os.Open(validatingFile)

	if nil != err {
		panic(err)
	}
	newChunk := EachChunk(f)

	for chunk, err := newChunk(); "end" != chunk; chunk, err = newChunk() {
		if nil != err {
			handleError(err)
		}
		CommandToExecute = strings.TrimSpace(chunk)
		inputedCode := CodeInput(chunk, false)
		COMMAND_COUNTER++

		if filter(inputedCode) {
			variables, err = dynamicValidateCommand(inputedCode, variables)
			if nil != err {
				handleError(err)
			}
		}
	}

	err = f.Close()
	if nil != err {
		panic(err)
	}
}
