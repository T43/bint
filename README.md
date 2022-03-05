#Интерпретатор языков B и Basm 
##Синтаксис языка Basm
Язык Basm - C-подобный язык

1. Каждое команда заканчивается символом ";"
2. Типы данных - float, int, string, stack
3. Возможно явное приведение типов данных функциями float(), int(), str.
###Пример явного преобрзования типов 
    string buf;
    int a;
    a = int(buf);
4. Условная дизъюнкция имеет следующий синтаксис
###Синтаксис условной дизъюнкции
    [goto(#mark1)/print("text1"), (<условие>), goto(#mark2)/print("text2")];
В случае, когда условие истинно, выполняется команда слева от условия (goto(#mark1), либо print("text1")).
В противном случае выполняется команда справа

В условии обязательны скобки для каждой элементарной операции,
а также для каждого операнда перед логической операцией (см. ниже)

5. Логические операторы: AND, OR, XOR, NOT
###Пример логического выражения
    bool b
    b = ((True)AND(False));

6. Метка объявляется следующим образом: 
###Пример метки 
    #mark1:
7. Оператор goto позволяет перейти к метке, объявленной в любом месте программы:
###Пример использования оператора goto
    goto(#mark);
8. Приведем пример условной дизъюнкции 
###Условная дизъюнкция 
    [goto(#left), (4 > 5), goto(#right];
В данном случае интерпретатор перейдет к метке #right 
9. Как было сказано выше, скобки обязательны для каждой элементарной операции.
Приведем пример корректного арифметического выражения 
###Корректное арифметическое выражение 
    float res; 
    res = (((5^2) + 4) - 3);
10. Арифметические операции: +, -, *, /, ^
11. Оператор print печатает выражение на экран 
###Пример использования print 
    print("Hello world!\n");
Примечание: русский язык не поддерживается 
12. Оператор len позволяет вычислить длину строки 
###Пример использования len 
    int a; 
    string buf; 
    buf = "Hello world!";
    a = len(buf);
Примечание: использование len возможно только после переменной и знака присваивания
13. Оператор index позволяет определить индекс первого вхождения подстроки в строку 
###Синтаксис index 
    index(<строка>, <подстрока>);
###Пример использования index 
    int a;
    a = index("banana", "nan");
В случае отсуствия вхождения index возвращает -1. 
Как и для оператора len, использование index возможно только после переменной и знака присваивания
14. Оператор push позволяет положить переменную любого типа в стек
###Пример использования push 
    push(100);
15. Оператор pop позволяет достать содержимое с вершины стека. 
При этом переменная соответствующего типа передается оператору pop в качестве аргумента 
###Пример использования pop для того, чтобы достать с вершины стека число 100
    int a; 
    pop(a); 
Теперь в переменной a находится число 100 
16. Специальный тип stack позволяет определять стек пользовательского типа.
Чтобы применять к переменной типа stack операции push и pop необходимо после
переменной поставить символ "."
###Пример использования пользовательского стека 
    stack my_stack; 
    my_stack.push(5);
    my_stack.push("Hello world!"); 
    string buf; 
    my_stack.pop(buf);
    print(buf);
    print("\n");
    int a; 
    my_stack.pop(a); 
    buf = str(a); 
    print(buf);
    print("\n");
17. Операция SET_SOURCE("<файл>") открывает системный файл на чтение 
###Пример использования SET_SOURCE
    SET_SOURCE("program.b");
18. Операция SET_DEST("<файл>") открывает системный файл на запись
19. Операция UNSET_SOURCE() закрывает системный файл, открытый на чтение 
20. Операция UNSET_DEST() закрывает системный файл, открытый на запись 
Примечание: единовременно может быть открыт только один системный файл на запись и
один системный файл на чтение 
21. Операция next_command(<переменная типа string>) позволяет считать очередную 
команду из системного файла до символа ";"
22. Операция send_command(<переменная типа string>) позволяет послать очередную команду
на запись 
###Пример считывания и пересылки команды 
    SET_SOURCE("program.b");
    SET_DEST("program.basm");
    string command; 
    next_command(command);
    send_command(command);
    UNSET_SOURCE();
    USET_DEST();
23. Оператор UNDEFINE(<переменная>) сообщает интерпретатору о том, что
необходимо "забыть" о существовании переменной 
24. Логические значения: True и False 
25. Операции сравнения: <, <=, ==, >=, >
Примечание: операция "!=" отсутствует. Вместо этого стоит использовать
оператор NOT 
###Пример условия неравенства
    int a; 
    int b; 
    a = 5; 
    b = 10;
    [goto(#left), (NOT(a == b)), goto(#right)];
26. Для ввода данных используется оператор input
###Пример использования input 
    string buf; 
    input(buf);
27. Пробелы, символы табуляции и перехода на следующую строку
игнорируются 
28. Комментарии запрещены
29. Приведем пример программы на языке Basm
###Пример программы решения квадратных уравнений на языке Basm
    print("Solving equation of the form ax^2 + bx + c = 0\n");
    #begin:
    print("Input a\n");
    float a;
    string buf;
    input(buf);
    a = float(buf);
    print("Input b\n");
    float b;
    input(buf);
    b = float(buf);
    print("Input c\n");
    float c;
    input(buf);
    c = float(buf);
    float d;
    d = ((b^2) - ((4*a)*c));
    float x1;
    float x2;
    [print(""), (d >= 0), goto(#no_solution)];
    x1 = ((((-1)*b) - (d^0.5)) / (2*a));
    x2 = ((((-1)*b) + (d^0.5)) / (2*a));
    print("x1 = ");
    buf = str(x1);
    print(buf);
    print("\n");
    print("x2 = ");
    buf = str(x2);
    print(buf);
    print("\n");
    goto(#end_iter);
    #no_solution:
    print("No solution\n");
    #end_iter:
    print("Exit? y/n\n");
    input(buf);
    [print(""), ("y" == buf), goto(#begin)];
    
## Синтаксис языка B 
1. B является языком над языком Basm. Язык B транслируется в язык Basm, а 
затем полученная программа интерпретируется. Чтобы запустить трансляцию,
необходимо в main.go взвести флаг toTranslate в true. Тогда программа program.b,
находящаяся в корне, будет траслинрована в программу program.basm, записываемую также в корень
с помощью препроцессора, находящегося в папке benv с именем func.basm. Если взвести флаг
toTranslate в false, то будет интерпретирована программа program.basm.
2. В настоящее время язык B находится в разработке 

##Примечания 
1. С целью отладки модуля синтаксического анализа parser.go, был добавлен  модуль
отрисовки абстактных синтаксических деревьев (АСД) drawModule.go. Чтобы воспользоваться
модулем отрисовки, нужно передать в функцию parser.Parse последний параметр (showTree) true
