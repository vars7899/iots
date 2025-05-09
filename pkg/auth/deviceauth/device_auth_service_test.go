package deviceauth_test

// type mockTokenService struct {
// 	mock.Mock
// }

// func (m *mockTokenService) GenerateToken(deviceID uuid.UUID) (*deviceauth.DeviceConnectionTokens, error) {
// 	args := m.Called(deviceID)
// 	return args.Get(0).(*deviceauth.DeviceConnectionTokens), args.Error(1)
// }

// func TestIssueTokens_Success(t *testing.T) {
// 	mockRedis := new(mocks.JTIStore)
// 	mockToken := new(mockTokenService)
// 	testLogger := logger.NewTestLogger()

// 	deviceID := uuid.New()
// 	tokenData := &deviceauth.DeviceConnectionTokens{
// 		ConnectionToken:          "conn.jwt",
// 		ConnectionTokenJTI:       "conn-jti",
// 		ConnectionTokenExpiresAt: time.Now().Add(time.Hour),
// 		RefreshToken:             "ref.jwt",
// 		RefreshTokenJTI:          "ref-jti",
// 		RefreshTokenExpiresAt:    time.Now().Add(2 * time.Hour),
// 	}

// 	mockToken.On("GenerateTokens", deviceID).Return(tokenData, nil)
// 	mockRedis.On("RecordJTI", mock.Anything, "conn-jti", mock.Anything).Return(nil)
// 	mockRedis.On("RecordJTI", mock.Anything, "ref-jti", mock.Anything).Return(nil)

// 	manager := deviceauth.NewDeviceAuthManager(mockToken, mockRedis, testLogger)
// 	tokens, err := manager.IssueTokens(context.Background(), deviceID)

// 	assert.NoError(t, err)
// 	assert.Equal(t, tokenData, tokens)
// 	mockToken.AssertExpectations(t)
// 	mockRedis.AssertExpectations(t)
// }
