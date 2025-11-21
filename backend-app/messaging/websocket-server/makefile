tidy:
	rm -rf node_modules package-lock.json
	npm install

run:
	set -a; source src/.env; set +a; node dist/main.js

build:
	npx tsc


buildrun:
	make build & make run
