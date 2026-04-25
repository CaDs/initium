package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/eridia/initium/backend/internal/domain"
	"github.com/eridia/initium/backend/internal/service"
	"github.com/eridia/initium/backend/internal/testutil"
)

// UserService is a thin pass-through to UserRepository. The value of
// these tests is asserting the wrapping behavior: error wrapping on
// repo failures, name-update mutation order, and that the returned
// user reflects the new name (not a stale read).

func TestUserService_GetProfile_Success_ReturnsUser(t *testing.T) {
	t.Parallel()

	repo := &testutil.MockUserRepository{
		FindByIDFn: func(_ context.Context, id string) (*domain.User, error) {
			require.Equal(t, testutil.RegularUser.ID, id)
			u := testutil.RegularUser
			return &u, nil
		},
	}
	svc := service.NewUserService(repo)

	user, err := svc.GetProfile(context.Background(), testutil.RegularUser.ID)

	require.NoError(t, err)
	assert.Equal(t, testutil.RegularUser.Email, user.Email)
}

func TestUserService_GetProfile_RepoError_WrapsError(t *testing.T) {
	t.Parallel()

	repoErr := errors.New("connection refused")
	repo := &testutil.MockUserRepository{
		FindByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, repoErr
		},
	}
	svc := service.NewUserService(repo)

	user, err := svc.GetProfile(context.Background(), "u")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, repoErr, "service must wrap repo errors with errors.Is preserved")
}

func TestUserService_UpdateProfile_Success_AppliesNewName(t *testing.T) {
	t.Parallel()

	stored := testutil.RegularUser
	stored.Name = "Old Name"

	repo := &testutil.MockUserRepository{
		FindByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			u := stored
			return &u, nil
		},
		UpdateFn: func(_ context.Context, u *domain.User) error {
			require.Equal(t, "New Name", u.Name,
				"Update must receive the user with the new name applied")
			stored = *u
			return nil
		},
	}
	svc := service.NewUserService(repo)

	user, err := svc.UpdateProfile(context.Background(), testutil.RegularUser.ID, "New Name")

	require.NoError(t, err)
	assert.Equal(t, "New Name", user.Name,
		"returned user must reflect the new name, not a stale read")
}

func TestUserService_UpdateProfile_FindError_WrapsError(t *testing.T) {
	t.Parallel()

	repo := &testutil.MockUserRepository{
		FindByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	svc := service.NewUserService(repo)

	user, err := svc.UpdateProfile(context.Background(), "u", "name")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserService_UpdateProfile_RepoUpdateError_WrapsError(t *testing.T) {
	t.Parallel()

	updateErr := errors.New("update failed")
	repo := &testutil.MockUserRepository{
		FindByIDFn: func(_ context.Context, _ string) (*domain.User, error) {
			u := testutil.RegularUser
			return &u, nil
		},
		UpdateFn: func(_ context.Context, _ *domain.User) error {
			return updateErr
		},
	}
	svc := service.NewUserService(repo)

	user, err := svc.UpdateProfile(context.Background(), "u", "name")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, updateErr)
}
