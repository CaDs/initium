import { z } from "zod";

export const userSchema = z.object({
  id: z.string(),
  email: z.email(),
  name: z.string(),
  avatar_url: z.string(),
  role: z.enum(["user", "admin"]),
  created_at: z.string(),
});

export const tokenPairSchema = z.object({
  access_token: z.string(),
  refresh_token: z.string(),
});

export const errorSchema = z.object({
  code: z.string(),
  message: z.string(),
  request_id: z.string().optional(),
});

export const messageSchema = z.object({
  message: z.string(),
});
