package crypto

import "github.com/stretchr/testify/mock"

type generatorMock struct {
	mock.Mock
}

func (self *generatorMock) SecureId(length int) (string, error) {
	args := self.Called(length)

	return args.String(0), args.Error(1)
}

func (self *generatorMock) RandomString(length int) string {
	args := self.Called(length)

	return args.String(0)
}

type passwordManagerMock struct {
	mock.Mock
}

func (self *passwordManagerMock) Hash(password string) (string, error) {
	args := self.Called(password)

	return args.String(0), args.Error(1)
}

func (self *passwordManagerMock) Compare(password, hash string) (bool, error) {
	args := self.Called(password, hash)

	return args.Bool(0), args.Error(1)
}

type cryptoMock struct {
	mock.Mock
	GeneratorMock       generatorMock
	PasswordManagerMock passwordManagerMock
}

func CreateMock() cryptoMock {
	return cryptoMock{}
}

func (self *cryptoMock) Generator() generator {
	return &self.GeneratorMock
}

func (self *cryptoMock) PasswordManager() passwordManager {
	return &self.PasswordManagerMock
}
