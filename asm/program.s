.text	

__throughError:
 mov $fatalError, %rsi
 call __print 
 mov $60, %rax
 mov $1, %rdi
 syscall

__print:
 mov (%rsi), %al	
 cmp $0, %al	
 jz  __printEx			
 mov $1, %rdi	
 mov $1, %rdx
 mov $1, %rax	
 syscall		    
 inc %rsi		  		    
 jnz __print
__printEx:
 ret

# посчитать количество символов до 0
# адрес начала в %rsi
# результат в %rax  
 __len:
 xor %rax, %rax 
 __lenLocal:
 mov (%rsi), %dl	
 cmp $0, %dl	
 jz  __lenEx				    
 inc %rsi		  	
 inc %rax 	    
 jmp __lenLocal  
__lenEx:
 ret

 __printSymbol:
 mov (%rsi), %al				
 mov $1, %rdi	
 mov $1, %rdx
 mov $1, %rax	
 syscall		    
 ret

__printHeap:  
 mov %r13, %r8  
 __printHeapLoop:
 cmp (strMax), %r8 
 jz __printHeapEx
 mov %r8, %rsi 
 mov $1, %rdi	
 mov $1, %rdx
 mov $1, %rax	
 syscall
 inc %r8 
 jmp __printHeapLoop
 __printHeapEx:
 ret 


__toStr:
 # число в %rax 
 # подготовка преобразования числа в строку
  cmp $0, %rax 
  jg __toStrPos
  mov $0, %rdx 
  sub $1, %rdx 
  imul %rdx  
  movb $'-', (buf2)
  jmp __toStrNeg 
  __toStrPos:
  movq $0, (buf2)
  __toStrNeg:
  mov $10, %r8    # делитель
  mov $buf, %rsi  # адрес начала буфера 
  xor %rdi, %rdi  # обнуляем счетчик
# преобразуем путем последовательного деления на 10
__toStrlo:
  xor %rdx, %rdx  # число в rdx:rax
  div %r8         # делим rdx:rax на r8
  add $48, %dl    # цифру в символ цифры
  mov %dl, (%rsi) # в буфер
  inc %rsi        # на следующую позицию в буфере
  inc %rdi        # счетчик увеличиваем на 1
  cmp $0, %rax    # проверка конца алгоритма
  jnz __toStrlo          # продолжим цикл?
# число записано в обратном порядке,
# вернем правильный, перенеся в другой буфер 
  mov $buf2, %rbx # начало нового буфера
  mov (%rbx), %dl 
  cmp $'-', %dl
  jnz __toStrEmpty
  inc %rbx 
  __toStrEmpty:
  mov $buf, %rcx  # старый буфер
  add %rdi, %rcx  # в конец
  dec %rcx        # старого буфера
  mov %rdi, %rdx  # длина буфера
# перенос из одного буфера в другой
__toStrexc:
  mov (%rcx), %al # из старого буфера
  mov %al, (%rbx) # в новый
  dec %rcx        # в обратном порядке  
  inc %rbx        # продвигаемся в новом буфере
  dec %rdx        # а в старом в обратном порядке
  jnz __toStrexc         # проверка конца алгоритма 
  movb $0, (%rbx)
  mov $buf2, %rbx 
  mov (%rbx), %dl 
  cmp $'-', %dl
  jnz __toStrEnd 
  inc %rbx 
  mov (%rbx), %dl 
  cmp $'0', %dl
  jnz __toStrEnd 
  inc %rbx 
  mov (%rbx), %dl 
  cmp $0, %dl 
  jnz __toStrEnd
  mov $buf2, %rbx 
  movb $'0', (%rbx) 
  inc %rbx 
  movb $0, (%rbx) 
__toStrEnd:
  ret

__set: #set strings
 # входные параметры 
 # rsi - длина буфера назначения 
 # rdx - адрес буфера назначения
 # rax - длина буфера источника 
 # rdi - адрес буфера источника 

 mov %rdx, %r8 
 mov %rsi, %r9
  
 __setClear:
 cmp $0, %rsi
 jz __setClearEnd
 movb $'*', (%rdx)
 dec %rsi
 inc %rdx  
 jmp __setClear
 __setClearEnd:
 dec %rdx 
 movb $0, (%rdx)

 mov %r8, %rdx  
 mov %r9, %rsi  
 
 __setLocal:
 cmp $0, %rax 
 jz __setLocalEnd
 cmp $2, %rsi
 jz __setLocalEnd
 mov (%rdi), %r11b 
 movb %r11b, (%rdx)
 inc %rdx
 inc %rdi
 dec %rax  
 dec %rsi 
 jmp __setLocal
 __setLocalEnd:
 dec %rdx 
 mov (%rdx), %rax 
 cmp $'*', %rax 
 jz __star
 inc %rdx
 __star: 
 movb $0, (%rdx)
 ret 

