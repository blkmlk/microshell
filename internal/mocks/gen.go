package mocks

import "github.com/stretchr/testify/mock"

var AnyArgument = mock.MatchedBy(func(arg interface{}) bool {
	return true
})

// nolint:lll
//go:generate mockery -name=TestExec -case=snake -dir=../parser/ -inpkg -output=../parser/ -note=mock

//go:generate mockery -name=Expression -case=snake -dir=../parser/ -inpkg -output=../parser/ -note=mock
