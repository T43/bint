#import "stdlib/core.b"

string root_source;
int num;
bool first_file;

void init(){
	num = 0;
	first_file = True; 
	get_root_source(root_source);
	SET_SOURCE(root_source);
	SET_DEST("benv/for_program.b");
};

void finish(){
	UNSET_SOURCE();
	UNSET_DEST();
};

bool is_for(string command){
	stack s;
	string buf;
	int pos;

	buf = "for(";
	s = ops(command, buf);
	s.pop(buf);

	if (NOT("end" == buf)){
		pos = int(buf);
		if (NOT(0 == pos)){
			println("for: ERROR");
			exit(1);	
		};	
	};

	return (NOT("end" == buf));
};

void switch_files(){
	finish();
	[print(""), (first_file), goto(#first_end)];
	SET_SOURCE("benv/for_program.b");
	SET_DEST("benv/for_program2.b");
	first_file = False;
	goto(#switch_files_e);
	#first_end:
	SET_SOURCE("benv/for_program2.b");
	SET_DEST("benv/for_program.b");
	first_file = True; 
	#switch_files_e:
	print("");
};

void clear_files(){
	[goto(#first_file), (first_file), print("")];
	switch_files();
	switch_command();
	#clear_files_s:
	[goto(#clear_files_e), ("end" == command), print("")];
	send_command(command);
	switch_command();
	goto(#clear_files_s);

	#first_file:
	print("");
	#clear_files_e:
	finish();
	DEL_DEST("benv/for_program2.b");
};

void main(){
	init();
	int counter;
	int command_len;
	string snum;
	string buf;
	string inc;
	int pos;
	
	#next:
	switch_command();
	if (NOT("end" == command)){
		if (is_for(command)){
			command_len = len(command);
			command = command[4:command_len];
			send_command(command);
			switch_command();
			send_command(command);
			switch_command();
			snum = str(num);
			buf = (("#for" + snum) + ":print(\"\")");
			send_command(buf);  
			buf = (("if(" + command) + "){print(\"\")");
			send_command(buf);
			switch_command();
			counter = block_end();
			pos = index(command, "{");
			if (-1 == pos){
				println("for: ERROR");
				exit(1);			
			};
			pos = (pos - 1);
			inc = command[0:pos];
			#next_internal:
			switch_command(); 
			if (COMMAND_COUNTER < counter){
				send_command(command);
				goto(#next_internal);			
			};
			send_command(inc);
			buf = (("goto(#for" + snum) + ")");
			send_command(buf); 
			num = (num + 1);
		}else{
			send_command(command);
		};
		goto(#next);	
	};
	finish();
};

main();
