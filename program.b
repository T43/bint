#import "program2.b"

void main(){
	string s;
	s = "Hello";
	print((("!" + s[1:(2+1)]) + "\n"));
	print((("!" + s[0:(len(s) - 1)]) + "\n")); 
};

main();
