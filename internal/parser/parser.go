package parser

import (
	. "bint.com/internal/executor"
	. "bint.com/pkg/serviceTools"
	"errors"
	"fmt"
	"os"
	"strconv"
)

func maxBraces(exprList [][]interface{}) int {
	// считаем максимальное количество подряд идущих скобок
	nbraces := 0
	maxNbraces := 0

	for i := 0; i < len(exprList); i++ {
		if "(" == fmt.Sprintf("%v", exprList[i][1]) {
			nbraces += 1
		} else {
			if nbraces > maxNbraces {
				maxNbraces = nbraces
			}

			nbraces = 0
		}

	}

	return maxNbraces
}

func toStandardBinaryOperation(exprListInput [][]interface{}) [][]interface{} {
	exprList := exprListInput
	operation := fmt.Sprintf("%v", exprList[0][1])
	exprList = Pop(exprList, 0) // выталкиваем операцию
	exprList = Pop(exprList, 0) // выталкиваем скобки рядом с именем операции
	exprList = Pop(exprList, len(exprList)-1)
	exprList = Pop(exprList, 1) // выталкиваем запятую

	exprList = Insert(exprList, 1, []interface{}{"OP", operation})
	//exprList = Insert(exprList, 0, []interface{}{"BR", "("})
	//exprList = Insert(exprList, 2, []interface{}{"BR", ")"})
	//exprList = Insert(exprList, 4, []interface{}{"BR", "("})
	//exprList = Insert(exprList, len(exprList), []interface{}{"BR", ")"})

	return exprList
}

func makeOperationBinary(exprListInput [][]interface{}) [][]interface{} {
	exprList := exprListInput
	i := 0
	operation := fmt.Sprintf("%v", exprList[i][1])
	exprList = Pop(exprList, i) // выталкиваем операцию
	exprList = Pop(exprList, i) // выталкиваем скобки рядом с именем операции
	exprList = Pop(exprList, len(exprList)-1)

	exprList = append(exprList, []interface{}{"OP", operation})
	exprList = append(exprList, []interface{}{"VAL", "null", "null"})
	i += 3

	return exprList
}