__concatinate:
 # входные параметры 
 # r8 - длина буфера первого операнда 
 # r9 - адрес буфера первого операнда
 # r11 - адрес буфера второго операнда 
 # выход
 # userData   
 call __clearUserData
 mov $lenUserData, %rsi # присваиваем в userData первый операнд  
 mov $userData, %rdx 
 mov %r8, %rax 
 mov %r9, %rdi 

 mov %r8, %rbx 
 mov %r11, %rcx
 call __set
 mov %rbx, %r8 
 mov %rcx, %r11 

 mov $userData, %r8
 __concNext:
 mov (%r8), %dl 
 cmp $0, %dl 
 jz __concLocal 
 inc %r8   
 jmp __concNext 
  
 __concLocal:
 
 mov (%r11), %dl 
 movb %dl, (%r8)
 inc %r8 
 inc %r11
 #dec %r10  
 cmp $0, %dl 
 jnz __concLocal
 #movb $0, (%r8)
 
 ret 

__toNumber:
 # вход: buf 
 # выход:  %rax 
 mov $buf, %rdx # our string
 movzx (%rdx), %rcx 
 cmp $'-', %rcx 
 jnz __toNumberAtoi
 inc %rdx 
 __toNumberAtoi:
 xor %rax, %rax # zero a "result so far"
 __toNumberTop:
 movzx (%rdx), %rcx # get a character
 inc %rdx # ready for next one
 cmp $0, %rcx # end?
 jz __toNumberDone
 sub $48, %rcx # "convert" character to number
 imul $10, %rax # multiply "result so far" by ten
 add %rcx, %rax # add in current digit
 jmp __toNumberTop # until done
 __toNumberDone:
 mov $buf, %rdx 
 movzx (%rdx), %rcx 
 cmp $'-', %rcx 
 jnz __toNumberIsPos
 mov $0, %rdx 
 sub $1, %rdx 
 //mov %rdx, (buf)
 imul %rdx 
 __toNumberIsPos:
 ret


__defineVar:
 # адрес имени переменной в $varName
 # адрес типа переменой в $varType
 
 mov $varName, %rcx 
 mov $varType, %rdx 
 mov %r14, %rax

 cmp %rax, %r15
 jg __defOk 
 mov %r15, %r8
 call __newMem
 mov $varName, %rcx 
 mov $varType, %rdx
 __defOk:
 mov %r14, %r8 
 __defOkLocal:
 movb (%rcx), %r11b 
 cmp $'*', %r11b
 jz __defOkLocalEx
 movb %r11b, (%r8)
 inc %rcx 
 inc %r8 
 jmp __defOkLocal
 __defOkLocalEx:
 mov %r14, %r8 
 add (varNameSize), %r8 
 __defOkTypeLocal:
 movb (%rdx), %r11b
 cmp $'*', %r11b 
 jz __defOkTypeLocalEx
 movb %r11b, (%r8)
 inc %rdx
 inc %r8 
 jmp __defOkTypeLocal
 __defOkTypeLocalEx:
 
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenVarType, %rax 
 mov $varType, %rdi 
 call __set 
 mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenIntType, %rax 
 mov $intType, %rdi 
 call __set 
  
 call __compare 
 cmp $1, %rax 
 jz __defInt 

 mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenFloatType, %rax 
 mov $floatType, %rdi 
 call __set 
 call __compare
 cmp $1, %rax 
 jz __defFloat 

 mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBoolType, %rax 
 mov $boolType, %rdi 
 call __set 
 call __compare
 cmp $1, %rax  
 jz __defBool 

 mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenStringType, %rax 
 mov $stringType, %rdi 
 call __set 
 call __compare
 cmp $1, %rax 
 jz __defString 
 call __throughError

 __defInt:
 mov %r14, %r8 
 add (varNameSize), %r8 
 add (typeSize), %r8
 movb $'0',(%r8)
 inc %r8 
 movb $0, (%r8)
 jmp __defEnd
 __defFloat:
 mov %r14, %r8 
 add (varNameSize), %r8 
 add (typeSize), %r8
 movb $'0',(%r8)
 inc %r8 
 movb $'.', (%r8)
 inc %r8 
 movb $'0', (%r8)
 inc %r8 
 movb $0, (%r8)
 mov %r14, %r8 
 add (varNameSize), %r8 
 add (typeSize), %r8 
 jmp __defEnd
 __defBool:
 mov %r14, %r8 
 add (varNameSize), %r8 
 add (typeSize), %r8
 movb $'0',(%r8)
 inc %r8 
 movb $0, (%r8) 
 jmp __defEnd
 __defString:
 mov %r14, %r8 
 add (varNameSize), %r8 
 add (typeSize), %r8
 mov (strPointer), %rax 
 movb $0, (%rax)
 mov %r8, %r12 
 call __toStr 
 mov %r12, %r8 
 mov $buf2, %rax
 __defAddr:
 mov (%rax), %dl 
 cmp $0, %dl  
 jz __defStrEnd 
 mov %dl, (%r8)
 inc %rax 
 inc %r8 
 jmp __defAddr
 __defStrEnd:
 mov (strPointer), %rax 
 add (valSize), %rax 
 cmp (strMax), %rax 
 jg __defStrNewMem
 mov %rax, (strPointer)
 jmp __defEnd 
 __defStrNewMem:
 mov %rax, (strPointer)
 call __newStrMem
 __defEnd:

 mov %r14, %rax 
 add (varSize), %rax 
 mov %rax, %r14
 ret 

