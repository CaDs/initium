import {
  AdminPingResponseSchema,
  LandingInfoSchema,
  MessageResponseSchema,
  TokenPairSchema,
  type UpdateProfileInput,
  UserSchema,
} from '@initium/core';

import { APIClient } from './client';

export const requestMagicLink = (client: APIClient, email: string) =>
  client.send('/api/auth/magic-link', MessageResponseSchema, {
    method: 'POST',
    body: { email },
    skipAuth: true,
  });

export const verifyMagicLink = (client: APIClient, token: string) =>
  client.send('/api/auth/mobile/verify', TokenPairSchema, {
    method: 'POST',
    body: { token },
    skipAuth: true,
  });

export const verifyGoogleIDToken = (client: APIClient, idToken: string) =>
  client.send('/api/auth/mobile/google', TokenPairSchema, {
    method: 'POST',
    body: { id_token: idToken },
    skipAuth: true,
  });

export const fetchMe = (client: APIClient) => client.send('/api/me', UserSchema);

export const updateMe = (client: APIClient, update: UpdateProfileInput) =>
  client.send('/api/me', UserSchema, { method: 'PATCH', body: update });

export const logout = (client: APIClient) =>
  client.send('/api/auth/logout', MessageResponseSchema, { method: 'POST' });

export const logoutAll = (client: APIClient) =>
  client.send('/api/auth/logout-all', MessageResponseSchema, { method: 'POST' });

export const fetchLanding = (client: APIClient) =>
  client.send('/api/landing', LandingInfoSchema, { skipAuth: true });

export const adminPing = (client: APIClient) =>
  client.send('/api/admin/ping', AdminPingResponseSchema);
