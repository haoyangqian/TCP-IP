GOPATH = ${PWD}
export GOPATH

all:		
	go install  ./src/... 
clean:	
	rm -f bin/tcp_node