# r12 - pointer (общего назначения)
# r13 - heapBegin 
# r14 - heapPointer 
# r15 - heapMax 

__firstMem:
 # получить адрес начала области для выделения памяти
 mov $12, %rax
 xor %rdi, %rdi
 syscall
# запомнить адрес начала выделяемой памяти
 mov %rax, %r8  
 mov %rax, %r13
 mov %rax, %r14
 mov %rax, %r9 
 add (pageSize), %r9
 mov %r9, %r15 
# выделить динамическую память
 mov (pageSize), %rdi
 add %rax, %rdi
 mov $12, %rax
 syscall
# обработка ошибки
 cmp $-1, %rax
 jz __throughError
# заполним выделенную память
 mov $'*', %dl
 mov $0, %rbx
 __lo:
 movb %dl, (%r8)
 inc %rbx
 inc %r8 
 cmp (pageSize), %rbx
 jz  __ex
 jmp __lo
 __ex:
 ret 

 __firstStrMem:
 # адрес начала области для выделения памяти
 mov %r15, %rax
# запомнить адрес начала выделяемой памяти
 mov %rax, %r8  
 mov %rax, (strBegin)
 mov %rax, (strPointer)
 mov %rax, %r9 
 add (pageSize), %r9
 mov %r9, (strMax)
# выделить динамическую память
 mov (pageSize), %rdi
 add %rax, %rdi
 mov $12, %rax
 syscall
# обработка ошибки
 cmp $-1, %rax
 jz __throughError
# заполним выделенную память
 mov $'*', %dl
 mov $0, %rbx
 __firstStrMemLo:
 movb %dl, (%r8)
 inc %rbx
 inc %r8 
 cmp (pageSize), %rbx
 jz  __firstStrMemEx
 jmp __firstStrMemLo
 __firstStrMemEx:
 ret 


 __newMem:
 # адрес начала выделяемой памяти в  %r8 
# запомнить адрес начала выделяемой памяти
 #mov %r8, %r14
 mov %r8, %r9 
 add (pageSize), %r9 
 mov %r9, %r15
# выделить динамическую память
 mov (pageSize), %rdi
 add %rax, %rdi
 mov $12, %rax
 syscall
# обработка ошибки
 cmp $-1, %rax
 jz __throughError
# заполним выделенную память
 mov $'*', %dl
 mov $0, %rbx
 __newMemlo:
 movb %dl, (%r8)
 inc %rbx
 inc %r8 
 cmp (pageSize), %rbx
 jz  __newMemEx
 jmp __newMemlo
 __newMemEx:
 ret  

 __newStrMem:
 # адрес начала выделяемой памяти в  %r8 
# запомнить адрес начала выделяемой памяти
 #mov %r8, %r14

 mov %r8, %r9 
 add (pageSize), %r9 
 mov %r9, (strMax)
