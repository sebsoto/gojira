# gojira

A tool to help manage OpenShift operator releases through konflux.

## Warning

This tool is under active development and the interface may change as features are added.

## Building
```
go build -o gojira
```

## Requirements

* The konflux application being built should have a stage ReleasePlan with automatic releases, as well as a production
  ReleasePlan with which to use this tool. This ensures that snapshots used to generate a new release are able to pass
  any integration tests defined for the stage release.
* A Jira personal access token must be provisioned and saved to ~/.jira/token
* `git tag` should be used to tag the commits associated with a konflux release. This tag should have the semver format vX.Y.Z.
* [Recommended] A Github personal access token saved to ~/.github/token. Without this token rate limiting may occur.

## Usage

```
# Output a konflux release object for the release as well as information of all stories included in the release
$ ./gojira release status --releaseplan windows-machine-config-operator-10-19-prod --project WINC --version v10.19.0 --namespace windows-machine-conf-tenant
```

