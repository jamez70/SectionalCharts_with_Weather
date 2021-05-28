TARGETS := getwx cgimap
SRCS := getwx.go cgimap.go
INSTALL_TARGET := /var/www/html/map

all: $(TARGETS)

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
