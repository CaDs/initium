import { APIClient } from './client';
import type { LandingInfo, TokenPair, User } from './models';

export const requestMagicLink = (client: APIClient, email: string) =>
  client.send<{ message: string }>('/api/auth/magic-link', {
    method: 'POST',
    body: { email },
    skipAuth: true,
  });

export const verifyMagicLink = (client: APIClient, token: string) =>
  client.send<TokenPair>('/api/auth/mobile/verify', {
    method: 'POST',
    body: { token },
    skipAuth: true,
  });

export const verifyGoogleIDToken = (client: APIClient, idToken: string) =>
  client.send<TokenPair>('/api/auth/mobile/google', {
    method: 'POST',
    body: { id_token: idToken },
    skipAuth: true,
  });

export const fetchMe = (client: APIClient) => client.send<User>('/api/me');

export const updateMe = (client: APIClient, update: { name?: string }) =>
  client.send<User>('/api/me', { method: 'PATCH', body: update });

export const logout = (client: APIClient) =>
  client.send<{ message: string }>('/api/auth/logout', { method: 'POST' });

export const logoutAll = (client: APIClient) =>
  client.send<{ message: string }>('/api/auth/logout-all', { method: 'POST' });

export const fetchLanding = (client: APIClient) =>
  client.send<LandingInfo>('/api/landing', { skipAuth: true });

export const adminPing = (client: APIClient) =>
  client.send<{ role: 'admin' }>('/api/admin/ping');
