# gojira

Manage releases in a less frustrating way

## Building
```
go build -o gojira
```

## Usage

```
# WIP Give information about pending release
./gojira release check --releaseplan windows-machine-config-operator-10-17-staging --version v10.17.1 --project WINC


# create JIRA issues to track a new minor release
./gojira release new --date 2024-05-13 --version 8.1.3 --major false

# List pending releases in the given JIRA project
./gojira release list --project WINC
```
