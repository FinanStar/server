package user

import "context"

type expectTuple struct {
	User  *userEntity
	Error error
}

type testUserRepository struct {
	createExpect     *expectTuple
	getByLoginExpect *expectTuple
	updateExpect     *expectTuple
}

func NewTestUserRepository() testUserRepository {
	return testUserRepository{}
}

func (self *testUserRepository) Create(
	ctx context.Context,
	dto createUserRepositoryDto,
) (*userEntity, error) {
	if self.createExpect != nil {
		return self.createExpect.User, self.createExpect.Error
	}

	return nil, nil
}

func (self *testUserRepository) CreateExpectResult(
	user *userEntity,
	err error,
) {
	self.createExpect = &expectTuple{
		User:  user,
		Error: err,
	}
}

func (self *testUserRepository) Update(
	ctx context.Context,
	id uint32,
	dto updateUserRepositoryDto,
) (*userEntity, error) {
	if self.updateExpect != nil {
		return self.updateExpect.User, self.updateExpect.Error
	}

	return nil, nil
}

func (self *testUserRepository) UpdateExpectResult(
	user *userEntity,
	err error,
) {
	self.updateExpect = &expectTuple{
		User:  user,
		Error: err,
	}
}

func (self *testUserRepository) GetByLogin(
	ctx context.Context,
	login string,
) (*userEntity, error) {
	if self.getByLoginExpect != nil {
		return self.getByLoginExpect.User, self.getByLoginExpect.Error
	}

	return nil, nil
}

func (self *testUserRepository) GetByLoginExpectResult(
	user *userEntity,
	err error,
) {
	self.getByLoginExpect = &expectTuple{
		User:  user,
		Error: err,
	}
}
