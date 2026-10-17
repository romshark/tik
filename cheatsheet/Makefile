TEMPL = go run github.com/a-h/templ/cmd/templ@v0.3.1070

all: gomod templ gendist

gomod:
	go mod tidy

gendist:
	go run .

templ:
	$(TEMPL) generate

dev:
	$(TEMPL) generate --watch --proxy=http://127.0.0.1:8080 --cmd="go run . -serve 127.0.0.1:8080"
