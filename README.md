<p align="center">
    <img src="ui/src/icons/fox.svg" width="25%">
</p>

# Thrifty

Please be gracious with me, this is my first ever app 🙈.  Thrifty is a simple web application that helps you manage your income and expenses.
It's focused on simplicity and is not aimed to track every single penny you spend.
The idea is to get a rough overview of your monthly cash flow and what's left to spend.

Features:
- Add income and expenses
- Edit existing entries
- Delete entries
- Support for SVG icons (default: <img width=19 align=center alt="dollar" src="doc/default-icon.svg"/>)
- Two rows for income and expenses (collapse into single one on smaller devices)
- API documentation at `/swagger/index.html`

Frameworks used:

- __UI__: SvelteKit and Flowbite
- __API__: golang

## Screenshots
![Screenshot1](doc/screenshot_1.png)
![Screenshot2](doc/screenshot_2.png)

## Running the app

Every release since 1.2.0 should create a docker image on DockerHub.
Use [docker-compose.yaml](docker-compose.yaml) to run the app with the latest version.

```bash
# run the app
docker compose -f docker-compose.yaml up -d
```
The app should now be available at `http://localhost:9090`.

If you want to build the image yourself, you can use the following commands.
```bash
# run the app
docker compose -f docker-compose-build.yaml build

# run the app
docker compose -f docker-compose-build.yaml up -d
```

### Tips
#### Want to use `$` as the currency symbol?
>Change `currency: 'EUR'` to `currency: 'USD'` in [+page.svelte](ui/src/routes/+page.svelte), delete old docker stack, build containers (docker-compose-build.yaml) and start the app like mentioned above.
#### How about other currencies?
> The step above should apply for any other currency ISO-code in this list: [ISO 4217](https://de.wikipedia.org/wiki/ISO_4217)
#### Where is the data stored?
> The data is stored in a SQLite database in a docker volume. You can also use a custom path in the [docker-compose.yaml](docker-compose.yaml) to expose the database.
#### I need a docker image for a different architecture.
> You can build the image yourself like described above.

## Developing

### Frontend
Install node and node modules.
Running locally requires you to change the API URL in [+page.svelte](ui/src/routes/+page.svelte).
Prepend the URL paths in the fetch calls with `http://localhost:8080/`.
```bash
cd ui

# install dependencies
pnpm i

# run ui
pnpm dev
```
You could also use regular npm instead of pnpm.
### Backend
Install golang and run the following commands.
```bash
cd api

# install dependencies
go get .

# install swag (optional)
go install github.com/swaggo/swag/cmd/swag@latest; \

# generate swagger documentation (optional)
swag init

# run api
go run .
```

## Credits
- [NumberFlow](https://number-flow.barvian.me/svelte)
- [Swag](https://github.com/swaggo/swag)