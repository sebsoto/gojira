package semver

import (
	"fmt"
	"strconv"
	"strings"
)

type Semver struct {
	Major int
	Minor int
	Patch int
}

func New(semver string) (*Semver, error) {
	semverSplit := strings.Split(semver, ".")
	if len(semverSplit) != 3 {
		return nil, fmt.Errorf("expected a semver of format vX.Y.Z")
	}
	major, err := strconv.Atoi(strings.TrimPrefix(semverSplit[0], "v"))
	if err != nil {
		return nil, err
	}
	minor, err := strconv.Atoi(semverSplit[1])
	if err != nil {
		return nil, err
	}
	patch, err := strconv.Atoi(semverSplit[2])
	if err != nil {
		return nil, err
	}
	return &Semver{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}
