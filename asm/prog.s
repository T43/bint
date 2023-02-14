.data
pageSize:
.quad 4096 
varNameSize:
.quad 32
varSize:
.quad 128 
typeSize:
.quad 32 
buf:
.quad 0, 0, 0, 0
lenBuf = . - buf 
varType:
.quad 0, 0, 0, 0
lenVarType = . - varType 
varName:
.quad 0, 0, 0, 0
lenVarName = . - varName 
varName0:
.ascii "iVar"
lenVarName0 = . - varName0
varType0:
.ascii "int"
lenVarType0 = . - varType0
varName1:
.ascii "sVar"
lenVarName1 = . - varName1
varType1:
.ascii "string"
lenVarType1 = . - varType1
data0:
.ascii "25"
.space 1, 0
data1:
.ascii "Hello world!\n"
.space 1, 0

fatalError:
.ascii "fatal error: internal error\n"
.space 1, 0 

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

__printHeap:  
 mov %r13, %r8  
 __printHeapLoop:
 cmp %r15, %r8 
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

__set: #set strings
 # входные параметры 
 # rsi - длина буфера назначения 
 # rdx - адрес буфера назначения
 # rax - длина буфера источника 
 # rdi - адрес буфера источника 
 mov %rdx, %r8 
 mov %rsi, %r9
  
 __setClear:
 movb $'0', (%rdx)
 dec %rsi
 inc %rdx
 cmp $0, %rsi  
 jnz __setClear
 dec %rdx  
 movb $0, (%rdx)

 mov %r8, %rdx   
 
 __setLocal:
 mov (%rdi), %r11b
 movb %r11b, (%rdx)
 inc %rdx
 inc %rdi
 dec %rax  
 cmp $0, %rax
 jnz __setLocal
 
  
 ret 

__defineVar:
 # адрес имени переменной в %rcx
 # адрес типа переменой в %rdx
  
 mov %r14, %rax 
 cmp %rax, %r15
 jg __defOk 
 mov %r15, %r8
 call __newMem
 __defOk:
 mov %r14, %r8 
 __defOkLocal:
 mov (%rcx), %r11b 
 cmp $'0', %r11b
 jz __defOkLocalEx
 mov %r11b, (%r8)
 inc %rcx 
 inc %r8 
 jmp __defOkLocal
 __defOkLocalEx:
 mov %r14, %r8 
 add (varNameSize), %r8 
  __defOkTypeLocal:
 mov (%rdx), %r11b
 cmp $'0', %r11b 
 jz __defOkTypeLocalEx
 mov %r11b, (%r8)
 inc %rdx
 inc %r8 
 jmp __defOkTypeLocal
 __defOkTypeLocalEx:
 
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
 mov $'0', %dl
 mov $0, %rbx
 __lo:
 movb %dl, (%r8, %rbx)
 inc %rbx
 cmp (pageSize), %rbx
 jz  __ex
 jmp __lo
 __ex:
 ret 

 __newMem:
 # адрес начала выделяемой памяти в  %r8 
# запомнить адрес начала выделяемой памяти
 mov %r8, %r14
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
 mov $0, %dl
 mov $0, %rbx
 __newMemlo:
 movb %dl, (%r8, %rbx)
 inc %rbx
 cmp (pageSize), %rbx
 jz  __newMemEx
 jmp __newMemlo
 __newMemEx:
 ret 

__read: 
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
 mov (%r8), %r9b
 cmp $'0', %r9b  
 jz __readEx
 mov %r9b, (%r10)
 inc %r10 
 inc %r8 
 jmp __readLocal
 __readEx:
 ret 

__compare:
 mov $buf, %rax 
 mov $varName, %rax 
 
__setVar:
 # имя переменной в %rcx 
 # значение переменной в %rdx 
 mov %r13, %rax 
 cmp %r15, %rax 
 jg __setVarEx 
 mov %rax, %rsi
 mov %r13, %r12 
 call __read
 mov $buf, %rsi 
 call __print  
 __setVarEx:
 ret



.globl _start
_start:
 call __firstMem

 #iVar 
 mov $lenVarName, %rsi 
 mov $varName, %rdx 
 mov $lenVarName0, %rax 
 mov $varName0, %rdi 
 call __set 

 mov $lenVarType, %rsi 
 mov $varType, %rdx 
 mov $lenVarType0, %rax 
 mov $varType0, %rdi 
 call __set 

 mov $varName, %rcx 
 mov $varType, %rdx  
 call __defineVar
 
 #sVar
 mov $lenVarName, %rsi 
 mov $varName, %rdx 
 mov $lenVarName1, %rax 
 mov $varName1, %rdi 
 call __set 

 mov $lenVarType, %rsi 
 mov $varType, %rdx 
 mov $lenVarType1, %rax 
 mov $varType1, %rdi 
 call __set

 mov $varName, %rcx 
 mov $varType, %rdx  
 call __defineVar
 #mov (varName1), %rcx 
 #mov (varType1), %rdx
 #call __defineVar
 #mov $varType, %rsi 
 #call __print 
 #call __printHeap
 call __setVar 

__stop:
 mov $60,  %rax      # номер системного вызова exit
 xor %rdi, %rdi      # код возврата (0 - выход без ошибок)
 syscall   