# выделить динамическую память
 mov (pageSize), %rdi
 add %rax, %rdi
 mov $12, %rax
 syscall
# обработка ошибки
 cmp $-1, %rax
 jz __throughError
# заполним выделенную память
 mov $'*', %dl
 mov $0, %rbx
 __newStrMemlo:
 movb %dl, (%r8)
 inc %rbx
 inc %r8 
 cmp (pageSize), %rbx
 jz  __newStrMemEx
 jmp __newStrMemlo
 __newStrMemEx:
 ret 

__read: 
 # считать в buf по указателю в %r12 
 mov %r12, %r8
 mov $buf, %r10 
 mov $lenBuf, %rsi 
__readClear:
 movb $0, (%r10)
 dec %rsi
 inc %r10
 cmp $0, %rsi  
 jnz __readClear
 mov $buf, %r10
 __readLocal: 
 movb (%r8), %r9b
 cmp $'*', %r9b  
 jz __readEx
 mov %r9b, (%r10)
 inc %r10 
 inc %r8 
 jmp __readLocal
 __readEx:
 mov $buf, %r10 
 cmp $0, (%r10)
 jnz __readOk
 movb $'*', (%r10)
 __readOk:
 ret 
 

__compare:
 # сравнить строки по адресу $buf и $buf2
 mov $buf, %rax 
 mov $buf2, %rbx 
 __compareLocal:
 movb (%rax), %dl 
 cmp $0, %dl 
 jz __equal
 movb (%rax), %dl
 cmp %dl, (%rbx) 
 jnz __notEqual
 inc %rax 
 inc %rbx 
 jmp __compareLocal

 __notEqual:
 mov $0, %rax 
 ret 
 __equal:
 mov $1, %rax  
 ret 

__setVar:
 # вход: 
 # имя переменной по адресу $varName 
 # данные по указателю в (userData) 
 mov %r13, %rbx
 __setVarLocal:
 cmp %r15, %rbx
 jg __setVarEnd

 mov %rbx, %r12 
 call __read 
 cmp $'*', (buf)
 jz __setVarEnd 
 
 add (varSize), %rbx 
 jmp __setVarLocal  
  
 __setVarEnd:
 __setVarSearch:
 sub (varSize), %rbx 
 mov %rbx, %r12 
 call __read 
 cmp $'*', (buf)
 jz __throughError
 mov $buf, %rsi 
 mov %rbx, %r12 
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenVarName, %rax 
 mov $varName, %rdi 
 call __set
 call __compare
 mov %r12, %rbx 
 cmp $0, %rax 
 jz __setVarSearch
 

 add (varNameSize), %rbx 
 mov %rbx, %rax 
 mov %rbx, %r12 
 call __read  
 add (typeSize), %rbx 

 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenStringType, %rax 
 mov $stringType, %rdi 
 call __set
 mov %rbx, %r12 
 call __compare
 mov %r12, %rbx 
 cmp $0, %rax 
 jnz __setVarStr
 
 #add (varNameSize), %rbx 
 #add (typeSize), %rbx
 mov %rbx, %r8 
 mov %rbx, %rax 
 __setVarClear:
 add (valSize), %rax 
 __setVarClearLocal: 
 cmp %rax, %rbx 
 jz __setVarClearEnd
 movb $'*', (%rbx) 
 inc %rbx 
 jmp __setVarClearLocal
 __setVarClearEnd:
 mov %r8, %rbx 
 jmp __setVarIsNotStr

 __setVarStr:
 mov %rbx, %r12 
 call __read 
 call __toNumber
 mov %rax, %rbx  

 __setVarIsNotStr:
 
 #call __throughError
 mov %rbx, %r10 # сохраняем значение %rbx  
 mov (userData), %rax 
 xor %rdi, %rdi # счетчик количества реально записанных байт 
 __setNow:
 mov (%rax), %dl
 cmp $0, %dl 
 jz __setVarRet 
 mov %dl, (%rbx)
 inc %rbx 
 inc %rax 
 inc %rdi 
 jmp __setNow 

 __setVarRet:
 movb $0, (%rbx) 
 ret

 __getVar:
 # вход: имя переменной по адресу $varName 
 # выход: указатель на данные в (userData)
 mov %r13, %rbx
 __getVarLocal:
 cmp %r15, %rbx
 jg __getVarEnd

 mov %rbx, %r12 
 call __read 
 cmp $'*', (buf)
 jz __getVarEnd 
 
 add (varSize), %rbx 
 jmp __getVarLocal  
  
 __getVarEnd:
 __getVarSearch:
 sub (varSize), %rbx 
 mov %rbx, %r12 
 call __read 
 cmp $'*', (buf)
 jz __throughError
 mov $buf, %rsi 
 mov %rbx, %r12 
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenVarName, %rax 
 mov $varName, %rdi 
 call __set
 call __compare
 mov %r12, %rbx 
 cmp $0, %rax 
 jz __getVarSearch
 
 add (varNameSize), %rbx 
 mov %rbx, %r12 
 call __read 
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenStringType, %rax 
 mov $stringType, %rdi 
 call __set
 call __compare 
 cmp $1, %rax 
 jnz __getVarNotStr
 mov $1, %r11
 jmp __getVarNotStrEnd
 __getVarNotStr:
 mov $0, %r11 
 __getVarNotStrEnd:
 mov %r12, %rbx 
 add (typeSize), %rbx  
 
 __getNow:
 #call __toNumber
 cmp $1, %r11
 jz __getVarGetStr
 mov %rbx, (userData)
 ret 
 __getVarGetStr:
 mov %rbx, %r12 
 call __read 
 call __toNumber
 mov %rax, (userData) 
 ret 
 /*mov $userData, %rsi
 __getNowLocal:  
 cmp $0, %rax 
 jz __getVarRet
 mov (%rbx), %dl 
 mov %dl, (%rsi)
 inc %rsi 
 inc %rbx  
 dec %rax 
 jmp __getNowLocal 
 __getVarRet:
 ret*/

