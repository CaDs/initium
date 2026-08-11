import { Schema } from "effect";

const EmailString = Schema.String.pipe(
  Schema.pattern(/^[^\s@]+@[^\s@]+\.[^\s@]+$/, {
    message: () => "Expected a valid email address",
  }),
);

export const APIErrorSchema = Schema.Struct({
  code: Schema.String,
  message: Schema.String,
  request_id: Schema.optional(Schema.String),
});

export const AdminPingResponseSchema = Schema.Struct({
  role: Schema.Literal("admin"),
});

export const LandingInfoSchema = Schema.Struct({
  name: Schema.String,
  description: Schema.String,
  version: Schema.String,
});

export const MagicLinkInputSchema = Schema.Struct({
  email: EmailString,
});

export const MessageResponseSchema = Schema.Struct({
  message: Schema.String,
});

export const MobileGoogleInputSchema = Schema.Struct({
  id_token: Schema.NonEmptyString,
});

export const MobileVerifyInputSchema = Schema.Struct({
  token: Schema.NonEmptyString,
});

export const RefreshInputSchema = Schema.Struct({
  refresh_token: Schema.optional(Schema.String),
});

export const TokenPairSchema = Schema.Struct({
  access_token: Schema.String,
  refresh_token: Schema.String,
});

export const UpdateProfileInputSchema = Schema.Struct({
  name: Schema.String.pipe(Schema.minLength(1), Schema.maxLength(100)),
});

export const UserSchema = Schema.Struct({
  id: Schema.String,
  email: EmailString,
  name: Schema.String,
  avatar_url: Schema.String,
  role: Schema.Literal("user", "admin"),
  created_at: Schema.String,
});

export type APIError = typeof APIErrorSchema.Type;
export type AdminPingResponse = typeof AdminPingResponseSchema.Type;
export type LandingInfo = typeof LandingInfoSchema.Type;
export type MagicLinkInput = typeof MagicLinkInputSchema.Type;
export type MessageResponse = typeof MessageResponseSchema.Type;
export type MobileGoogleInput = typeof MobileGoogleInputSchema.Type;
export type MobileVerifyInput = typeof MobileVerifyInputSchema.Type;
export type RefreshInput = typeof RefreshInputSchema.Type;
export type TokenPair = typeof TokenPairSchema.Type;
export type UpdateProfileInput = typeof UpdateProfileInputSchema.Type;
export type User = typeof UserSchema.Type;
