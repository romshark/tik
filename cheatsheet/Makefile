all: gomod templ gendist

gomod:
	go mod tidy

gendist:
	go run .

templ:
	go run github.com/a-h/templ/cmd/templ@v0.3.898 generate