__clearBuf:
mov $buf, %rsi 
mov $lenBuf, %rdi
__clearBufLocal: 
cmp $1, %rdi 
jz __clearBufEnd
movb $'*', (%rsi)
inc %rsi 
dec %rdi 
jmp __clearBufLocal

__clearBufEnd:
movb $0, (%rsi)
ret

__clearBuf2:
mov $buf2, %rsi 
mov $lenBuf2, %rdi
__clearBufLocal2: 
cmp $1, %rdi 
jz __clearBufEnd2
movb $'*', (%rsi)
inc %rsi 
dec %rdi 
jmp __clearBufLocal2

__clearBufEnd2:
movb $0, (%rsi)
ret

__clearUserData:
mov $userData, %rsi 
mov $lenUserData, %rdi
__clearUserDataLocal: 
cmp $1, %rdi 
jz __clearUserDataEnd
movb $'*', (%rsi)
inc %rsi 
dec %rdi 
jmp __clearUserDataLocal

__clearUserDataEnd:
movb $0, (%rsi)
ret

__clearBuf3:
mov $buf3, %rsi 
mov $lenBuf3, %rdi
__clearBufLocal3: 
cmp $1, %rdi 
jz __clearBufEnd3
movb $'*', (%rsi)
inc %rsi 
dec %rdi 
jmp __clearBufLocal3

__clearBufEnd3:
movb $0, (%rsi)
ret

__clearBuf4:
mov $buf4, %rsi 
mov $lenBuf4, %rdi
__clearBufLocal4: 
cmp $1, %rdi 
jz __clearBufEnd4
movb $'*', (%rsi)
inc %rsi 
dec %rdi 
jmp __clearBufLocal4

__clearBufEnd4:
movb $0, (%rsi)
ret

