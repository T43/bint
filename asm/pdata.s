.data
pageSize:
.quad 4096
varNameSize:
.quad 32
varSize:
.quad 128
typeSize:
.quad 32 
valSize:
.quad 64 
labelSize:
.quad 128 
labelsMax:
.quad 0, 0, 0, 0, 0, 0, 0, 0
buf:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenBuf = . - buf 
buf2:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenBuf2 = . - buf2 
buf3:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenBuf3 = . - buf3
buf4:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenBuf4 = . - buf4
strBegin:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenStrBegin = . - strBegin
oldHeapMax:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenOldHeapMax = . - oldHeapMax
strPointer:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenStrPointer = . - strPointer
strMax:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenStrMax = . - strMax 
isNeg:
.byte 0 
lenIsNeg = . - isNeg
isExpNeg:
.byte 0
lenIsExpNeg = . - isExpNeg
varType:
.quad 0, 0, 0, 0
lenVarType = . - varType 
varName:
.quad 0, 0, 0, 0
lenVarName = . - varName 
userData:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenUserData = . - userData
lastUserData:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenLastUserData = . - lastUserData
strPageNumber:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenStrNumber = . - strPageNumber
memorySize:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenMemorySize = . - memorySize
memoryBegin:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenMemoryBegin = . - memoryBegin
labelsEnd:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenLabelsEnd = . - labelsEnd
labelsPointer:
.quad 0, 0, 0, 0, 0, 0, 0, 0
lenLabelsPointer = . - labelsPointer 
intType:
.ascii "int"
.space 1, 0
lenIntType = . - intType
floatType:
.ascii "float"
.space 1, 0
lenFloatType = . - floatType
boolType:
.ascii "bool"
.space 1, 0
lenBoolType = . - boolType
stringType:
.ascii "string"
.space 1, 0
lenStringType = . - stringType
enter:
.ascii "\n"
.space 1, 0
lenEnter = . - enter 
ten:
.float 10.0 
one:
.float 1.0 
zero:
.float 0.0 
floatTail:
.ascii ".0"
.space 1, 0 
lenFloatTail =  . - floatTail 

fatalError:
.ascii "fatal error: internal error\n"
.space 1, 0 
divZeroError:
.ascii "runtime error: dividing by zero\n"
.space 1, 0
divINegError:
.ascii "runtime error: @ is not defined for negative numbers\n"
.space 1, 0
powNegError:
.ascii "runtime error: ^ is not defined for negative base and fractional exponent\n"
.space 1, 0
powZeroNegError:
.ascii "runtime error: ^ is not defined for zero base and negative exponent\n"
.space 1, 0
powZeroZeroError:
.ascii "runtime error: ^ is not defined for zero base and zero exponent\n"
.space 1, 0
noSuchMarkError:
.ascii "runtime error: no such mark: "
.space 1, 0

