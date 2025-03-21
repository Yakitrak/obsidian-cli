package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockUriManager struct {
	mock.Mock
}

func (m *MockUriManager) Construct(base string, params map[string]string) string {
	args := m.Called(base, params)
	return args.String(0)
}

func (m *MockUriManager) Execute(uri string) error {
	args := m.Called(uri)
	return args.Error(0)
}