__add:
 # вход: buf и buf2
 # %rax - тип операции 
 # 0 - целочисленное сложение 
 # 1 - сложение вещественных чисел   
 # выход: userData 
 call __clearUserData
 cmp $0, %rax 
 jz __addInt 
 cmp $1, %rax 
 jz __addFloat

 call __throughError

 __addInt:
 call __toNumber
 mov %rax, %rbx 
 call __clearBuf
 mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf2, %rax 
 mov $buf2, %rdi 
 call __set 
 call __toNumber
 add %rbx, %rax
 call __toStr 
 mov $lenUserData, %rsi 
 mov $userData, %rdx 
 mov $lenBuf2, %rax 
 mov $buf2, %rdi 
 call __set 
 ret 
 __addFloat:
 call __clearBuf4
 mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenBuf2, %rax 
 mov $buf2, %rdi 
 call __set
 call __parseFloat
 movss %xmm0, %xmm1 
 call __clearBuf
 mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi 
 call __set
 call __parseFloat
 movss %xmm0, (buf)
 fld (buf)
 movss %xmm1, (buf)
 fadd (buf)
 fstp (buf)
 call __floatToStr

 ret 

  __sub:
 # вход: buf и buf2
 # %rax - тип операции 
 # 0 - целочисленное сложение 
 # 1 - сложение вещественных чисел   
 # выход: userData 
 call __clearUserData
 cmp $0, %rax 
 jz __subInt 
 cmp $1, %rax 
 jz __subFloat

 call __throughError

 __subInt:
 call __toNumber
 mov %rax, %rbx 
 call __clearBuf
 mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf2, %rax 
 mov $buf2, %rdi 
 call __set 
 call __toNumber
 sub %rax, %rbx
 mov %rbx, %rax 
 call __toStr 
 mov $lenUserData, %rsi 
 mov $userData, %rdx 
 mov $lenBuf2, %rax 
 mov $buf2, %rdi 
 call __set 
 ret 
 __subFloat:
 call __clearBuf4
 mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenBuf2, %rax 
 mov $buf2, %rdi 
 call __set
 call __parseFloat
 movss %xmm0, %xmm1 
 call __clearBuf
 mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi 
 call __set
 call __parseFloat
 movss %xmm1, (buf)
 fld (buf)
 movss %xmm0, (buf)
 fsub (buf)
 fstp (buf)
 call __floatToStr

 ret 


__floatToStr:
# вход: buf
# выход userData
# проверяем, является ли число отрицательным
movss (zero), %xmm0 
movss (buf), %xmm1 
cmpss $1, %xmm0, %xmm1
pextrb $3, %xmm1, %rax
cmp $0, %rax 
jz __floatToStrIsPos
# меняем знак 
fld (zero) 
fsub (one)
fmul (buf)
fstp (buf)
movb $1, (isNeg) # признак отрицательного числа 
jmp __floatToStrIsNeg
__floatToStrIsPos:
movb $0, (isNeg)
__floatToStrIsNeg:
fld (buf)
movss (buf), %xmm0 
roundps $3, %xmm0, %xmm0
movss %xmm0, (buf)
cvtss2si (buf), %r12
fsub (buf) # вычитаем из значения целое значение, получаем дробное 
fstp (buf)


mov $6, %r10 # 6 знаков после запятой 

__floatToStrLocal:
fld (buf)
cmp $0, %r10
jz __floatToStrOk
dec %r10 
movss (ten), %xmm0
movss %xmm0, (buf)
fmul (buf)
fstp (buf)
jmp __floatToStrLocal
__floatToStrOk:
cvtss2si (buf), %rax # здесь содержится дробное значение 
call __toStr

call __clearBuf
mov $buf, %rax 
movb $48, (%rax)
inc %rax 
movb $0, (%rax)

__floatToStrZeros:
mov $buf2, %rsi 
call __len 
cmp $6, %rax 
jz __floatToStrEndZeros

mov $lenBuf, %r8 
mov $buf, %r9 
mov $buf2, %r11
call __concatinate

mov $lenBuf2, %rsi 
mov $buf2, %rdx 
mov $lenUserData, %rax 
mov $userData, %rdi 
call __set

jmp __floatToStrZeros

__floatToStrEndZeros:

call __clearBuf3 # в buf3 содержится дробное значение в виде строки 
mov $lenBuf3, %rsi 
mov $buf3, %rdx 
mov $lenBuf2, %rax 
mov $buf2, %rdi 
call __set 

mov %r12, %rax 
call __toStr # в buf2 содержится целое значение в виде строки
call __clearBuf
mov $buf, %rax 
movb $'.', (%rax)
inc %rax 
movb $0, (%rax) 

mov $lenBuf2, %r8 
mov $buf2, %r9 
mov $buf, %r11 
call __concatinate



call __clearBuf 
mov $lenBuf, %rsi 
mov $buf, %rdx 
mov $lenUserData, %rax 
mov $userData, %rdi 
call __set


mov $lenBuf, %r8
mov $buf, %r9 
mov $buf3, %r11 
call __concatinate 

# число отрицательное?
mov (isNeg), %dl 
cmp $1, %dl  
jnz __floatToStrEnd 
call __clearBuf
mov $buf, %rax 
movb $'-', (%rax)
inc %rax 
movb $0, (%rax)