func makePrintBinary(exprListInput [][]interface{}, variables [][]interface{}, usersStack []interface{}, showTree bool,
	toTranspile bool, toPrimitive bool, primitiveDest *os.File, transpileDest *os.File) ([][]interface{}, error) {
	var exprList [][]interface{}
	exprList = exprListInput
	i := 0

	if i+2 >= len(exprList) || "str" != exprList[i+2][1] {
		exprList = Pop(exprList, i) // выталкиваем print
		exprList = Pop(exprList, i) // выталкиваем скобки рядом с print
		exprList = Pop(exprList, len(exprList)-1)

		exprList = append(exprList, []interface{}{"OP", "print"})
		exprList = append(exprList, []interface{}{"VAL", "null", "null"})
		i += 3
	} else {
		j := i + 4
		if IsNumber(fmt.Sprintf("%v", exprList[j][1])) {
			exprList = Pop(exprList, i)
			exprList[i+1] = exprList[i+3]
			exprList[i+3] = []interface{}{"BR", ")"}
			exprList[i+2] = []interface{}{"OP", "str"}
			exprList = Insert(exprList, i+3, []interface{}{"VAL", "null", "null"})
			exprList = Insert(exprList, i+5, []interface{}{"OP", "print"})
			exprList = Insert(exprList, i+6, []interface{}{"VAL", "null", "null"})

			// выталкиваем последние две скобки
			exprList = Pop(exprList, len(exprList)-1)
			exprList = Pop(exprList, len(exprList)-1)
			i += 7
		} else { // булевское выражение
			startPos := j
			varFlag := true

			for j+1 < len(exprList) && (CanBePartOfBoolExpr(fmt.Sprintf("%v", exprList[j][1])) || varFlag) {
				j++
				newVariable := EachVariable(variables)
				for v := newVariable(); "end" != v[0]; v = newVariable() {
					if fmt.Sprintf("%v", v[1]) == fmt.Sprintf("%v", exprList[j][1]) {
						varFlag = true
						break
					}
				}
			}
			// возвращаемся на три символа назад:
			// найденный, не булевский
			// на закрывающуюся скобку str и
			// на закрывающуюся скобку print

			endPos := j - 3

			if j+1 > len(exprList) { // достигнут конец выражения
				// значит,  не булевский символ отсутствует
				endPos++
			}

			boolExpr := exprList[startPos : endPos+1]

			boolExpr = Insert(boolExpr, 0, []interface{}{"BR", "("})
			boolExpr = append(boolExpr, []interface{}{"BR", ")"})

			var finalRes interface{}

			if len(boolExpr) > 3 { // составное логическое выражение
				_, infoListList, _, err := Parse(boolExpr, variables, usersStack, showTree, toTranspile, toPrimitive,
					primitiveDest, transpileDest, nil)
				if nil != err {
					return exprList, err
				}
				infoList := infoListList[0]
				ExecuteTree(infoList, variables, usersStack, toTranspile, toPrimitive, primitiveDest, transpileDest)

			} else {
				finalRes = boolExpr[1][1]
			}

			// выталкиваем булевское выражение из expr_list
			for j = 0; j < len(boolExpr)-2; j++ { // длина boolExpr за исключением скобки вначале и вконце
				exprList = Pop(exprList, startPos)
			}
			// и помещаем на его место результат этого булевского выражения
			var strBoolExpr string
			for _, el := range boolExpr {
				strBoolExpr += fmt.Sprintf("%v", el[1])
			}
			exprList = Insert(exprList, startPos, []interface{}{"VAL", finalRes, strBoolExpr})

			// меняем формат под стандартную бинарную операцию
			exprList = Pop(exprList, i)
			exprList[i+1] = exprList[i+3]
			exprList[i+3] = []interface{}{"BR", ")"}
			exprList[i+2] = []interface{}{"OP", "str"}
			exprList = Insert(exprList, i+3, []interface{}{"VAL", "null", "null"})
			exprList = Insert(exprList, i+5, []interface{}{"OP", "print"})
			exprList = Insert(exprList, i+6, []interface{}{"VAL", "null", "null"})

			// выталкиваем последние две скобки
			exprList = Pop(exprList, len(exprList)-1)
			exprList = Pop(exprList, len(exprList)-1)

		}

	}

	return exprList, nil
}

func codeTree(exprList [][]interface{}, treeStructure string, infoList []interface{}) (string, []interface{}, error) {
	var err error

	if "" == treeStructure {
		treeStructure = "1"
	}
	i := 0

	for i < len(exprList) && "SUBEXPR" != fmt.Sprintf("%v", exprList[i][1]) {
		i += 1
	}

	if 1 == len(exprList) && 0 == i {
		treeStructure, infoList, err = codeTree(UnfoldInterfaceSlice(exprList[0][2].([]interface{})), treeStructure, infoList)
		return treeStructure, infoList, err
	}

	if i >= len(exprList) {
		treeStructure += "100100"
		i = 0
		for i < len(exprList) && "OP" != fmt.Sprintf("%v", exprList[i][0]) {
			i += 1
		}
		infoList = append(infoList, exprList[i][1])
		infoList = append(infoList, exprList[i-1][1])
		infoList = append(infoList, exprList[i+1][1])
		return treeStructure, infoList, nil
	}

	if i+1 < len(exprList) && "OP" == fmt.Sprintf("%v", exprList[i+1][0]) &&
		"SUBEXPR" == fmt.Sprintf("%v", exprList[i+2][1]) { // выражение и слева, и справа
		treeStructure += "1"
		infoList = append(infoList, fmt.Sprintf("%v", exprList[i+1][1]))
		treeStructure, infoList, err = codeTree(UnfoldInterfaceSlice(exprList[i][2].([]interface{})), treeStructure, infoList)
		if nil != err {
			return treeStructure, infoList, err
		}
		treeStructure += "1"
		treeStructure, infoList, err = codeTree(UnfoldInterfaceSlice(exprList[i+2][2].([]interface{})), treeStructure, infoList)
		if nil != err {
			return treeStructure, infoList, err
		}
	} else if i+1 < len(exprList) && "OP" == fmt.Sprintf("%v", exprList[i+1][0]) { // выражение слева
		treeStructure += "1"
		infoList = append(infoList, fmt.Sprintf("%v", exprList[i+1][1]))
		treeStructure, infoList, err = codeTree(UnfoldInterfaceSlice(exprList[i][2].([]interface{})), treeStructure, infoList)
		if nil != err {
			return treeStructure, infoList, err
		}
		infoList = append(infoList, fmt.Sprintf("%v", exprList[i+2][1]))
		treeStructure += "100"
	} else if i-1 >= 0 && "OP" == exprList[i-1][0] { // выражение справа
		treeStructure += "1001"
		infoList = append(infoList, fmt.Sprintf("%v", exprList[i-1][1]))
		infoList = append(infoList, fmt.Sprintf("%v", exprList[i-2][1]))
		treeStructure, infoList, err = codeTree(UnfoldInterfaceSlice(exprList[i][2].([]interface{})), treeStructure, infoList)
		if nil != err {
			return treeStructure, infoList, err
		}
	} else {
		return treeStructure, infoList, errors.New("codeTree: ERROR: wrong syntax")
	}

	return treeStructure, infoList, nil
}

