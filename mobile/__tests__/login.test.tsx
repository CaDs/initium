import { render } from '@testing-library/react-native';
import React from 'react';

import LoginScreen from '../app/login';
import { useAuthStore } from '@/auth/store';

jest.mock('expo-router', () => ({
  Redirect: () => null,
  Link: ({ children }: { children: React.ReactNode }) => children,
}));

jest.mock('expo-web-browser', () => ({
  maybeCompleteAuthSession: jest.fn(),
}));

const mockUseAuthRequest = jest.fn((..._args: unknown[]) => [null, null, jest.fn()]);
jest.mock('expo-auth-session/providers/google', () => ({
  useAuthRequest: (config: unknown) => mockUseAuthRequest(config),
}));

let mockGoogleConfigured = false;
jest.mock('@/config', () => ({
  config: {
    apiBaseURL: 'http://localhost:8000',
    devBypassAuth: false,
    google: {
      iosClientId: '',
      androidClientId: '',
      webClientId: '',
    },
  },
  googleConfigured: () => mockGoogleConfigured,
}));

jest.mock('react-native-safe-area-context', () => {
  const actualReact = jest.requireActual('react');
  const { View } = jest.requireActual('react-native');
  return {
    SafeAreaView: ({ children, ...rest }: { children: React.ReactNode }) =>
      actualReact.createElement(View, rest, children),
    SafeAreaProvider: ({ children }: { children: React.ReactNode }) => children,
  };
});

beforeEach(() => {
  mockUseAuthRequest.mockClear();
  mockGoogleConfigured = false;
  useAuthStore.setState({ status: 'unauthenticated', user: null, error: null });
});

describe('LoginScreen', () => {
  it('renders without invoking Google.useAuthRequest when no client IDs are configured', () => {
    const { getByText } = render(<LoginScreen />);

    expect(getByText('Initium')).toBeTruthy();
    expect(getByText('Email magic link')).toBeTruthy();
    expect(getByText(/Configure GOOGLE_\*_CLIENT_ID/)).toBeTruthy();
    expect(mockUseAuthRequest).not.toHaveBeenCalled();
  });

  it('mounts the Google button (and runs the hook) when configured', () => {
    mockGoogleConfigured = true;
    const { getByText } = render(<LoginScreen />);

    expect(getByText('Continue with Google')).toBeTruthy();
    expect(mockUseAuthRequest).toHaveBeenCalledTimes(1);
  });
});
