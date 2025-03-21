package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockVaultOperator struct {
	DefaultNameErr error
	PathError      error
	Name           string
}

func (m *MockVaultOperator) DefaultName() (string, error) {
	if m.DefaultNameErr != nil {
		return "", m.DefaultNameErr
	}
	return m.Name, nil
}

func (m *MockVaultOperator) SetDefaultName(_ string) error {
	if m.DefaultNameErr != nil {
		return m.DefaultNameErr
	}
	return nil
}

func (m *MockVaultOperator) Path() (string, error) {
	if m.PathError != nil {
		return "", m.PathError
	}
	return "path", nil
}

type VaultManager struct {
	mock.Mock
}

func (m *VaultManager) DefaultName() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *VaultManager) SetDefaultName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *VaultManager) Path() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}