mov $lenBuf3, %rsi 
mov $buf3, %rdx 
mov $lenUserData, %rax 
mov $userData, %rdi 
call __set 

mov $lenBuf, %r8 
mov $buf, %r9 
mov $buf3, %r11 
call __concatinate

__floatToStrEnd:
ret 

__parseFloat:
# $buf - источник (строка)
# %xmm0 - результат
mov $buf, %rax 
call __clearBuf2
call __clearBuf3
mov $buf2, %rbx # здесь будет содержаться целая часть 
mov $buf3, %rcx # здесь будет содержаться дробная часть
mov (%rax), %dl 
cmp $'-', %dl 
jnz isPos
mov $1, %r12 # признак отрицательного числа 
inc %rax 
jmp __parseFloatLocal 
isPos:
mov $0, %r12 
__parseFloatLocal: 
mov (%rax), %dl 
cmp $'.', %dl
jz __point
mov %dl, (%rbx)
inc %rax 
inc %rbx 
jmp __parseFloatLocal
__point:
movb $0, (%rbx)
mov %rax, %rbx 
mov $buf2, %rsi 
call __len 
cmp $8, %rax # целое число - не более 7 цифр 
jl __parseFloatZ
mov $buf2, %rbx 
movb $48, (%rbx)
inc %rbx 
movb $0, (%rbx) 
mov $buf3, %rbx 
movb $48, (%rbx)
inc %rbx 
movb $0, (%rbx)
jmp __parseNow 
__parseFloatZ: 
mov %rbx, %rax 

__pointLocal:             
inc %rax
mov (%rax), %dl
cmp $0, %dl 
jz __parseNow
mov (%rax), %dl 
mov %dl, (%rcx) 
inc %rcx 
jmp __pointLocal   
__parseNow:
movb $0, (%rcx)

call __clearBuf
mov $lenBuf, %rsi 
mov $buf, %rdx 
mov $lenBuf2, %rax 
mov $buf2, %rdi 
call __set 
call __toNumber
mov %rax, %r10 # целая часть числа в %r10

call __clearBuf
mov $lenBuf, %rsi 
mov $buf, %rdx 
mov $lenBuf3, %rax 
mov $buf3, %rdi 
call __set

mov $buf, %rsi 
call __len 
mov %rax, %rbx # длина дробной части числа в %rbx 
cmp $6, %rbx # дробная часть не более шести знаков 
jl __parseFloatCut
mov $buf, %rsi 
add $6, %rsi 
movb $0, (%rsi)
mov $6, %rbx 
__parseFloatCut: 
call __toNumber
mov %rax, (buf)
cvtsi2ss (buf), %xmm0  
movss %xmm0, (buf)

__floatLocal:
fld (buf)
cmp $0, %rbx 
jz __floatOk
dec %rbx 
movss (ten), %xmm0
movss %xmm0, (buf)
fdiv (buf)
fstp (buf)
jmp __floatLocal
__floatOk:
mov %r10, (buf)
cvtsi2ss (buf), %xmm0    
movss %xmm0, (buf) # целая часть числа 
fadd (buf)
fstp (buf)
movss (buf), %xmm0  
cmp $1, %r12 
jnz __pos
mov (zero), %rax   
mov %rax, (buf)
fld (buf)
movss %xmm0, (buf)
fsub (buf)
fstp (buf)
movss (buf), %xmm0 
__pos:
ret 

.globl _start
_start:
 call __firstMem
 call __firstStrMem


mov $lenVarName, %rsi 
 mov $varName, %rdx
 mov $lenVarName0, %rax 
 mov $varName0, %rdi 
 call __set 
 mov $lenVarType, %rsi 
 mov $varType, %rdx 
 mov $lenIntType, %rax 
 mov $intType, %rdi 
 call __set 
 mov $varName, %rcx 
 mov $varType, %rdx  
 call __defineVar
mov $lenVarName, %rsi 
 mov $varName, %rdx
 mov $lenVarName1, %rax 
 mov $varName1, %rdi 
 call __set 
 mov $lenVarType, %rsi 
 mov $varType, %rdx 
 mov $lenIntType, %rax 
 mov $intType, %rdi 
 call __set 
 mov $varName, %rcx 
 mov $varType, %rdx  
 call __defineVar
