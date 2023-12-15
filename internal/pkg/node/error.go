package node

import "errors"

var ErrNotFound = errors.New("node/profile not found")
var ErrNoUnconfigured = errors.New("no unconfigured node")
