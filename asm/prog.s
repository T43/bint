.data
pageSize:
.quad 4096
varNameSize:
.quad 64
varSize:
.quad 256 
typeSize:
.quad 64 
enter:
.ascii "\n"
space:
.ascii " "
heapBegin:
.quad 0 
heapSize:
.quad 0
heapMax:
.quad 0
heapPointer:
.quad 0
getPointer:
.quad 0 
var0:
.ascii "sVar"
.space 1, 0
varT0:
.ascii "string"
.space 1, 0
var1:
.ascii "iVar"
.space 1, 0
varT1:
.ascii "int"
.space 1, 0
var2:
.ascii "fVar"
.space 1, 0
varT2:
.ascii "float"
.space 1, 0 

buf:
.space 256, 0
buf2:
.space 256, 0
s:
.space 200, 0
msg1:
.ascii "Hello world!\n"
len = . - msg1

.text

__print:
 mov (%rsi), %al	
 cmp $0, %al	
 jz  __ex			
 mov $1, %rdi	
 mov $1, %rdx
 mov $1, %rax	
 syscall		    
 inc %rsi		  
 dec %r8		    
 jnz __print
__ex:
 ret

__set: #set strings
 # %edi = %esi
 mov (%esi), %al
 mov %al, (%edi)
 inc %esi
 inc %edi
 cmp $0, (%esi)
 jnz __set
 ret 

__toStr:
 # число в %rax 
 # подготовка преобразования числа в строку
  movq $0, (buf2)
  mov $10, %r8    # делитель
  mov $buf, %rsi  # адрес начала буфера 
  xor %rdi, %rdi  # обнуляем счетчик
# преобразуем путем последовательного деления на 10
__lo:
  xor %rdx, %rdx  # число в rdx:rax
  div %r8         # делим rdx:rax на r8
  add $48, %dl    # цифру в символ цифры
  mov %dl, (%rsi) # в буфер
  inc %rsi        # на следующую позицию в буфере
  inc %rdi        # счетчик увеличиваем на 1
  cmp $0, %rax    # проверка конца алгоритма
  jnz __lo          # продолжим цикл?
# число записано в обратном порядке,
# вернем правильный, перенеся в другой буфер 
  mov $buf2, %rbx # начало нового буфера
  mov $buf, %rcx  # старый буфер
  add %rdi, %rcx  # в конец
  dec %rcx        # старого буфера
  mov %rdi, %rdx  # длина буфера
# перенос из одного буфера в другой
__exc:
  mov (%rcx), %al # из старого буфера
  mov %al, (%rbx) # в новый
  dec %rcx        # в обратном порядке  
  inc %rbx        # продвигаемся в новом буфере
  dec %rdx        # а в старом в обратном порядке
  jnz __exc         # проверка конца алгоритма
  ret 

__newMem:
# адрес начала выделяемой памяти в %r8
# получить адрес начала области для выделения памяти
 mov $12, %rax
 xor %rdi, %rdi
 syscall
# запомнить адрес начала выделяемой памяти
 mov %rax, %r8 
 mov %r8, (heapPointer)
# выделить динамическую память
 mov (pageSize), %rdi
 add %rax, %rdi
 mov $12, %rax
 syscall
# обработка ошибки
 cmp $-1, %rax
 jz __stop
 movq (heapSize), %rax
 mov (pageSize), %rbx
 call __sum 
 mov %rax, (heapSize) 
 mov (pageSize), %rax
 mov (heapPointer), %rbx
 call __sum 
 mov %rax, (heapMax) 
 
 mov $0, %dl
 mov $0, %rbx
 __newMemlo:
 movb %dl, (%r8, %rbx)
 inc %rbx
 cmp (pageSize), %rbx
 jz  __newMemex
 jmp __newMemlo
 
 __newMemex:
 ret
 
__sum:
# операнды в rax и в rbx
# результат в rax

 cmp $0, %rbx
 je __end_sum
 dec %rbx
 inc %rax
 jmp __sum
 __end_sum:
 ret

__sub:
# операнды в rax и в rbx
# результат в rax

 cmp $0, %rbx
 je __end_sub
 dec %rbx
 dec %rax
 jmp __sub
 __end_sub:
 ret

__defineVar:
 movq (heapMax), %rax
 cmp (heapPointer), %rax 
 jg __defOk
 call __newMem
 movq %r8, (heapPointer)
 __defOk:
 mov (heapPointer), %r8
 mov $0, %rbx
 movq %rcx, (%r8, %rbx)
 movq (heapPointer), %rax
 movq (varNameSize), %rbx
 call __sum
 movq %rax, %r8
 mov $0, %rbx 
 movq %rdx, (%r8, %rbx)
 
 movq (varSize), %rax
 movq (heapPointer), %rbx
 call __sum
 movq %rax, (heapPointer)
 ret 

__getVar:
 movq (heapBegin), %rax
 movq %rax, (getPointer)
 movq (heapMax), %rax 
 cmp (getPointer), %rax 
 jg __search
 jmp __stop # переменная не найдена, ошибка 
 __search:
 movq (getPointer), %r8
 movq $0, %rbx
 mov (%r8, %rbx), %rsi 
 call __print 
 mov $space, %rsi 
 call __print
 movq (getPointer), %rax
 movq (varNameSize), %rbx 
 call __sum 
 movq %rax, %r8
 movq $0, %rbx
 mov (%r8, %rbx), %rsi 
 call __print
 mov $enter, %rsi 
 call __print
 
 movq (getPointer), %rax
 movq (varSize), %rbx 
 call __sum 
 movq %rax, (getPointer)
 
 movq (heapMax), %rax 
 cmp (getPointer), %rax
 jg __search
 ret 

.globl _start
_start:
call __newMem
movq (heapPointer), %rax 
movq %rax, (heapBegin)

movq $var0, %rcx
movq $varT0, %rdx 
call __defineVar

movq $var1, %rcx
movq $varT1, %rdx
call __defineVar 

mov $var2, %rcx
movq $varT2, %rdx
call __defineVar 

call __getVar

movq (heapSize), %rax
call __toStr
mov $buf2, %rsi
call __print
 
__stop:
mov $60, %rax
xor %rdi, %rdi
syscall

