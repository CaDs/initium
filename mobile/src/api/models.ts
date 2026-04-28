export type TokenPair = {
  access_token: string;
  refresh_token: string;
};

export type User = {
  id: string;
  email: string;
  name: string;
  role: 'user' | 'admin';
  avatar_url?: string | null;
};

export type ErrorEnvelope = {
  code: string;
  message: string;
  request_id?: string;
};

export type LandingInfo = {
  name: string;
  description: string;
  version: string;
};
