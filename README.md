# Компилятор языков B и Basm с защитой от несанкционированного исполнения

## Синтаксис языков B и Basm
**B** является **C**-подобным языком над языком **Basm**. Язык **B** транслируется в язык **Basm**, а 
затем полученная программа может быть исполнена. Подробное описание синтаксиса языков может быть найдено в 
**docs/Basm.odt** и **docs/B.odt**. Примеры программ на языке **B** могут быть найдены в папке **examples**

## Разложение на элементарные операции и защита от несанкционированного исполнения
Разложение программы на языке **Basm** на элементарные операции заключается в представлении
исходной программы на языке (**Bend**), который может быть исполнен минимальным ядром компилятора 
Защита заключается в добавлении мусора (реально существующих случайных команд) внутрь конструкций 
языка **Bend** и запоминании смещений, по которым лежат реальные команды, в отдельный файл с раширением **k**.
Полученный защищенный файл имеет расширение **benc** и может быть исполнен с файлом-ключом **k**
## Начать работу 
### Трансляция программы с языка B на язык Basm 
```
./bint -i input.b -o output.basm
```
### Исполнить программу на языке Basm 
```
./bint -e output.basm 
```
### Разложить программу на языке Basm на элементарные операции языка Bend 
```
./bint -pi input.basm -po output.bend 
```
### Исполнить программу на языке Bend 
```
./bint -pe output.bend 
```
### Защитить программу на языке Bend, сгенерировав файл с ключом key.k
```
./bint -ci input.bend -co output.benc -k key.k
```
### Выполнить защищенную программу output.benc с ключом key.k
```
./bint -ce output.benc -k key.k 
```
## Примечания
Рекомендуется использовать последний релиз, распоковав его по пути **/usr/local/bint**. На папку 
bint необхоидимо выдать соответствующие права 
```
sudo chmod -R 0777 bint 
```
После рекомендуется выполнить следующую команду 
```
sudo nano ~/.bashrc
```
И прописываем путь к bint в переменную окружения PATH в конце файла 
```
export PATH=$PATH:/usr/local/bint
```
После этого 
```
source ~/.bashrc
```
Компилятор готов к использованию. Введите в терминал
```
bint 
```
Файл **b.lang** описывает синтаксис языков **B** и **Basm** для текстового редактора gedit (протестировано
на Ubuntu 18.04).
Чтобы gedit распознал языки **B** и **Basm**, необходимо добавить данный файл по следующему пути:
```
usr/share/gtksourceview-3.0/language-specs
```