func Parse(exprListInput [][]interface{}, variables [][]interface{}, usersStack []interface{}, showTree bool,
	toTranspile bool, toPrimitive bool, primitiveDest *os.File, transpileDest *os.File, programDest *os.File) ([]string, [][]interface{}, []interface{}, error) {
	const imgWidth = 1600
	const imgHeight = 800
	var treeStructure string
	var infoList []interface{}
	var err error
	var treeStructureList []string
	var infoListList [][]interface{}
	var sizeError error

	exprList := exprListInput

	maxNbraces := maxBraces(exprList)

	wasGoto := false
	wasSetSource := false
	wasNextCommand := false
	wasGetRootSource := false
	wasGetRootDest := false
	wasSendCommand := false
	wasUndefine := false
	wasPush := false
	wasPop := false
	wasSetDest := false
	wasDelDest := false
	wasSendDest := false
	wasPoint := false
	wasLen := false
	wasIndex := false
	wasIsLetter := false
	wasIsDigit := false
	wasExit := false
	wasExists := false

	if "goto" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasGoto = true
	}
	if "exit" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasExit = true
	}
	if "SET_SOURCE" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasSetSource = true
	}
	if "SET_DEST" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasSetDest = true
	}
	if "SEND_DEST" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasSendDest = true
	}
	if "DEL_DEST" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasDelDest = true
	}
	if "UNDEFINE" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasUndefine = true
	}
	if "push" == fmt.Sprintf("%v", exprList[0][1]) {
		//exprList = Insert(exprList, 1, []interface{}{"BR", "("})
		//exprList = append(exprList, []interface{}{"BR", ")"})
		exprList = makeOperationBinary(exprList)
		wasPush = true
	}
	if "pop" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasPop = true
	}
	if "next_command" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasNextCommand = true
	}
	if "get_root_source" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasGetRootSource = true
	}
	if "get_root_dest" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasGetRootDest = true
	}
	if "send_command" == fmt.Sprintf("%v", exprList[0][1]) {
		exprList = makeOperationBinary(exprList)
		wasSendCommand = true
	}

	if len(exprList) > 1 && "." == fmt.Sprintf("%v", exprList[1][1]) {
		exprList[2] = []interface{}{"VAL", []string{exprList[2][1].(string), exprList[4][1].(string)}}
		exprList = Pop(exprList, 3)
		exprList = Pop(exprList, 3)
		exprList = Pop(exprList, 3)
		wasPoint = true
	}

	// заменяем NOT на стандартную бинарную операцию
	i := 0
	wasNOT := false
	var newNotList [][]interface{}
	for i < len(exprList) {
		if "NOT" == fmt.Sprintf("%v", exprList[i][1]) && "OP" == fmt.Sprintf("%v", exprList[i][0]) &&
			"null" != fmt.Sprintf("%v", exprList[i+1][1]) {
			wasNOT = true
			endNOT := FindExprListEnd(exprList, i+2)

			newNotList = nil
			for _, el := range makeOperationBinary(exprList[i:endNOT]) {
				newNotList = append(newNotList, el)
			}

			newNotList = Insert(newNotList, len(newNotList)-2, []interface{}{"BR", ")"})
			newNotList = Insert(newNotList, 0, []interface{}{"BR", "("})
			//exprList = newNotList
			var tempList [][]interface{}
			for _, el := range exprList[:i] {
				tempList = append(tempList, el)
			}
			tempList = append(tempList, newNotList...)
			exprList = append(tempList, exprList[endNOT:]...)

			/*var res string
			for _, el := range exprList{
				res += el[1].(string)
			}
			fmt.Println(res)*/
			i = 0
		}
		i++
	}

	i = 1

	for i < len(exprList) {
		// режем строку
		if "[" == fmt.Sprintf("%v", exprList[i][1]) && (IsNumber(fmt.Sprintf("%v", exprList[i+1][1])) ||
			"VAR" == fmt.Sprintf("%v", exprList[i+1][0])) {
			varName := fmt.Sprintf("%v", exprList[i-1][1])
			var varVal string

			newVariable := EachVariable(variables)
			for v := newVariable(); "end" != v[0]; v = newVariable() {
				if fmt.Sprintf("%v", v[1]) == varName {
					varVal = fmt.Sprintf("%v", ValueFoldInterface(v[2]))

					if len(varVal) > 2 && `"` == string([]rune(varVal)[0]) && `"` == string([]rune(varVal)[len(varVal)-1]) {
						varVal = string([]rune(varVal)[1 : len([]rune(varVal))-1])
					}
					//if "\"" == string(varVal[0]) {
					//	varVal = varVal[1 : len(varVal)-1]
					//}

					break
				}
			}

			j := i + 1

			for "]" != exprList[j][1] {
				j += 1
			}
			exprListInside := exprList[i+1 : j]

			isColon := false
			for _, el := range exprListInside {
				if ":" == fmt.Sprintf("%v", el[1]) {
					isColon = true
					break
				}
			}
			if isColon {
				if "VAR" == fmt.Sprintf("%v", exprListInside[0][0]) {
					newVariable = EachVariable(variables)

					for v := newVariable(); "end" != v[0]; v = newVariable() {
						if fmt.Sprintf("%v", exprListInside[0][1]) == fmt.Sprintf("%v", v[1]) {
							if !toTranspile {
								exprListInside[0][0] = "VAL"
								exprListInside[0][1] = v[2]
							} else {
								exprListInside[0][0] = "VAL"
								exprListInside[0][1] = "toInt(getVar(\"" + fmt.Sprintf("%v", v[1]) + "\"))"
							}
						}
					}
				}

				if "VAR" == fmt.Sprintf("%v", exprListInside[2][0]) {
					newVariable = EachVariable(variables)
					for v := newVariable(); "end" != v[0]; v = newVariable() {
						if fmt.Sprintf("%v", exprListInside[2][1]) == fmt.Sprintf("%v", v[1]) {
							if !toTranspile {
								exprList[2][0] = "VAL"
								exprListInside[2][1] = v[2]
							} else {
								exprListInside[2][0] = "VAL"
								exprListInside[2][1] = "toInt(getVar(\"" + fmt.Sprintf("%v", v[1]) + "\"))"
							}

						}
					}
				}
				var leftNumber interface{}
				var rightNumber interface{}

				leftNumber, err := strconv.Atoi(fmt.Sprintf("%v", ValueFoldInterface(exprListInside[0][1])))
				if nil != err && !toTranspile {
					err = errors.New("parser: ERROR: data type mismatch")
					panic(err)
				} else if toTranspile {
					leftNumber = fmt.Sprintf("%v", ValueFoldInterface(exprListInside[0][1]))
				}

				rightNumber, err = strconv.Atoi(fmt.Sprintf("%v", ValueFoldInterface(exprListInside[2][1])))
				if nil != err && !toTranspile {
					err = errors.New("parser: ERROR: data type mismatch")
					panic(err)
				} else if toTranspile {
					rightNumber = fmt.Sprintf("%v", ValueFoldInterface(exprListInside[2][1]))
				}

				for k := 0; k < 6; k++ {
					exprList = Pop(exprList, i-1) // выражение
				}
				if !toTranspile {
					if rightNumber.(int) >= leftNumber.(int) && rightNumber.(int) <= len(varVal) {
						exprList = Insert(exprList, i-1, []interface{}{"VAL", "\"" +
							string([]rune(varVal)[leftNumber.(int):rightNumber.(int)]) + "\""})
					} else {
						exprList = Insert(exprList, i-1, []interface{}{"VAL", "\"\""})
						sizeError = errors.New("slice bounds out of range")
					}
				} else {
					exprList = Insert(exprList, i-1, []interface{}{"VAL", "\"" + "getVar(\"" + varName + "\").(string)[" +
						fmt.Sprintf("%v", leftNumber) + ":" + fmt.Sprintf("%v", rightNumber) + "]\""})
				}

			} else {
				var number interface{}
				number, err := strconv.Atoi(fmt.Sprintf("%v", exprListInside[0][1]))
				if nil != err {
					newVariable = EachVariable(variables)
					for v := newVariable(); "end" != v[0]; v = newVariable() {
						if fmt.Sprintf("%v", v[1]) == fmt.Sprintf("%v", exprListInside[0][1]) {
							if !toTranspile {
								exprListInside[0][0] = "VAL"
								exprListInside[0][1] = v[2]
							} else {
								exprListInside[0][0] = "VAL"
								exprListInside[0][1] = "toInt(getVar(\"" + fmt.Sprintf("%v", v[1]) + "\"))"
							}
						}
					}

					number, err = strconv.Atoi(fmt.Sprintf("%v", ValueFoldInterface(exprListInside[0][1])))

					if nil != err && !toTranspile {
						err = errors.New("parser: ERROR: data type mismatch")
						panic(err)
					} else if toTranspile {
						number = fmt.Sprintf("%v", ValueFoldInterface(exprListInside[0][1]))
					}
				}

				for k := 0; k < 4; k++ {
					exprList = Pop(exprList, i-1)
				}
				if !toTranspile {
					if number.(int) < len(varVal) {
						exprList = Insert(exprList, i-1, []interface{}{"VAL", "\"" + string([]rune(varVal)[number.(int)]) + "\""})
					} else {
						exprList = Insert(exprList, i-1, []interface{}{"VAL", "\"\""})
						sizeError = errors.New("slice bounds out of range")
					}
				} else {
					exprList = Insert(exprList, i-1, []interface{}{"VAL", "string(getVar(\"" + varName + "\").(string)[" +
						fmt.Sprintf("%v", number) + "])"})
				}
			}

		}
		i += 1
	}

	if maxNbraces > 0 {
		i = 0
		// убираем скобки непосредственно рядом с выражениями
		for i < len(exprList) {
			if ("VAL" == fmt.Sprintf("%v", exprList[i][0]) || "VAR" == fmt.Sprintf("%v", exprList[i][0])) &&
				(i-1 >= 0 && i+1 < len(exprList)) &&
				("(" == fmt.Sprintf("%v", exprList[i-1][1]) && ")" == fmt.Sprintf("%v", exprList[i+1][1])) {
				if (i-2 < 0) || ("OP" != exprList[i-2][0] || !IsUnaryOperation(fmt.Sprintf("%v", exprList[i-2][1]))) {
					exprList = Pop(exprList, i-1)
					exprList = Pop(exprList, i)
				}
			}
			i++
		}
	}

	i = 0

	wasCd := false
	wasAssignment := false

	for i < len(exprList) {
		if "=" == exprList[i][1] {
			wasAssignment = true
			exprOutside := exprList[:i]
			exprInside := exprList[i+1:]
			if "int" == exprInside[0][1] || "float" == exprInside[0][1] || "bool" == exprInside[0][1] ||
				"str" == exprInside[0][1] {
				exprInside = makeOperationBinary(exprInside)
				exprInside = Insert(exprInside, 0, []interface{}{"BR", "("})
				exprInside = append(exprInside, []interface{}{"BR", ")"})
				i += 7
			}
			if "len" == exprInside[0][1] {
				exprInside = makeOperationBinary(exprInside)
				exprInside = Insert(exprInside, 0, []interface{}{"BR", "("})
				exprInside = append(exprInside, []interface{}{"BR", ")"})
				wasLen = true
				i += 7
			}
			if "exists" == exprInside[0][1] {
				exprInside = makeOperationBinary(exprInside)
				exprInside = Insert(exprInside, 0, []interface{}{"BR", "("})
				exprInside = append(exprInside, []interface{}{"BR", ")"})
				wasExists = true
				i += 7
			}
			if "index" == exprInside[0][1] {
				wasIndex = true

				operation := fmt.Sprintf("%v", exprInside[0][1])
				exprInside = Pop(exprInside, 0) // выталкиваем операцию
				exprInside = Pop(exprInside, 0) // выталкиваем скобки рядом с именем операции
				exprInside = Pop(exprInside, 1) // выталкиваем запятую
				rightVal := fmt.Sprintf("%v", exprInside[1][1])
				exprInside = Pop(exprInside, 1)
				exprInside = Pop(exprInside, len(exprInside)-1)
				exprInside = append(exprInside, []interface{}{"OP", operation})
				exprInside = append(exprInside, []interface{}{"VAL", rightVal})
				exprInside = Insert(exprInside, 0, []interface{}{"BR", "("})
				exprInside = append(exprInside, []interface{}{"BR", ")"})

			}
			if "is_letter" == exprInside[0][1] {
				exprInside = makeOperationBinary(exprInside)
				exprInside = Insert(exprInside, 0, []interface{}{"BR", "("})
				exprInside = append(exprInside, []interface{}{"BR", ")"})
				wasIsLetter = true
				i += 7
			}
			if "is_digit" == exprInside[0][1] {
				exprInside = makeOperationBinary(exprInside)
				exprInside = Insert(exprInside, 0, []interface{}{"BR", "("})
				exprInside = append(exprInside, []interface{}{"BR", ")"})
				wasIsDigit = true
				i += 7
			}
			if "reg_find" == exprInside[0][1] {
				exprInside = toStandardBinaryOperation(exprInside)
				exprInside = Insert(exprInside, 0, []interface{}{"BR", "("})
				exprInside = append(exprInside, []interface{}{"BR", ")"})
			}

			exprList = append(exprOutside, []interface{}{"OP", "="})
			exprList = append(exprList, exprInside...)
			break
		}
		i++
	}

	i = 0

	// выполняем условную дизъюнкцию
	if "[" == exprList[i][1] {
		wasCd = true
		j := i + 1
		var leftExpr [][]interface{}
		for "," != exprList[j][1] {
			leftExpr = append(leftExpr, exprList[j])
			j++
		}
		j++

		var condition [][]interface{}

		for "," != exprList[j][1] {
			condition = append(condition, exprList[j])
			j++
		}
		j++

		var rightExpr [][]interface{}

		for "]" != exprList[j][1] {
			rightExpr = append(rightExpr, exprList[j])
			j++
		}

		// заменяем выражение на более простое
		// выталкиваем выражение
		// выражение начинается с индекса i
		for "]" != exprList[i][1] {
			exprList = Pop(exprList, i)
		}
		exprList = Pop(exprList, i)

		var resCon []interface{}

		if "print" == leftExpr[0][1] {
			exprList = append(exprList, []interface{}{"BR", "("})
			binaryPrint, err := makePrintBinary(leftExpr, variables,
				usersStack, showTree, toTranspile, toPrimitive, primitiveDest, transpileDest)
			if nil != err {
				return treeStructureList, infoListList, usersStack, err
			}
			exprList = append(exprList, binaryPrint...)
			exprList = append(exprList, []interface{}{"BR", ")"})
		} else if "goto" == leftExpr[0][1] {
			exprList = append(exprList, []interface{}{"BR", "("})
			exprList = append(exprList, makeOperationBinary(leftExpr)...)
			exprList = append(exprList, []interface{}{"BR", ")"})
		} else {
			_, infoListList, _, err = Parse(leftExpr, variables, usersStack, showTree, toTranspile, toPrimitive,
				primitiveDest, transpileDest, nil)
			if nil != err {
				return treeStructureList, infoListList, usersStack, err
			}
			infoList := infoListList[0]
			_, variables, usersStack = ExecuteTree(infoList, variables, usersStack, toTranspile, toPrimitive,
				primitiveDest, transpileDest)
		}

		// неоднозначное условие
		if len(condition) > 1 {

			_, infoListList, _, err = Parse(condition, variables, usersStack, showTree, toTranspile, toPrimitive,
				primitiveDest, transpileDest, nil)
			if nil != err {
				return treeStructureList, infoListList, usersStack, err
			}
			infoList := infoListList[0]

			resCon, variables, usersStack = ExecuteTree(infoList, variables, usersStack, toTranspile, toPrimitive,
				primitiveDest, transpileDest)

			if toTranspile {
				_, err := transpileDest.WriteString("if " + fmt.Sprintf("%v", fmt.Sprintf("%v", resCon[0])) + "{\n")
				if nil != err {
					panic(err)
				}
			}
		} else {
			transpileVar := condition[0][1]
			var wasVar bool

			newVariable := EachVariable(variables)
			for v := newVariable(); "end" != v[0]; v = newVariable() {
				if v[1] == condition[0][1] {
					condition[0][1] = v[2]
					transpileVar = v[1]
					wasVar = true
					if toTranspile {
						_, err := transpileDest.WriteString("if toBool(getVar(\"" + fmt.Sprintf("%v", transpileVar) + "\")){\n")
						if nil != err {
							panic(err)
						}
					}
				}
			}

			if toTranspile && !wasVar {
				_, err := transpileDest.WriteString("if " + fmt.Sprintf("%v", StrToBool(fmt.Sprintf("%v", transpileVar))) + "{\n")
				if nil != err {
					panic(err)
				}
			}

			resCon = []interface{}{ValueFoldInterface(condition[0][1])}

			if nil != programDest {
				if "True" == condition[0][1] || "true" == condition[0][1] {
					_, err := programDest.Write([]byte("\n movb $1, (userData)"))
					if nil != err {
						fmt.Println(err)
						os.Exit(1)
					}
				}
				if "False" == condition[0][1] || "false" == condition[0][1] {
					_, err := programDest.Write([]byte("\n movb $0, (userData)"))
					if nil != err {
						fmt.Println(err)
						os.Exit(1)
					}
				}

			}

			if "bool" != WhatsType(fmt.Sprintf("%v", resCon[0])) && !toTranspile {
				panic("parser: ERROR: data type mismatch")
			}
		}

		if !toTranspile && !toPrimitive {
			exprList = append(exprList, []interface{}{"OP", "L: " + fmt.Sprintf("%v", resCon[0])})
		} else {
			exprList = append(exprList, []interface{}{"OP", "CD"})
		}

		if "print" == rightExpr[0][1] {
			exprList = append(exprList, []interface{}{"BR", "("})
			binaryPrint, err := makePrintBinary(rightExpr, variables, usersStack, showTree,
				toTranspile, toPrimitive, primitiveDest, transpileDest)
			if nil != err {
				return treeStructureList, infoListList, usersStack, err
			}
			exprList = append(exprList, binaryPrint...)
			exprList = append(exprList, []interface{}{"BR", ")"})
		} else if "goto" == rightExpr[0][1] {
			exprList = append(exprList, []interface{}{"BR", "("})
			exprList = append(exprList, makeOperationBinary(rightExpr)...)
			exprList = append(exprList, []interface{}{"BR", ")"})
		} else {
			_, infoListList, _, err = Parse(rightExpr, variables, usersStack, showTree, toTranspile, toPrimitive,
				primitiveDest, transpileDest, nil)
			if nil != err {
				return treeStructureList, infoListList, usersStack, err
			}
			infoList := infoListList[0]
			_, variables, usersStack = ExecuteTree(infoList, variables, usersStack, toTranspile, toPrimitive,
				primitiveDest, transpileDest)

		}
	}

	maxNbraces = maxBraces(exprList)

	if maxNbraces > 0 || wasCd || wasAssignment || wasNOT || wasGoto || wasSetSource ||
		wasNextCommand || wasSendCommand || wasUndefine || wasPop || wasPush || wasSetDest || wasDelDest ||
		wasSendDest || wasPoint || wasLen || wasIndex || wasGetRootSource || wasGetRootDest ||
		wasIsLetter || wasIsDigit || wasExit || wasExists {
		if !wasCd {
			if !wasAssignment {
				//if not was_NOT:
				// безусловный print
				// меняем print под стандартную бинарную операцию
				i = 0
				if "print" == fmt.Sprintf("%v", exprList[i][1]) {
					exprList, err = makePrintBinary(exprList, variables, usersStack, showTree, toTranspile, toPrimitive,
						primitiveDest, transpileDest)
					if nil != err {
						return treeStructureList, infoListList, usersStack, err
					}

					i += 7
				} else if "input" == fmt.Sprintf("%v", exprList[i][1]) {
					exprList = makeOperationBinary(exprList)
					i += 7
				}

			}
		}

		nops := 0 // число операторов
		for _, el := range exprList {
			if "OP" == fmt.Sprintf("%v", el[0]) {
				nops++
			}
		}

		if nops > 0 { // разбираем на внутренние подвыражения
			for {
				maxNbraces = maxBraces(exprList)

				if maxNbraces < 1 {
					break
				}

				i = 1
				wasHere := false

				for i < len(exprList)-1 {
					if "OP" == fmt.Sprintf("%v", exprList[i][0]) && ("VAL" == fmt.Sprintf("%v", exprList[i-1][0]) ||
						"VAR" == fmt.Sprintf("%v", exprList[i-1][0])) &&
						("VAL" == fmt.Sprintf("%v", exprList[i+1][0]) ||
							"VAR" == fmt.Sprintf("%v", exprList[i+1][0])) {
						wasHere = true
						subExpr := []interface{}{exprList[i-1], exprList[i], exprList[i+1]}
						exprList = Pop(exprList, i-1) // выталкиваем отдельные части  подвыражения
						exprList = Pop(exprList, i-1)
						exprList = Pop(exprList, i-1)
						// помещаем целое выражение как подвыражение
						exprList = Insert(exprList, i-1, []interface{}{"VAL", "SUBEXPR", subExpr})
						exprList = Pop(exprList, i-2) // выталкиваем лишние скобки
						exprList = Pop(exprList, i-1)

					}
					i++
				}

				if !wasHere && "(" == fmt.Sprintf("%v", exprList[0][1]) &&
					")" == fmt.Sprintf("%v", exprList[len(exprList)-1][1]) {
					exprList = Pop(exprList, 0)
					exprList = Pop(exprList, len(exprList)-1)
				}

				i = 0
				// убираем скобки непосредственно рядом с подвыражениями
				for i < len(exprList) {
					if ("SUBEXPR" == fmt.Sprintf("%v", exprList[i][1])) &&
						(i-1 >= 0 && i+1 < len(exprList)) &&
						("(" == fmt.Sprintf("%v", exprList[i-1][1]) && ")" == fmt.Sprintf("%v", exprList[i+1][1])) {
						if (i-2 < 0) || ("OP" != exprList[i-2][0] || !IsUnaryOperation(fmt.Sprintf("%v", exprList[i-2][1]))) {
							exprList = Pop(exprList, i-1)
							exprList = Pop(exprList, i)
						}
					}
					i++
				}

			}

			var err error
			treeStructure, infoList, err = codeTree(exprList, "", nil)
			if nil != err {
				return treeStructureList, infoListList, usersStack, err
			}
		} else {
			treeStructure = "100"
			infoList = []interface{}{fmt.Sprintf("%v", exprList[0][1])}
		}
	} else {
		treeStructure = "100"
		infoList = []interface{}{fmt.Sprintf("%v", exprList[0][1])}
	}

	/*if showTree {
		drawModule.DrawTree(treeStructure, infoList)
	}*/

	treeStructureList = append(treeStructureList, treeStructure)
	infoListList = append(infoListList, infoList)

	return treeStructureList, infoListList, usersStack, sizeError
}
