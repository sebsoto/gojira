# gojira

Make release management less frustrating

## Building
```
go build -o gojira
```

## Usage

```
# create JIRA issues to track a new minor release
./gojira release new --date 2024-05-13 --version 8.1.3 --major false

# List pending releases in the given JIRA project
./gojira release list --project WINC
```
