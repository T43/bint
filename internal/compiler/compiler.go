package compiler

import (
	. "bint.com/internal/compilerVars"
	. "bint.com/pkg/serviceTools"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func CompileAsm(rootDest string) error {
	cmd := exec.Command("as", "--64", "asm/data.s", "asm/program.s", "-o", "asm/program.o")

	err := cmd.Start()
	if nil != err {
		return err
	}
	err = cmd.Wait()
	if nil != err {
		return err
	}

	cmd = exec.Command("ld", "-s", "asm/program.o", "-o", rootDest)

	err = cmd.Start()
	if nil != err {
		return err
	}
	err = cmd.Wait()
	if nil != err {
		return err
	}

	return nil
}

func InitData() (*os.File, error) {
	_, err := Copy("asm/pdata.s", "asm/data.s")
	if nil != err {
		fmt.Println("could not init file data.s")
		return nil, err
	}
	f, err := os.OpenFile("asm/data.s", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if nil != err {
		fmt.Println("could not create file data.s")
		return f, err
	}

	return f, nil
}

func InitProg() (*os.File, error) {
	_, err := Copy("asm/pprogram.s", "asm/program.s")
	if nil != err {
		fmt.Println("could not init file program.s")
		return nil, err
	}
	f, err := os.OpenFile("asm/program.s", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if nil != err {
		fmt.Println("could not create file program.s")
		return f, err
	}

	return f, nil
}

func FinishProg(f *os.File) error {
	_, err := f.Write([]byte("\nmov $60,  %rax"))
	if nil != err {
		return err
	}
	_, err = f.Write([]byte("\nxor %rdi, %rdi"))
	if nil != err {
		return err
	}
	_, err = f.Write([]byte("\nsyscall\n"))

	return err
}

func FinishData(f *os.File) error {
	_, err := f.Write([]byte("\n"))
	return err
}

func compile(systemStack []interface{}, OP string, LO []interface{}, RO []interface{}, dataFile *os.File, progFile *os.File,
	variables [][]interface{}) ([]interface{}, []interface{}, error) {
	if "print" == OP {
		return []interface{}{"print", LO}, systemStack, nil
	} else if "reg_find" == OP {
		if len(fmt.Sprintf("%v", LO[0])) >= 2 && `"` == string(fmt.Sprintf("%v", LO[0])[0]) &&
			`"` == string(fmt.Sprintf("%v", LO[0])[len(fmt.Sprintf("%v", LO[0]))-1]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		if len(fmt.Sprintf("%v", RO[0])) >= 2 && `"` == string(fmt.Sprintf("%v", RO[0])[0]) &&
			`"` == string(fmt.Sprintf("%v", RO[0])[len(fmt.Sprintf("%v", RO[0]))-1]) {
			RO[0] = RO[0].(string)[1 : len(RO[0].(string))-1]
		}

		pattern, err := regexp.Compile(fmt.Sprintf("%v", LO[0]))

		if nil != err {
			return []interface{}{-1}, systemStack, err
		}
		intListList := pattern.FindAllIndex([]byte(fmt.Sprintf("%v", RO[0])), -1)
		var res []interface{}
		res = append(res, []interface{}{"end"})

		for i := len(intListList) - 1; i >= 0; i-- {
			res = append(res, []interface{}{[]interface{}{"end"}})
			for j := len(intListList[i]) - 1; j >= 0; j-- {
				res[len(res)-1] = append(res[len(res)-1].([]interface{}), []interface{}{intListList[i][j]})
			}
		}

		return []interface{}{res}, systemStack, nil

	} else if "index" == OP {

		if len(fmt.Sprintf("%v", LO[0])) >= 2 && `"` == string(fmt.Sprintf("%v", LO[0])[0]) &&
			`"` == string(fmt.Sprintf("%v", LO[0])[len(fmt.Sprintf("%v", LO[0]))-1]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		if len(fmt.Sprintf("%v", RO[0])) >= 2 && `"` == string(fmt.Sprintf("%v", RO[0])[0]) &&
			`"` == string(fmt.Sprintf("%v", RO[0])[len(fmt.Sprintf("%v", RO[0]))-1]) {
			RO[0] = RO[0].(string)[1 : len(RO[0].(string))-1]
		}

		return []interface{}{strings.Index(fmt.Sprintf("%v", LO[0]), fmt.Sprintf("%v", RO[0]))}, systemStack, nil
	} else if "len" == OP {
		if 0 == len(fmt.Sprintf("%v", LO[0])) {
			return []interface{}{0}, systemStack, nil
		}
		if `"` == string(fmt.Sprintf("%v", LO[0])[0]) && `"` == string(fmt.Sprintf("%v", LO[0])[len(fmt.Sprintf("%v", LO[0]))-1]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}

		return []interface{}{len([]rune(fmt.Sprintf("%v", LO[0])))}, systemStack, nil
	} else if "exists" == OP {
		if `"` == string(fmt.Sprintf("%v", LO[0])[0]) && `"` == string(fmt.Sprintf("%v", LO[0])[len(fmt.Sprintf("%v", LO[0]))-1]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		return []interface{}{Exists(fmt.Sprintf("%v", LO[0]))}, systemStack, nil
	} else if "AND" == OP {
		if "bool" == WhatsType(fmt.Sprintf("%v", LO[0])) && "bool" == WhatsType(fmt.Sprintf("%v", RO[0])) {
			return []interface{}{StrToBool(fmt.Sprintf("%v", LO[0])) && StrToBool(fmt.Sprintf("%v", RO[0]))}, systemStack, nil
		}
		err := errors.New("executor: AND: error: data type mismatch")
		return LO, systemStack, err
	} else if "OR" == OP {
		if "bool" == WhatsType(fmt.Sprintf("%v", LO[0])) && "bool" == WhatsType(fmt.Sprintf("%v", RO[0])) {
			return []interface{}{StrToBool(fmt.Sprintf("%v", LO[0])) || StrToBool(fmt.Sprintf("%v", RO[0]))}, systemStack, nil
		}
		err := errors.New("executor: OR: error: data type mismatch")
		return LO, systemStack, err
	} else if "XOR" == OP {
		if "bool" == WhatsType(fmt.Sprintf("%v", LO[0])) && "bool" == WhatsType(fmt.Sprintf("%v", RO[0])) {
			return []interface{}{StrToBool(fmt.Sprintf("%v", LO[0])) != StrToBool(fmt.Sprintf("%v", RO[0]))}, systemStack, nil
		}
		err := errors.New("executor: XOR: error: data type mismatch")
		return LO, systemStack, err
	} else if "NOT" == OP {
		if "bool" == WhatsType(fmt.Sprintf("%v", LO[0])) {
			return []interface{}{!StrToBool(fmt.Sprintf("%v", LO[0]))}, systemStack, nil
		}
		err := errors.New("executor: NOT: error: data type mismatch")
		return LO, systemStack, err
	} else if "L: True" == OP || "L: true" == OP {
		return LO, systemStack, nil
	} else if "L: False" == OP || "L: false" == OP {
		return RO, systemStack, nil
	} else if "<" == OP {
		if ("int" != WhatsType(fmt.Sprintf("%v", LO[0])) && "float" != WhatsType(fmt.Sprintf("%v", LO[0]))) ||
			("int" != WhatsType(fmt.Sprintf("%v", RO[0])) && "float" != WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: < : error: data type mismatch")
			return LO, systemStack, err
		}

		floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)

		if nil != err {
			return LO, systemStack, err
		}

		floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)

		if nil != err {
			return RO, systemStack, err
		}

		return []interface{}{floatLO < floatRO}, systemStack, nil

	} else if "<=" == OP {
		if ("int" != WhatsType(fmt.Sprintf("%v", LO[0])) && "float" != WhatsType(fmt.Sprintf("%v", LO[0]))) ||
			("int" != WhatsType(fmt.Sprintf("%v", RO[0])) && "float" != WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: <= : error: data type mismatch")
			return LO, systemStack, err
		}

		floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)

		if nil != err {
			return LO, systemStack, err
		}

		floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)

		if nil != err {
			return RO, systemStack, err
		}

		return []interface{}{floatLO <= floatRO}, systemStack, nil

	} else if "==" == OP {
		LOType := WhatsType(fmt.Sprintf("%v", LO[0]))
		ROType := WhatsType(fmt.Sprintf("%v", RO[0]))

		if "string" == LOType {
			ROType = "string"
		} else if "string" == ROType {
			LOType = "string"
		}

		if LOType != ROType {
			err := errors.New("executor: == : ERROR: data type mismatch")
			return LO, systemStack, err
		}

		if "int" == LOType {
			intLO, err := strconv.Atoi(fmt.Sprintf("%v", LO[0]))
			if nil != err {
				return LO, systemStack, err
			}

			intRO, err := strconv.Atoi(fmt.Sprintf("%v", RO[0]))
			if nil != err {
				return LO, systemStack, err
			}

			return []interface{}{BoolToStr(intLO == intRO)}, systemStack, nil
		}

		if "float" == LOType {
			floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)
			if nil != err {
				return LO, systemStack, err
			}

			floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)
			if nil != err {
				return LO, systemStack, err
			}

			return []interface{}{BoolToStr(floatLO == floatRO)}, systemStack, nil
		}

		if "string" == LOType {
			if len(fmt.Sprintf("%v", LO[0])) > 0 && "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
				LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
			}
			if len(fmt.Sprintf("%v", RO[0])) > 0 && "\"" == string(fmt.Sprintf("%v", RO[0])[0]) {
				RO[0] = RO[0].(string)[1 : len(RO[0].(string))-1]
			}

			if "\\n" == fmt.Sprintf("%v", LO[0]) {
				LO[0] = "\n"
			}

			if "\\n" == fmt.Sprintf("%v", RO[0]) {
				RO[0] = "\n"
			}

			if "\\t" == fmt.Sprintf("%v", LO[0]) {
				LO[0] = "\t"
			}

			if "\\t" == fmt.Sprintf("%v", RO[0]) {
				RO[0] = "\t"
			}

			return []interface{}{BoolToStr(fmt.Sprintf("%v", LO[0]) == fmt.Sprintf("%v", RO[0]))}, systemStack, nil
		}

		err := errors.New("executor: == : ERROR: data type mismatch")
		return LO, systemStack, err
	} else if ">" == OP {
		if ("int" != WhatsType(fmt.Sprintf("%v", LO[0])) && "float" != WhatsType(fmt.Sprintf("%v", LO[0]))) ||
			("int" != WhatsType(fmt.Sprintf("%v", RO[0])) && "float" != WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: > : error: data type mismatch")
			return LO, systemStack, err
		}

		floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)

		if nil != err {
			return LO, systemStack, err
		}

		floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)

		if nil != err {
			return LO, systemStack, err
		}

		return []interface{}{BoolToStr(floatLO > floatRO)}, systemStack, nil
	} else if ">=" == OP {
		if ("int" != WhatsType(fmt.Sprintf("%v", LO[0])) && "float" != WhatsType(fmt.Sprintf("%v", LO[0]))) ||
			("int" != WhatsType(fmt.Sprintf("%v", RO[0])) && "float" != WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: >= : error: data type mismatch")
			return LO, systemStack, err
		}

		floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)

		if nil != err {
			return LO, systemStack, err
		}

		floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)

		if nil != err {
			return LO, systemStack, err
		}

		return []interface{}{BoolToStr(floatLO >= floatRO)}, systemStack, nil
	} else if "+" == OP {
		var typeLO string
		var typeRO string
		var lenLO string
		var lenRO string
		isVarLO := false
		isVarRO := false

		newVariable := EachVariable(variables)

		for v := newVariable(); "end" != fmt.Sprintf("%v", v[0]); v = newVariable() {
			if fmt.Sprintf("%v", LO[0]) == fmt.Sprintf("%v", v[1]) {
				isVarLO = true
				typeLO = fmt.Sprintf("%v", v[0])
				LO[0] = "varName" + fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", LO[0])])
				lenLO = "lenVarName" + fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", LO[0])])
			}
			if fmt.Sprintf("%v", RO[0]) == fmt.Sprintf("%v", v[1]) {
				isVarRO = true
				typeRO = fmt.Sprintf("%v", v[0])
				RO[0] = "varName" + fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", RO[0])])
				lenRO = "lenVarName" + fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", RO[0])])
			}
		}

		if "" == typeLO && nil != LO[0] {
			typeLO = WhatsType(fmt.Sprintf("%v", LO[0]))
		}
		if "" == typeRO && nil != RO[0] {
			typeRO = WhatsType(fmt.Sprintf("%v", RO[0]))
		}

		if ("int" == typeLO || nil == LO[0]) && ("int" == typeRO || nil == RO[0]) {
			if isVarLO && isVarRO {
				// lastUserData = userData
				_, err := progFile.Write([]byte("\nmov $lenLastUserData, %rsi \n mov $lastUserData, %rdx \n mov $lenUserData, %rax \n" +
					"mov $userData, %rdi \n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// присвоить значение переменной в buf3
				_, err = progFile.Write([]byte("\nmov $lenVarName, %rsi \n mov $varName, %rdx \n" +
					"mov $" + lenLO + ", %rax \n mov $" + fmt.Sprintf("%v", LO[0]) + ", %rdi " +
					"\n call __set\n call __getVar \n mov $lenBuf3, %rsi \n mov $buf3, %rdx \n mov $lenUserData," +
					"%rax \n mov (userData), %rdi \n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// присвоить значение переменной в buf4
				_, err = progFile.Write([]byte("\nmov $lenVarName, %rsi \n mov $varName, %rdx \n" +
					"mov $" + lenRO + ", %rax \n mov $" + fmt.Sprintf("%v", RO[0]) + ", %rdi " +
					"\n call __set\n call __getVar \n mov $lenBuf4, %rsi \n mov $buf4, %rdx \n mov $lenUserData," +
					"%rax \n mov (userData), %rdi \n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// buf = buf3
				// buf2 = buf4
				// rax = 0
				// add
				_, err = progFile.Write([]byte("\nmov $lenBuf, %rsi \n mov $buf, %rdx \n mov $lenBuf3, %rax \n" +
					"mov $buf3, %rdi \n call __set\n mov $lenBuf2, %rsi \n mov $buf2, %rdx \n mov $lenBuf4, %rax \n" +
					"mov $buf4, %rdi \n call __set\n\n mov $0, %rax \n call __add "))

				return []interface{}{nil}, systemStack, nil // успех, результат по адресу $userData
			} else if nil == LO[0] && isVarRO {
				// LO[0] по адресу $userData
				// присвоить значение переменной в buf3
				_, err := progFile.Write([]byte("\nmov $lenBuf3, %rsi \n mov $buf3, %rdx \n" +
					"mov $lenUserData, %rax \n mov $userData, %rdi " +
					"\n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// присвоить значение переменной в buf4
				_, err = progFile.Write([]byte("\nmov $lenVarName, %rsi \n mov $varName, %rdx \n" +
					"mov $" + lenRO + ", %rax \n mov $" + fmt.Sprintf("%v", RO[0]) + ", %rdi " +
					"\n call __set\n call __getVar \n mov $lenBuf4, %rsi \n mov $buf4, %rdx \n mov $lenUserData," +
					"%rax \n mov (userData), %rdi \n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// buf = buf3
				// buf2 = buf4
				// rax = 0
				// add
				_, err = progFile.Write([]byte("\nmov $lenBuf, %rsi \n mov $buf, %rdx \n mov $lenBuf3, %rax \n" +
					"mov $buf3, %rdi \n call __set\n mov $lenBuf2, %rsi \n mov $buf2, %rdx \n mov $lenBuf4, %rax \n" +
					"mov $buf4, %rdi \n call __set\n\n mov $0, %rax \n call __add "))

				return []interface{}{nil}, systemStack, nil // успех, результат по адресу $userData
			} else if nil == RO[0] && isVarLO {
				// RO[0] по адресу $userData
				// присвоить значение переменной в buf3
				_, err := progFile.Write([]byte("\nmov $lenBuf3, %rsi \n mov $buf3, %rdx \n" +
					"mov $lenUserData, %rax \n mov $userData, %rdi " +
					"\n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// присвоить значение переменной в buf4
				_, err = progFile.Write([]byte("\nmov $lenVarName, %rsi \n mov $varName, %rdx \n" +
					"mov $" + lenLO + ", %rax \n mov $" + fmt.Sprintf("%v", LO[0]) + ", %rdi " +
					"\n call __set\n call __getVar \n mov $lenBuf4, %rsi \n mov $buf4, %rdx \n mov $lenUserData," +
					"%rax \n mov (userData), %rdi \n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// buf = buf3
				// buf2 = buf4
				// rax = 0
				// add
				_, err = progFile.Write([]byte("\nmov $lenBuf, %rsi \n mov $buf, %rdx \n mov $lenBuf3, %rax \n" +
					"mov $buf3, %rdi \n call __set\n mov $lenBuf2, %rsi \n mov $buf2, %rdx \n mov $lenBuf4, %rax \n" +
					"mov $buf4, %rdi \n call __set\n\n mov $0, %rax \n call __add "))

				return []interface{}{nil}, systemStack, nil // успех, результат по адресу $userData
			} else if nil == LO[0] && nil == RO[0] {
				// LO[0] в lastUserData
				// RO[0] в userData
				// присвоить значение переменной в buf3
				_, err := progFile.Write([]byte("\nmov $lenBuf3, %rsi \n mov $buf3, %rdx \n" +
					"mov $lenLastUserData, %rax \n mov $lastUserData, %rdi " +
					"\n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// присвоить значение переменной в buf4
				_, err = progFile.Write([]byte("\nmov $lenBuf4, %rsi \n mov $buf4, %rdx \n" +
					"mov $lenUserData, %rax \n mov $userData, %rdi " +
					"\n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// buf = buf3
				// buf2 = buf4
				// rax = 0
				// add
				_, err = progFile.Write([]byte("\nmov $lenBuf, %rsi \n mov $buf, %rdx \n mov $lenBuf3, %rax \n" +
					"mov $buf3, %rdi \n call __set\n mov $lenBuf2, %rsi \n mov $buf2, %rdx \n mov $lenBuf4, %rax \n" +
					"mov $buf4, %rdi \n call __set\n\n mov $0, %rax \n call __add "))

				return []interface{}{nil}, systemStack, nil // успех, результат по адресу $userData
			} else {
				panic("unrealized case")
			}

		}

		if "float" == typeLO && "float" == typeRO {
			//
		}

		if "int" == typeLO && "float" == typeRO {
			//
		}
		if "float" == typeLO && "int" == typeRO {
			//
		}
		if len(fmt.Sprintf("%v", LO[0])) >= 2 && "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		if len(fmt.Sprintf("%v", RO[0])) >= 2 && "\"" == string(fmt.Sprintf("%v", RO[0])[0]) {
			RO[0] = RO[0].(string)[1 : len(RO[0].(string))-1]
		}

		return []interface{}{"\"" + fmt.Sprintf("%v", LO[0]) + fmt.Sprintf("%v", RO[0]) + "\""}, systemStack, nil

	} else if "-" == OP {
		if ("int" != WhatsType(fmt.Sprintf("%v", LO[0])) && "float" != WhatsType(fmt.Sprintf("%v", LO[0]))) ||
			("int" != WhatsType(fmt.Sprintf("%v", RO[0])) && "float" != WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: - : error: data type mismatch")
			return LO, systemStack, err
		}

		if "int" == WhatsType(fmt.Sprintf("%v", LO[0])) {
			intLO, err := strconv.Atoi(fmt.Sprintf("%v", LO[0]))
			if nil != err {
				return LO, systemStack, err
			}

			if "int" == WhatsType(fmt.Sprintf("%v", RO[0])) {
				intRO, err := strconv.Atoi(fmt.Sprintf("%v", RO[0]))
				if nil != err {
					return LO, systemStack, err
				}

				return []interface{}{intLO - intRO}, systemStack, nil
			}

			if "float" == WhatsType(fmt.Sprintf("%v", RO[0])) {
				floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)
				if nil != err {
					return LO, systemStack, err
				}
				floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)
				if nil != err {
					return LO, systemStack, err
				}

				return []interface{}{floatLO - floatRO}, systemStack, nil
			}
			err = errors.New("executor: - : error: data type mismatch")
			return LO, systemStack, err

		}

		if "float" == WhatsType(fmt.Sprintf("%v", LO[0])) {
			floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)

			if nil != err {
				return LO, systemStack, err
			}

			if "float" == WhatsType(fmt.Sprintf("%v", RO[0])) || "int" == WhatsType(fmt.Sprintf("%v", RO[0])) {
				floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)
				if nil != err {
					return LO, systemStack, err
				}

				return []interface{}{floatLO - floatRO}, systemStack, nil
			}

			err = errors.New("executor: - : error: data type mismatch")
			return LO, systemStack, err
		}
	} else if "*" == OP {
		if ("int" != WhatsType(fmt.Sprintf("%v", LO[0])) && "float" != WhatsType(fmt.Sprintf("%v", LO[0]))) ||
			("int" != WhatsType(fmt.Sprintf("%v", RO[0])) && "float" != WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: * : error: data type mismatch")
			return LO, systemStack, err
		}

		if "int" == WhatsType(fmt.Sprintf("%v", LO[0])) {
			intLO, err := strconv.Atoi(fmt.Sprintf("%v", LO[0]))
			if nil != err {
				return LO, systemStack, err
			}

			if "int" == WhatsType(fmt.Sprintf("%v", RO[0])) {
				intRO, err := strconv.Atoi(fmt.Sprintf("%v", RO[0]))
				if nil != err {
					return LO, systemStack, err
				}

				return []interface{}{intLO * intRO}, systemStack, nil
			}

			if "float" == WhatsType(fmt.Sprintf("%v", RO[0])) {
				floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)
				if nil != err {
					return LO, systemStack, err
				}
				floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)
				if nil != err {
					return LO, systemStack, err
				}

				return []interface{}{floatLO * floatRO}, systemStack, nil
			}

			err = errors.New("executor: * : error: data type mismatch")
			return LO, systemStack, err

		}

		if "float" == WhatsType(fmt.Sprintf("%v", LO[0])) {
			floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)

			if nil != err {
				return LO, systemStack, err
			}

			if "float" == WhatsType(fmt.Sprintf("%v", RO[0])) || "int" == WhatsType(fmt.Sprintf("%v", RO[0])) {
				floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)
				if nil != err {
					return LO, systemStack, err
				}

				return []interface{}{floatLO * floatRO}, systemStack, nil
			}

			err = errors.New("executor: * : error: data type mismatch")
			return LO, systemStack, err
		}
	} else if "/" == OP {
		if ("int" != WhatsType(fmt.Sprintf("%v", LO[0])) && "float" != WhatsType(fmt.Sprintf("%v", LO[0]))) ||
			("int" != WhatsType(fmt.Sprintf("%v", RO[0])) && "float" != WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: / : error: data type mismatch")
			return LO, systemStack, err
		}

		floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)
		if nil != err {
			return LO, systemStack, err
		}
		floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)
		if nil != err {
			return RO, systemStack, err
		}

		return []interface{}{floatLO / floatRO}, systemStack, nil

	} else if "@" == OP {
		if !("int" == WhatsType(fmt.Sprintf("%v", LO[0])) && "int" == WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: @ : error: data type mismatch")
			return LO, systemStack, err
		}
		intLO, err := strconv.Atoi(fmt.Sprintf("%v", LO[0]))
		if nil != err {
			return LO, systemStack, err
		}
		intRO, err := strconv.Atoi(fmt.Sprintf("%v", RO[0]))
		if nil != err {
			return LO, systemStack, err
		}
		return []interface{}{intLO / intRO}, systemStack, nil
	} else if "^" == OP {
		if ("int" != WhatsType(fmt.Sprintf("%v", LO[0])) && "float" != WhatsType(fmt.Sprintf("%v", LO[0]))) ||
			("int" != WhatsType(fmt.Sprintf("%v", RO[0])) && "float" != WhatsType(fmt.Sprintf("%v", RO[0]))) {
			err := errors.New("executor: ^ : error: data type mismatch")
			return LO, systemStack, err
		}

		floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)
		if nil != err {
			return LO, systemStack, err
		}

		floatRO, err := strconv.ParseFloat(fmt.Sprintf("%v", RO[0]), 64)
		if nil != err {
			return LO, systemStack, err
		}

		return []interface{}{math.Pow(floatLO, floatRO)}, systemStack, nil
	} else if "str" == OP {
		return []interface{}{fmt.Sprintf("%v", LO[0])}, systemStack, nil
	} else if "=" == OP {
		var wasVar bool

		if nil != RO[0] {
			newVariable := EachVariable(variables)
			for v := newVariable(); "end" != v[0]; v = newVariable() {
				if fmt.Sprintf("%v", ValueFoldInterface(RO[0])) == fmt.Sprintf("%v", v[1]) {
					wasVar = true
					// getVar, результат в userData
					numberS := fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", ValueFoldInterface(v[1]))])
					_, err := progFile.Write([]byte("\nmov $lenVarName, %rsi \n mov $varName, %rdx" +
						"\n mov $lenVarName" + numberS +
						", %rax \n mov $varName" + numberS + ", %rdi\n call __set\n call __getVar"))
					if nil != err {
						fmt.Println(err)
						os.Exit(1)
					}
					numberS = fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", ValueFoldInterface(LO[0]))])
					_, err = progFile.Write([]byte("\nmov $lenVarName, %rsi \n mov $varName, %rdx" +
						"\n mov $lenVarName" + numberS + ", %rax \n mov $varName" + numberS + ", %rdi\n call __set \n call __setVar"))
					if nil != err {
						fmt.Println(err)
						os.Exit(1)
					}
				}
			}
		}
		if !wasVar {
			if nil == RO[0] {
				// справа разультат вычислений, находящийся по адресу $userData
				varName := "varName" + fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", LO[0])])
				lenVarName := "lenVarName" + fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", LO[0])])

				// присвоить в varName имя текущей переменной
				_, err := progFile.Write([]byte("\nmov $lenVarName, %rsi \n mov $varName, %rdx \n mov $" + lenVarName +
					", %rax \n mov $" + varName + ", %rdi \n call __set"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				// buf3 = userData
				// (userData) = $buf3
				// call __setVar
				_, err = progFile.Write([]byte("\nmov $lenBuf3, %rsi \n mov $buf3, %rdx" +
					"\n mov $lenUserData, %rax \n mov $userData, %rdi \n call __set \n mov $buf3, %rax \n mov %rax, (userData) \n" +
					"call __setVar"))

				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
			} else {
				// справа данные
				_, err := dataFile.Write([]byte("\ndata" + fmt.Sprintf("%v", DataNumber) + ":\n"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				_, err = dataFile.Write([]byte(".ascii \"" + fmt.Sprintf("%v", RO[0]) + "\"\n.space 1, 0"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				_, err = dataFile.Write([]byte("\nlenData" + fmt.Sprintf("%v", DataNumber) + " = . - data" + fmt.Sprintf("%v", DataNumber)))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				_, err = progFile.Write([]byte("\n mov $lenVarName, %rsi \n mov $varName, %rdx " +
					"\n mov $lenVarName" + fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", LO[0])]) +
					", %rax \n mov $varName" + fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", LO[0])]) + ", %rdi \n" +
					"call __set\n\n mov $data" + fmt.Sprintf("%v", DataNumber) + ", %rax  \n mov %rax, (userData)\n call __setVar"))
				if nil != err {
					fmt.Println(err)
					os.Exit(1)
				}
				DataNumber++
			}
		}
		return []interface{}{0}, systemStack, nil // успех
	} else if "." == OP {
		return []interface{}{0}, systemStack, nil
	} else if "UNDEFINE" == OP {
		return []interface{}{0}, systemStack, nil
	} else if "int" == OP {
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		if "float" == WhatsType(fmt.Sprintf("%v", LO[0])) {
			floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)
			if nil != err {
				return LO, systemStack, err
			}

			return []interface{}{int(floatLO)}, systemStack, nil
		}

		intLO, err := strconv.Atoi(fmt.Sprintf("%v", LO[0]))

		if nil != err {
			return LO, systemStack, err
		}

		return []interface{}{intLO}, systemStack, nil
	} else if "float" == OP {
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}

		floatLO, err := strconv.ParseFloat(fmt.Sprintf("%v", LO[0]), 64)

		return []interface{}{floatLO}, systemStack, err
	} else if "bool" == OP {
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}

		return LO, systemStack, nil
	} else if "input" == OP {
		var s string
		_, err := fmt.Scan(&s)
		return []interface{}{s}, systemStack, err
	} else if "goto" == OP {
		if "string" != WhatsType(fmt.Sprintf("%v", LO[0])) {
			err := errors.New("executor: goto : error: data type mismatch")
			return LO, systemStack, err
		}
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		if "#" != string(fmt.Sprintf("%v", LO[0])[0]) {
			err := errors.New("executor: goto: ERROR: mark must start with \"#\", mark: " + fmt.Sprintf("%v", LO[0]))
			return LO, systemStack, err
		}

		return []interface{}{"goto", LO[0]}, systemStack, nil
	} else if "exit" == OP {
		if "int" != WhatsType(fmt.Sprintf("%v", LO[0])) {
			err := errors.New("executor: exit: error: data type mismatch")
			return LO, systemStack, err
		}
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		code, err := strconv.Atoi(fmt.Sprintf("%v", LO[0]))
		if nil != err {
			panic(err)
		}
		_, err = progFile.Write([]byte("mov $60,   %rax\n"))
		if nil != err {
			fmt.Println(err)
			os.Exit(1)
		}
		_, err = progFile.Write([]byte("mov $" + fmt.Sprintf("%v", code) + ", %rdi\n"))
		if nil != err {
			fmt.Println(err)
			os.Exit(1)
		}

		_, err = progFile.Write([]byte("syscall\n"))
		if nil != err {
			fmt.Println(err)
			os.Exit(1)
		}
		return []interface{}{0}, systemStack, nil
	} else if "is_letter" == OP {
		if "string" != WhatsType(fmt.Sprintf("%v", LO[0])) {
			err := errors.New("executor: is_letter : error: data type mismatch")
			return LO, systemStack, err
		}
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		if 1 != len(fmt.Sprintf("%v", LO[0])) {
			err := errors.New("executor: is_letter : error: length of the argument is more than 1, argument: " +
				fmt.Sprintf("%v", LO[0]))

			return LO, systemStack, err
		}
		return []interface{}{unicode.IsLetter([]rune(fmt.Sprintf("%v", LO[0]))[0])}, systemStack, nil
	} else if "is_digit" == OP {
		if "string" != WhatsType(fmt.Sprintf("%v", LO[0])) {
			err := errors.New("executor: is_digit : error: data type mismatch")
			return LO, systemStack, err
		}
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		if 1 != len(fmt.Sprintf("%v", LO[0])) {
			err := errors.New("executor: is_digit : error: length of the argument is more than 1, argument: " +
				fmt.Sprintf("%v", LO[0]))

			return LO, systemStack, err
		}
		return []interface{}{unicode.IsDigit([]rune(fmt.Sprintf("%v", LO[0]))[0])}, systemStack, nil
	} else if "SET_SOURCE" == OP {
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		f, err := os.Open(fmt.Sprintf("%v", LO[0]))
		return []interface{}{"SET_SOURCE", f}, systemStack, err
	} else if "SET_DEST" == OP {
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		f, err := os.Create(fmt.Sprintf("%v", LO[0]))
		return []interface{}{"SET_DEST", f}, systemStack, err
	} else if "SEND_DEST" == OP {
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		fin, err := os.Open("benv/program.basm")
		if nil != err {
			return []interface{}{1}, systemStack, err
		}

		fout, err := os.Create(fmt.Sprintf("%v", LO[0]))

		if nil != err {
			return []interface{}{1}, systemStack, err
		}
		_, err = io.Copy(fout, fin)

		if nil != err {
			return []interface{}{1}, systemStack, err
		}

		err = fin.Close()
		if nil != err {
			return []interface{}{1}, systemStack, err
		}
		err = fout.Close()
		if nil != err {
			return []interface{}{1}, systemStack, err
		}

		err = os.Remove("benv/program.basm")
		if nil != err {
			return []interface{}{1}, systemStack, err
		}

		return []interface{}{0}, systemStack, nil
	} else if "DEL_DEST" == OP {
		if "\"" == string(fmt.Sprintf("%v", LO[0])[0]) {
			LO[0] = LO[0].(string)[1 : len(LO[0].(string))-1]
		}
		if Exists(fmt.Sprintf("%v", LO[0])) {
			//err := LO[0].(*os.File).Close()
			//if nil != err{
			//	return []interface{}{1}, systemStack, err
			//}
			err := os.Remove(fmt.Sprintf("%v", LO[0]))
			if nil != err {
				return []interface{}{1}, systemStack, err
			}
		}

		return []interface{}{0}, systemStack, nil
	} else if "REROUTE" == OP {
		return []interface{}{"REROUTE", 0}, systemStack, nil
	} else if "UNSET_SOURCE" == OP {
		return []interface{}{"UNSET_SOURCE", 0}, systemStack, nil
	} else if "UNSET_DEST" == OP {
		return []interface{}{"UNSET_DEST", 0}, systemStack, nil
	} else if "RESET_SOURCE" == OP {
		return []interface{}{"RESET_SOURCE", 0}, systemStack, nil
	} else if "next_command" == OP {
		return []interface{}{"next_command", LO}, systemStack, nil
	} else if "get_root_source" == OP {
		return []interface{}{"get_root_source", LO}, systemStack, nil
	} else if "get_root_dest" == OP {
		return []interface{}{"get_root_dest", LO}, systemStack, nil
	} else if "send_command" == OP {
		return []interface{}{"send_command", LO}, systemStack, nil
	} else if "push" == OP {
		systemStack = append(systemStack, LO)
		return []interface{}{0}, systemStack, nil
	} else if "pop" == OP {
		res := systemStack[len(systemStack)-1]
		if "end" != res {
			systemStack = systemStack[:len(systemStack)-1]
		}
		if "string" == fmt.Sprintf("%T", ValueFoldInterface(res)) {
			return []interface{}{ValueFoldInterface(res)}, systemStack, nil
		} else {
			return ValueFoldInterface(res).([]interface{}), systemStack, nil
		}
	}
	err := errors.New("execute: ERROR: wrong syntax: OP=\"" + OP + "\", " +
		"LO=\"" + fmt.Sprintf("%v", LO) + "\", " + "RO=\"" + fmt.Sprintf("%v", RO) + "\"")
	return []interface{}{1}, systemStack, err

}

func sysCompileTree(infoList []interface{}, variables [][]interface{}, systemStack []interface{},
	OPPointer int, dataFile *os.File, progFile *os.File) ([]interface{}, [][]interface{}, []interface{}, int) {
	// заканчивает свою работу, когда выполнен первый оператор
	OP := fmt.Sprintf("%v", infoList[OPPointer])

	var LO []interface{}
	var RO []interface{}

	if 1 == len(infoList) && ("UNSET_SOURCE" == infoList[0] || "RESET_SOURCE" == infoList[0] ||
		"UNSET_DEST" == infoList[0] || "REROUTE" == infoList[0]) {
		infoList = append(infoList, "null")
		infoList = append(infoList, "null")
	}

	if "True" == infoList[OPPointer+1] {
		LO = []interface{}{"True"}
		OPPointer += 1
	} else if "False" == infoList[OPPointer+1] {
		LO = []interface{}{"False"}
		OPPointer += 1
	} else if !IsOp(fmt.Sprintf("%v", infoList[OPPointer+1])) {
		LO = []interface{}{infoList[OPPointer+1]}
		OPPointer += 1
	} else {
		// операция
		OPPointer += 1
		LO, variables, _, OPPointer = sysCompileTree(infoList, variables, systemStack, OPPointer, dataFile, progFile)
	}

	if "True" == infoList[OPPointer+1] {
		RO = []interface{}{"True"}
		OPPointer += 1
	} else if "False" == infoList[OPPointer+1] {
		RO = []interface{}{"False"}
		OPPointer += 1
	} else if !IsOp(fmt.Sprintf("%v", infoList[OPPointer+1])) {
		RO = []interface{}{infoList[OPPointer+1]}
		OPPointer += 1
	} else {
		// операция
		OPPointer += 1
		RO, variables, _, OPPointer = sysCompileTree(infoList, variables, systemStack, OPPointer, dataFile, progFile)
	}

	if "UNDEFINE" == OP {
		i := len(variables) - 1
		newVariable := EachVariable(variables)
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if fmt.Sprintf("%v", v[1]) == fmt.Sprintf("%v", LO[0]) {
				copy(variables[i:], variables[i+1:])
				variables = variables[:len(variables)-1]
				break
			}
			i -= 1
		}
	}

	if "input" == OP {
		newVariable := EachVariable(variables)
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if fmt.Sprintf("%v", v[1]) == fmt.Sprintf("%v", LO[0]) {
				var err error
				if "string" != v[0] {
					err = errors.New("executor: input: ERROR: data type mismatch")
					panic(err)
				}
				v[2], systemStack, err = compile(systemStack, OP, LO, RO, dataFile, progFile, variables)
				if nil != err {
					panic(err)
				}
				break
			}
		}
	}

	if "next_command" == OP || "get_root_source" == OP || "get_root_dest" == OP {
		newVariable := EachVariable(variables)
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if fmt.Sprintf("%v", v[1]) == fmt.Sprintf("%v", LO[0]) {

				v[2] = v[1]
				break
			}
		}
	}

	if "pop" == OP {
		newVariable := EachVariable(variables)
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if fmt.Sprintf("%v", v[1]) == fmt.Sprintf("%v", LO[0]) {

				var err error

				v[2], systemStack, err = compile(systemStack, OP, LO, RO, dataFile, progFile, variables)

				if nil != err {
					panic(err)
				}

				tempV2 := fmt.Sprintf("%v", ValueFoldInterface(v[2]))
				if "\"" == string(tempV2[0]) && "\"" == string(tempV2[len(tempV2)-1]) {
					tempV2 = tempV2[1 : len(tempV2)-1]
				}

				if fmt.Sprintf("%v", v[0]) != WhatsType(tempV2) && "string" != fmt.Sprintf("%v", v[0]) &&
					(fmt.Sprintf("%v", v[0]) != "float" && WhatsType(tempV2) != "int") &&
					!("stack" == fmt.Sprintf("%v", v[0]) && "[]interface {}" == fmt.Sprintf("%T", v[2]) &&
						"end" == ValueFoldInterface(v[2].([]interface{})[0])) {
					panic("pop: data type mismatch: " + fmt.Sprintf("%v", v[0]) + " and " +
						WhatsType(tempV2))
				}

				break
			}
		}
	}

	if "=" == OP {
		newVariable := EachVariable(variables)
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if fmt.Sprintf("%v", LO[0]) == fmt.Sprintf("%v", v[1]) {
				newRightVar := EachVariable(variables)
				for rightVar := newRightVar(); "end" != rightVar[0]; rightVar = newRightVar() {
					if fmt.Sprintf("%v", RO[0]) == fmt.Sprintf("%v", rightVar[1]) {
						if "[]interface {}" == fmt.Sprintf("%T", rightVar[2]) &&
							"string" == fmt.Sprintf("%T", rightVar[2].([]interface{})[0]) &&
							"end" == fmt.Sprintf("%v", rightVar[2].([]interface{})[0]) {
							rightVar[2].([]interface{})[0] = []interface{}{"end"}
						}
						RO[0] = fmt.Sprintf("%v", rightVar[1])
						break
					}
				}

				if nil != RO[0] {
					v[2] = RO
					break
				}
			}
		}

	}

	if "." == OP {
		newVariable := EachVariable(variables)
		breakFlag := false
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if breakFlag {
				break
			}
			if fmt.Sprintf("%v", LO[0]) == fmt.Sprintf("%v", v[1]) {
				typeLO := fmt.Sprintf("%v", v[0])

				if "stack" != typeLO {
					err := errors.New("executor: ERROR: the left operand of the operation \".\" must be of type stack")
					panic(err)
				}

				if "pop" != fmt.Sprintf("%v", RO[0].([]string)[0]) {
					newRightVar := EachVariable(variables)
					for rightVar := newRightVar(); "end" != fmt.Sprintf("%v", rightVar[0]); rightVar = newRightVar() {
						if fmt.Sprintf("%v", RO[0].([]string)[1]) == fmt.Sprintf("%v", rightVar[1]) {

							if "string" == fmt.Sprintf("%T", ValueFoldInterface(rightVar[2])) &&
								"stack" != fmt.Sprintf("%v", rightVar[0]) {

								RO[0].([]string)[1] = fmt.Sprintf("%v", rightVar[2].([]interface{})[0])
							} else {
								RO = nil
								RO = append(RO, []string{"push"})
								RO = append(RO, rightVar)
							}
							break

						}
					}
				}

				if "push" == fmt.Sprintf("%v", RO[0].([]string)[0]) {

					if 1 == len(RO) {
						v[2] = append(v[2].([]interface{}), []interface{}{RO[0].([]string)[1]})
					} else {
						v[2] = append(v[2].([]interface{}), RO[1].([]interface{})[2])
					}
					break
				} else if "pop" == fmt.Sprintf("%v", RO[0].([]string)[0]) {
					newPopVariable := EachVariable(variables)
					for popVar := newPopVariable(); "end" != fmt.Sprintf("%v", popVar[0]); popVar = newPopVariable() {
						if popVar[1] == RO[0].([]string)[1] {

							if "end" == fmt.Sprintf("%v", v[2].([]interface{})[0]) {
								v[2].([]interface{})[0] = []interface{}{"end"}
							}
							VFI := ValueFoldInterface(v[2])
							if "string" == fmt.Sprintf("%T", VFI) {
								VFI = []interface{}{VFI}
							}
							popVar[2] = VFI.([]interface{})[len(VFI.([]interface{}))-1]

							tempPopVar2 := fmt.Sprintf("%v", ValueFoldInterface(popVar[2]))
							if "\"" == string(tempPopVar2[0]) && "\"" == string(tempPopVar2[len(tempPopVar2)-1]) {
								tempPopVar2 = tempPopVar2[1 : len(tempPopVar2)-1]
							}

							if fmt.Sprintf("%v", popVar[0]) != WhatsType(tempPopVar2) && "string" != fmt.Sprintf("%v", popVar[0]) &&
								(fmt.Sprintf("%v", popVar[0]) != "float" && WhatsType(tempPopVar2) != "int") && !("stack" == fmt.Sprintf("%v", popVar[0]) &&
								"[]interface {}" == fmt.Sprintf("%T", popVar[2]) &&
								"end" == ValueFoldInterface(popVar[2].([]interface{})[0])) {
								if !("stack" == fmt.Sprintf("%v", popVar[0]) && "string" == WhatsType(fmt.Sprintf("%v", tempPopVar2)) &&
									"end" == tempPopVar2) {
									panic("pop: data type mismatch: " + fmt.Sprintf("%v", popVar[0]) + " and " +
										WhatsType(fmt.Sprintf("%v", tempPopVar2)))
								} else {
									VFI = []interface{}{[]interface{}{"end"}}
									popVar[2] = []interface{}{[]interface{}{"end"}}

								}
							}

							v[2] = VFI

							if len(VFI.([]interface{})) > 1 {
								v[2] = v[2].([]interface{})[:len(v[2].([]interface{}))-1]
							}
							breakFlag = true
							break
						}
					}
				} else {
					err := errors.New("executor: ERROR: after the operation \".\" must follow \"push\" or \"pop\"")
					panic(err)
				}
			}
		}
	}

	if "goto" == OP {
		newVariable := EachVariable(variables)
		for v := newVariable(); "end" != fmt.Sprintf("%v", v[0]); v = newVariable() {
			if fmt.Sprintf("%v", LO[0]) == fmt.Sprintf("%v", v[1]) {
				LO = v[2].([]interface{})
			}

		}
	}

	var res []interface{}
	if "input" != OP && "pop" != OP {
		var err error
		var passLO []interface{}
		var passRO []interface{}

		for _, el := range LO {
			passLO = append(passLO, el)
		}
		for _, el := range RO {
			passRO = append(passRO, el)
		}

		res, systemStack, err = compile(systemStack, OP, passLO, passRO, dataFile, progFile, variables)
		if nil != err {
			panic(err)
		}

	} else {
		res = []interface{}{0}
	}

	return res, variables, systemStack, OPPointer
}

func CompileTree(infoList []interface{}, variables [][]interface{},
	systemStack []interface{}, dataFile *os.File, progFile *os.File) {

	res, variables, systemStack, _ := sysCompileTree(infoList, variables, systemStack, 0, dataFile, progFile)

	if "print" == res[0] {
		var wasVar bool
		newVariable := EachVariable(variables)
		for v := newVariable(); "end" != v[0]; v = newVariable() {
			if fmt.Sprintf("%v", ValueFoldInterface(res[1])) == fmt.Sprintf("%v", v[1]) {
				wasVar = true
				numberS := fmt.Sprintf("%v", CompilerVars[fmt.Sprintf("%v", ValueFoldInterface(res[1]))])
				_, err := progFile.Write([]byte("\nmov $lenVarName, %rsi \n mov $varName, %rdx" +
					"\n mov $lenVarName" + numberS +
					", %rax \n mov $varName" + numberS + ", %rdi\n call __set\n call __getVar" +
					"\n mov (userData), %rsi \n call __print"))
				if nil != err {
					panic(err)
				}
			}
		}
		if !wasVar {
			_, err := dataFile.Write([]byte("\ndata" + fmt.Sprintf("%v", DataNumber) + ":"))
			if nil != err {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = dataFile.Write([]byte("\n.ascii " + fmt.Sprintf("%v", ValueFoldInterface(res[1])) + "\n.space 1, 0"))
			if nil != err {
				fmt.Println(err)
				os.Exit(1)
			}
			_, err = progFile.Write([]byte("\nmov $data" + fmt.Sprintf("%v", DataNumber) + ", %rsi\ncall __print"))
			if nil != err {
				fmt.Println(err)
				os.Exit(1)
			}
			DataNumber++
		}

	}
}
