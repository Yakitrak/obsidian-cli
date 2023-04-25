package mocks

type MockVaultOperator struct {
	ExecuteErr error
	PathError  error
	Name       string
}

func (m *MockVaultOperator) DefaultName() (string, error) {
	if m.ExecuteErr != nil {
		return "", m.ExecuteErr
	}
	return m.Name, nil
}

func (m *MockVaultOperator) SetDefaultName(_ string) error {
	if m.ExecuteErr != nil {
		return m.ExecuteErr
	}
	return nil
}

func (m *MockVaultOperator) Path() (string, error) {
	if m.PathError != nil {
		return "", m.PathError
	}
	return "path", nil
}
