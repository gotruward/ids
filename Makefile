
.PHONY: all
all: gen/base32_table.go

gen/base32_table.go: gen tools
	go run ./tools/base32_table_gen.go > gen/base32_table.go

gen:
	mkdir gen

.PHONY: clean
clean:
	rm -rf gen
