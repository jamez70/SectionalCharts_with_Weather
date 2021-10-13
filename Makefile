TARGETS := getwx cgimap cgipart 
SRCS := getwx.go cgimap.go cgipart.go
INSTALL_TARGET := /var/www/html/map

all: $(TARGETS)

cgipart:	cgipart.go
	go build cgipart.go

getwx:	getwx.go
	go build getwx.go

cgimap: cgimap.go
	go build cgimap.go

clean:
	rm -f $(TARGETS) 

fmt:
	go fmt $(SRCS)

run: $(TARGET)
	./$(TARGET)

.PHONY: run fmt clean