mov $lenVarName, %rsi 
 mov $varName, %rdx
 mov $lenVarName2, %rax 
 mov $varName2, %rdi 
 call __set 
 mov $lenVarType, %rsi 
 mov $varType, %rdx 
 mov $lenFloatType, %rax 
 mov $floatType, %rdi 
 call __set 
 mov $varName, %rcx 
 mov $varType, %rdx  
 call __defineVar
 mov $lenVarName, %rsi 
 mov $varName, %rdx 
 mov $lenVarName0, %rax 
 mov $varName0, %rdi 
call __set

 mov $data0, %rax  
 mov %rax, (userData)
 call __setVar
 mov $lenVarName, %rsi 
 mov $varName, %rdx 
 mov $lenVarName1, %rax 
 mov $varName1, %rdi 
call __set

 mov $data1, %rax  
 mov %rax, (userData)
 call __setVar
 mov $lenVarName, %rsi 
 mov $varName, %rdx 
 mov $lenVarName2, %rax 
 mov $varName2, %rdi 
call __set

 mov $data2, %rax  
 mov %rax, (userData)
 call __setVar
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenData3, %rax 
 mov $data3, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData4, %rax 
 mov $data4, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT0, %rsi 
 mov $t0, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData5, %rax 
 mov $data5, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT0, %rax 
 mov $t0, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT1, %rsi 
 mov $t1, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData6, %rax 
 mov $data6, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT1, %rax 
 mov $t1, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT2, %rsi 
 mov $t2, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData7, %rax 
 mov $data7, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT2, %rax 
 mov $t2, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT3, %rsi 
 mov $t3, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData8, %rax 
 mov $data8, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT3, %rax 
 mov $t3, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT4, %rsi 
 mov $t4, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData9, %rax 
 mov $data9, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT4, %rax 
 mov $t4, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT5, %rsi 
 mov $t5, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData10, %rax 
 mov $data10, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT5, %rax 
 mov $t5, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT6, %rsi 
 mov $t6, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData11, %rax 
 mov $data11, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT6, %rax 
 mov $t6, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT7, %rsi 
 mov $t7, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData12, %rax 
 mov $data12, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT7, %rax 
 mov $t7, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT8, %rsi 
 mov $t8, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData13, %rax 
 mov $data13, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT8, %rax 
 mov $t8, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT9, %rsi 
 mov $t9, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenBuf4, %rsi 
 mov $buf4, %rdx 
 mov $lenData14, %rax 
 mov $data14, %rdi
 call __set
mov $lenBuf3, %rsi 
 mov $buf3, %rdx 
 mov $lenT9, %rax 
 mov $t9, %rdi
 call __set
mov $lenBuf, %rsi 
 mov $buf, %rdx 
 mov $lenBuf3, %rax 
 mov $buf3, %rdi
 call __set
 mov $lenBuf2, %rsi 
 mov $buf2, %rdx 
 mov $lenBuf4, %rax 
 mov $buf4, %rdi
 call __set 
 xor %rax, %rax 

 call __add 
 mov $lenT10, %rsi 
 mov $t10, %rdx 
 mov $lenUserData, %rax 
 mov $userData, %rdi
 call __set
mov $lenVarName, %rsi 
 mov $varName, %rdx 
 mov $lenVarName0, %rax 
 mov $varName0, %rdi 
 call __set
mov $lenVarName, %rsi 
 mov $varName, %rdx 
 mov $lenVarName0, %rax 
 mov $varName0, %rdi
 call __set 
 mov $t10, %rax 
 mov %rax, (userData)
 call __setVar
mov $lenVarName, %rsi 
 mov $varName, %rdx 
 mov $lenVarName3, %rax 
 mov $varName3, %rdi
 call __set 
 mov $lenVarType, %rsi 
 mov $varType, %rdx 
 mov $lenStringType, %rax
 mov $stringType, %rdi
 call __set 
 call __defineVar
mov $lenVarName, %rsi 
 mov $varName, %rdx
 mov $lenVarName0, %rax 
 mov $varName0, %rdi
 call __set
 call __getVar
mov $lenVarName, %rsi 
 mov $varName, %rdx
 mov $lenVarName3, %rax 
 mov $varName3, %rdi
 call __set 
 call __setVar
mov $lenVarName, %rsi 
 mov $varName, %rdx
 mov $lenVarName3, %rax 
 mov $varName3, %rdi
 call __set
 call __getVar
 mov (userData), %rsi 
 call __print
mov $data15, %rsi
call __print
mov $60,  %rax
xor %rdi, %rdi
syscall
